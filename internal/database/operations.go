package database

import (
    "context"
    "database/sql"
    "fmt"
    "strings"
    "time"

	"github.com/danielcaldera/dbx/internal/models"
)

// ValidateConnectionString validates the format of connection strings
func ValidateConnectionString(driver, connectionStr string) error {
	if connectionStr == "" {
		return fmt.Errorf("connection string cannot be empty")
	}

	switch driver {
	case "postgres":
		if !strings.HasPrefix(connectionStr, "postgres://") && !strings.HasPrefix(connectionStr, "postgresql://") {
			return fmt.Errorf("PostgreSQL connection string should start with 'postgres://' or 'postgresql://'")
		}
	case "mysql":
		if strings.HasPrefix(connectionStr, "mysql://") {
			return fmt.Errorf("MySQL connection string should not include 'mysql://' prefix. Use format: user:password@tcp(host:port)/dbname")
		}
	case "sqlite3":
		return ValidateSQLiteConnection(connectionStr)
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	return nil
}

// ValidateSQLiteConnection validates SQLite database file
func ValidateSQLiteConnection(path string) error {
	if path == "" {
		return fmt.Errorf("SQLite database path cannot be empty")
	}

	if strings.Contains(path, "..") {
		return fmt.Errorf("relative paths with '..' are not allowed for security reasons")
	}

	if !strings.HasSuffix(path, ".db") && !strings.HasSuffix(path, ".sqlite") && !strings.HasSuffix(path, ".sqlite3") {
		return fmt.Errorf("SQLite file should have .db, .sqlite, or .sqlite3 extension")
	}

	return nil
}

// TestConnection tests a database connection with timeout
func TestConnectionWithTimeout(driver, connectionStr string) models.TestConnectionResult {
	timeout := 10 * time.Second
	done := make(chan models.TestConnectionResult, 1)

	go func() {
		db, err := sql.Open(driver, connectionStr)
		if err != nil {
			done <- models.TestConnectionResult{Success: false, Err: err}
			return
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		err = db.PingContext(ctx)
		if err != nil {
			done <- models.TestConnectionResult{Success: false, Err: err}
			return
		}

		done <- models.TestConnectionResult{Success: true, Err: nil}
	}()

	select {
	case result := <-done:
		return result
	case <-time.After(timeout):
		return models.TestConnectionResult{
			Success: false,
			Err:     fmt.Errorf("connection timeout after %v", timeout),
		}
	}
}

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
			return []models.SchemaInfo{{"public", "Default public schema"}}, nil
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
				emoji = "📊"
			}

			if info.TableType == "BASE TABLE" && estimatedRows.Valid && estimatedRows.Int64 > 0 {
				info.RowCount = estimatedRows.Int64
				if info.RowCount < 0 {
					info.RowCount = 0
				}
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
				emoji = "📊"
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
				emoji = "📊"
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
			Description: "📊 Table",
		})
	}

	return tableInfos, nil
}

// GetTablePreview returns first N rows from a table/view with column names
func GetTablePreview(db *sql.DB, driver, tableName, schema string, limit int) ([]string, [][]string, error) {
    if limit <= 0 {
        limit = 10
    }

    var query string
    switch driver {
    case "postgres":
        if schema == "" {
            schema = "public"
        }
        query = fmt.Sprintf("SELECT * FROM \"%s\".\"%s\" LIMIT %d", schema, tableName, limit)
    case "mysql":
        query = fmt.Sprintf("SELECT * FROM `%s` LIMIT %d", tableName, limit)
    case "sqlite3":
        query = fmt.Sprintf("SELECT * FROM \"%s\" LIMIT %d", tableName, limit)
    default:
        return nil, nil, fmt.Errorf("unsupported driver: %s", driver)
    }

    rows, err := db.Query(query)
    if err != nil {
        return nil, nil, err
    }
    defer rows.Close()

    cols, err := rows.Columns()
    if err != nil {
        return nil, nil, err
    }

    var result [][]string
    for rows.Next() {
        values := make([]interface{}, len(cols))
        valuePtrs := make([]interface{}, len(cols))
        for i := range values {
            valuePtrs[i] = &values[i]
        }
        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, nil, err
        }
        record := make([]string, len(cols))
        for i, v := range values {
            switch t := v.(type) {
            case nil:
                record[i] = "NULL"
            case []byte:
                record[i] = string(t)
            default:
                record[i] = fmt.Sprintf("%v", t)
            }
        }
        result = append(result, record)
    }
    return cols, result, nil
}
