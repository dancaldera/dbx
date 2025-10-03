package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/dancaldera/dbx/internal/models"
)

// GetTables retrieves all tables from the database
func GetTables(db *sql.DB, driver string) ([]string, error) {
	var query string
	switch driver {
	case "postgres":
		query = "SELECT tablename FROM pg_tables WHERE schemaname = 'public'"
	case "mysql":
		query = "SHOW TABLES"
	case "sqlite3":
		query = "SELECT name FROM sqlite_master WHERE type='table'"
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// GetSchemas retrieves schema information for PostgreSQL
func GetSchemas(db *sql.DB, driver string) ([]models.SchemaInfo, error) {
	var schemas []models.SchemaInfo

	switch driver {
	case "postgres":
		query := `
			SELECT
				schema_name,
				CASE
					WHEN schema_name = 'public' THEN 'Default public schema'
					WHEN schema_name IN ('information_schema', 'pg_catalog', 'pg_toast') THEN 'System schema'
					ELSE 'User schema'
				END as description
			FROM information_schema.schemata
			WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
			ORDER BY
				CASE WHEN schema_name = 'public' THEN 0 ELSE 1 END,
				schema_name`

		rows, err := db.Query(query)
		if err != nil {
			// If schema query fails, return just the public schema
			return []models.SchemaInfo{{Name: "public", Description: "Default public schema"}}, nil
		}
		defer rows.Close()

		for rows.Next() {
			var schema models.SchemaInfo
			err := rows.Scan(&schema.Name, &schema.Description)
			if err != nil {
				continue
			}
			schemas = append(schemas, schema)
		}

		// If no schemas found, add public as fallback
		if len(schemas) == 0 {
			schemas = append(schemas, models.SchemaInfo{Name: "public", Description: "Default public schema"})
		}

	case "mysql", "sqlite3":
		// MySQL and SQLite don't have schemas in the same way PostgreSQL does
		return []models.SchemaInfo{}, nil
	}

	return schemas, nil
}

// GetTableInfos retrieves detailed table information including row counts
func GetTableInfos(db *sql.DB, driver, schema string) ([]models.TableInfo, error) {
	var tableInfos []models.TableInfo

	switch driver {
	case "postgres":
		query := `
			SELECT
				table_name,
				table_schema as schema_name,
				table_type,
				CASE
					WHEN table_type = 'BASE TABLE' THEN COALESCE(s.n_tup_ins + s.n_tup_upd - s.n_tup_del, 0)
					ELSE 0
				END as estimated_rows
			FROM information_schema.tables t
			LEFT JOIN pg_stat_user_tables s ON t.table_name = s.relname AND t.table_schema = s.schemaname
			WHERE t.table_schema = $1
				AND t.table_type IN ('BASE TABLE', 'VIEW')
			ORDER BY t.table_type, t.table_name`

		rows, err := db.Query(query, schema)
		if err != nil {
			// Fallback to simple table list if stats are not available
			return GetSimpleTableInfos(db, driver, schema)
		}
		defer rows.Close()

		for rows.Next() {
			var info models.TableInfo
			var estimatedRows sql.NullInt64
			err := rows.Scan(&info.Name, &info.Schema, &info.TableType, &estimatedRows)
			if err != nil {
				continue
			}

			// Determine the object type display name
			var objectType string
			var emoji string
			if info.TableType == "VIEW" {
				objectType = "view"
				emoji = "👁️"
			} else {
				objectType = "table"
				emoji = ""
			}

			if info.TableType == "BASE TABLE" && estimatedRows.Valid && estimatedRows.Int64 > 0 {
				info.RowCount = estimatedRows.Int64
				info.RowCount = max(info.RowCount, 0)
				if info.Schema != "" && info.Schema != "public" {
					info.Description = fmt.Sprintf("%s %s.%s • ~%d rows", emoji, info.Schema, objectType, info.RowCount)
				} else {
					info.Description = fmt.Sprintf("%s %s • ~%d rows", emoji, strings.Title(objectType), info.RowCount)
				}
			} else {
				if info.Schema != "" && info.Schema != "public" {
					info.Description = fmt.Sprintf("%s %s.%s", emoji, info.Schema, objectType)
				} else {
					info.Description = fmt.Sprintf("%s %s", emoji, strings.Title(objectType))
				}
			}

			tableInfos = append(tableInfos, info)
		}

	case "mysql":
		query := `
			SELECT
				TABLE_NAME,
				TABLE_SCHEMA,
				TABLE_TYPE,
				COALESCE(TABLE_ROWS, 0) as table_rows
			FROM INFORMATION_SCHEMA.TABLES
			WHERE TABLE_SCHEMA = DATABASE()
				AND TABLE_TYPE IN ('BASE TABLE', 'VIEW')
			ORDER BY TABLE_TYPE, TABLE_NAME`

		rows, err := db.Query(query)
		if err != nil {
			return GetSimpleTableInfos(db, driver, schema)
		}
		defer rows.Close()

		for rows.Next() {
			var info models.TableInfo
			var tableRows sql.NullInt64
			err := rows.Scan(&info.Name, &info.Schema, &info.TableType, &tableRows)
			if err != nil {
				continue
			}

			// Determine the object type display name
			var objectType string
			var emoji string
			if info.TableType == "VIEW" {
				objectType = "view"
				emoji = "👁️"
			} else {
				objectType = "table"
				emoji = ""
			}

			if info.TableType == "BASE TABLE" && tableRows.Valid && tableRows.Int64 > 0 {
				info.RowCount = tableRows.Int64
				info.Description = fmt.Sprintf("%s %s • ~%d rows", emoji, strings.Title(objectType), info.RowCount)
			} else {
				info.Description = fmt.Sprintf("%s %s", emoji, strings.Title(objectType))
			}

			tableInfos = append(tableInfos, info)
		}

	case "sqlite3":
		// SQLite: Get both tables and views from sqlite_master
		query := `
			SELECT name, type
			FROM sqlite_master
			WHERE type IN ('table', 'view')
				AND name NOT LIKE 'sqlite_%'
			ORDER BY type, name`

		rows, err := db.Query(query)
		if err != nil {
			return GetSimpleTableInfos(db, driver, schema)
		}
		defer rows.Close()

		for rows.Next() {
			var name, objType string
			err := rows.Scan(&name, &objType)
			if err != nil {
				continue
			}

			info := models.TableInfo{
				Name:   name,
				Schema: "main", // SQLite uses "main" as the default schema
			}

			// Determine the object type display name
			var objectType string
			var emoji string
			if objType == "view" {
				info.TableType = "VIEW"
				objectType = "view"
				emoji = "👁️"
			} else {
				info.TableType = "BASE TABLE"
				objectType = "table"
				emoji = ""
			}

			// Try to get row count for tables only (views don't have meaningful row counts)
			if objType == "table" {
				countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, name)
				var count int64
				err := db.QueryRow(countQuery).Scan(&count)
				if err == nil {
					info.RowCount = count
					info.Description = fmt.Sprintf("%s %s • %d rows", emoji, strings.Title(objectType), count)
				} else {
					info.Description = fmt.Sprintf("%s %s", emoji, strings.Title(objectType))
				}
			} else {
				info.Description = fmt.Sprintf("%s %s", emoji, strings.Title(objectType))
			}

			tableInfos = append(tableInfos, info)
		}

	default:
		return GetSimpleTableInfos(db, driver, schema)
	}

	return tableInfos, nil
}

// GetSimpleTableInfos provides a fallback for basic table information
func GetSimpleTableInfos(db *sql.DB, driver, schema string) ([]models.TableInfo, error) {
	tables, err := GetTables(db, driver)
	if err != nil {
		return nil, err
	}

	var tableInfos []models.TableInfo
	for _, tableName := range tables {
		var schemaName string
		switch driver {
		case "postgres":
			schemaName = schema
		case "mysql":
			schemaName = "mysql" // Default schema name for MySQL
		case "sqlite3":
			schemaName = "main"
		default:
			schemaName = ""
		}

		tableInfos = append(tableInfos, models.TableInfo{
			Name:        tableName,
			Schema:      schemaName,
			TableType:   "BASE TABLE",
			Description: "Table",
		})
	}

	return tableInfos, nil
}
