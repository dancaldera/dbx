package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// GetColumns retrieves column information for a specific table
func GetColumns(db *sql.DB, driver, tableName, schema string) ([][]string, error) {
	var query string
	switch driver {
	case "postgres":
		query = `SELECT column_name, data_type, is_nullable, column_default 
				 FROM information_schema.columns 
				 WHERE table_name = $1 AND table_schema = $2
				 ORDER BY ordinal_position`
	case "mysql":
		query = `SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT 
				 FROM INFORMATION_SCHEMA.COLUMNS 
				 WHERE TABLE_NAME = ? 
				 ORDER BY ORDINAL_POSITION`
	case "sqlite3":
		query = fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	}

	var rows *sql.Rows
	var err error

	switch driver {
	case "postgres":
		rows, err = db.Query(query, tableName, schema)
	case "mysql":
		rows, err = db.Query(query, tableName)
	case "sqlite3":
		rows, err = db.Query(query)
	default:
		rows, err = db.Query(query, tableName)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns [][]string
	for rows.Next() {
		if driver == "sqlite3" {
			var cid int
			var name, dataType string
			var notNull int
			var defaultValue sql.NullString
			var pk int

			err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
			if err != nil {
				return nil, err
			}

			nullable := "YES"
			if notNull == 1 {
				nullable = "NO"
			}

			def := ""
			if defaultValue.Valid {
				def = defaultValue.String
			}

			columns = append(columns, []string{name, dataType, nullable, def})
		} else {
			var name, dataType, nullable string
			var defaultValue sql.NullString

			err := rows.Scan(&name, &dataType, &nullable, &defaultValue)
			if err != nil {
				return nil, err
			}

			def := ""
			if defaultValue.Valid {
				def = defaultValue.String
			}

			columns = append(columns, []string{name, dataType, nullable, def})
		}
	}

	return columns, nil
}

// GetIndexes retrieves index information for a specific table
func GetIndexes(db *sql.DB, driver, tableName, schema string) ([][]string, error) {
	var query string
	switch driver {
	case "postgres":
		query = `SELECT 
					i.indexname as index_name,
					i.indexdef as index_definition,
					CASE 
						WHEN i.indexdef LIKE '%UNIQUE%' THEN 'UNIQUE'
						WHEN i.indexdef LIKE '%PRIMARY%' THEN 'PRIMARY'
						ELSE 'INDEX'
					END as index_type,
					pg_get_indexdef(pg_class.oid) as columns
				FROM pg_indexes i
				JOIN pg_class ON pg_class.relname = i.indexname
				WHERE i.tablename = $1 AND i.schemaname = $2
				ORDER BY i.indexname`
	case "mysql":
		query = `SELECT 
					INDEX_NAME as index_name,
					CONCAT('INDEX ON ', COLUMN_NAME) as index_definition,
					CASE 
						WHEN NON_UNIQUE = 0 AND INDEX_NAME = 'PRIMARY' THEN 'PRIMARY'
						WHEN NON_UNIQUE = 0 THEN 'UNIQUE'
						ELSE 'INDEX'
					END as index_type,
					GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX) as columns
				FROM INFORMATION_SCHEMA.STATISTICS 
				WHERE TABLE_NAME = ? AND TABLE_SCHEMA = DATABASE()
				GROUP BY INDEX_NAME, NON_UNIQUE
				ORDER BY INDEX_NAME`
	case "sqlite3":
		query = fmt.Sprintf("PRAGMA index_list(%s)", tableName)
	}

	var rows *sql.Rows
	var err error

	switch driver {
	case "postgres":
		rows, err = db.Query(query, tableName, schema)
	case "mysql":
		rows, err = db.Query(query, tableName)
	case "sqlite3":
		rows, err = db.Query(query)
	default:
		rows, err = db.Query(query, tableName)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes [][]string
	for rows.Next() {
		if driver == "sqlite3" {
			var seq int
			var name string
			var unique int
			var origin, partial string

			err := rows.Scan(&seq, &name, &unique, &origin, &partial)
			if err != nil {
				return nil, err
			}

			indexType := "INDEX"
			if unique == 1 {
				indexType = "UNIQUE"
			}

			// Get columns for this index
			indexInfoQuery := fmt.Sprintf("PRAGMA index_info(%s)", name)
			indexInfoRows, err := db.Query(indexInfoQuery)
			if err != nil {
				continue
			}

			var columns []string
			for indexInfoRows.Next() {
				var seqno, cid int
				var colName string
				if err := indexInfoRows.Scan(&seqno, &cid, &colName); err == nil {
					columns = append(columns, colName)
				}
			}
			indexInfoRows.Close()

			columnsStr := strings.Join(columns, ", ")
			definition := fmt.Sprintf("INDEX ON (%s)", columnsStr)

			indexes = append(indexes, []string{name, indexType, columnsStr, definition})
		} else {
			var name, definition, indexType, columns string

			err := rows.Scan(&name, &definition, &indexType, &columns)
			if err != nil {
				return nil, err
			}

			indexes = append(indexes, []string{name, indexType, columns, definition})
		}
	}

	return indexes, nil
}

// GetConstraints retrieves constraint information for a specific table
func GetConstraints(db *sql.DB, driver, tableName, schema string) ([][]string, error) {
	var query string
	switch driver {
	case "postgres":
		query = `SELECT 
					tc.constraint_name,
					tc.constraint_type,
					kcu.column_name,
					COALESCE(ccu.table_name || '.' || ccu.column_name, '') as referenced_table_column
				FROM information_schema.table_constraints tc
				LEFT JOIN information_schema.key_column_usage kcu 
					ON tc.constraint_name = kcu.constraint_name 
					AND tc.table_schema = kcu.table_schema
				LEFT JOIN information_schema.constraint_column_usage ccu
					ON tc.constraint_name = ccu.constraint_name
					AND tc.table_schema = ccu.table_schema
				WHERE tc.table_name = $1 AND tc.table_schema = $2
				ORDER BY tc.constraint_name, kcu.ordinal_position`
	case "mysql":
		query = `SELECT 
					CONSTRAINT_NAME as constraint_name,
					CONSTRAINT_TYPE as constraint_type,
					COLUMN_NAME as column_name,
					COALESCE(CONCAT(REFERENCED_TABLE_NAME, '.', REFERENCED_COLUMN_NAME), '') as referenced_table_column
				FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
				JOIN INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc 
					ON kcu.CONSTRAINT_NAME = tc.CONSTRAINT_NAME 
					AND kcu.TABLE_SCHEMA = tc.TABLE_SCHEMA
				WHERE kcu.TABLE_NAME = ? AND kcu.TABLE_SCHEMA = DATABASE()
				ORDER BY kcu.CONSTRAINT_NAME, kcu.ORDINAL_POSITION`
	case "sqlite3":
		query = fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)
	}

	var rows *sql.Rows
	var err error

	switch driver {
	case "postgres":
		rows, err = db.Query(query, tableName, schema)
	case "mysql":
		rows, err = db.Query(query, tableName)
	case "sqlite3":
		rows, err = db.Query(query)
	default:
		rows, err = db.Query(query, tableName)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var constraints [][]string
	if driver == "sqlite3" {
		// Handle SQLite foreign keys
		for rows.Next() {
			var id, seq int
			var table, from, to, onUpdate, onDelete, match string

			err := rows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match)
			if err != nil {
				return nil, err
			}

			constraintName := fmt.Sprintf("fk_%s_%s", tableName, from)
			constraintType := "FOREIGN KEY"
			referencedTableColumn := fmt.Sprintf("%s.%s", table, to)

			constraints = append(constraints, []string{constraintName, constraintType, from, referencedTableColumn})
		}
	} else {
		for rows.Next() {
			var name, constraintType, column, referencedTableColumn string

			err := rows.Scan(&name, &constraintType, &column, &referencedTableColumn)
			if err != nil {
				return nil, err
			}

			constraints = append(constraints, []string{name, constraintType, column, referencedTableColumn})
		}
	}

	return constraints, nil
}

// GetForeignKeyRelationships retrieves all foreign key relationships in the database
func GetForeignKeyRelationships(db *sql.DB, driver, schema string) ([][]string, error) {
	var query string
	var args []interface{}

	switch driver {
	case "postgres":
		query = `
			SELECT 
				tc.table_name as from_table,
				kcu.column_name as from_column,
				ccu.table_name as to_table,
				ccu.column_name as to_column,
				tc.constraint_name
			FROM 
				information_schema.table_constraints AS tc 
				JOIN information_schema.key_column_usage AS kcu
					ON tc.constraint_name = kcu.constraint_name
					AND tc.table_schema = kcu.table_schema
				JOIN information_schema.constraint_column_usage AS ccu
					ON ccu.constraint_name = tc.constraint_name
					AND ccu.table_schema = tc.table_schema
			WHERE tc.constraint_type = 'FOREIGN KEY' 
				AND tc.table_schema = $1
			ORDER BY tc.table_name, kcu.ordinal_position`
		args = []interface{}{schema}

	case "mysql":
		query = `
			SELECT 
				TABLE_NAME as from_table,
				COLUMN_NAME as from_column,
				REFERENCED_TABLE_NAME as to_table,
				REFERENCED_COLUMN_NAME as to_column,
				CONSTRAINT_NAME
			FROM 
				INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
			WHERE 
				REFERENCED_TABLE_NAME IS NOT NULL
				AND TABLE_SCHEMA = DATABASE()
			ORDER BY TABLE_NAME, ORDINAL_POSITION`
		args = []interface{}{}

	case "sqlite3":
		// For SQLite, we need to get foreign keys from all tables
		tableQuery := "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"
		tableRows, err := db.Query(tableQuery)
		if err != nil {
			return nil, err
		}
		defer tableRows.Close()

		var relationships [][]string
		for tableRows.Next() {
			var tableName string
			if err := tableRows.Scan(&tableName); err != nil {
				continue
			}

			// Get foreign keys for this table
			fkQuery := fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)
			fkRows, err := db.Query(fkQuery)
			if err != nil {
				continue
			}

			for fkRows.Next() {
				var id, seq int
				var referencedTable, fromColumn, toColumn, onUpdate, onDelete, match string

				err := fkRows.Scan(&id, &seq, &referencedTable, &fromColumn, &toColumn, &onUpdate, &onDelete, &match)
				if err != nil {
					continue
				}

				constraintName := fmt.Sprintf("fk_%s_%s", tableName, fromColumn)
				relationships = append(relationships, []string{tableName, fromColumn, referencedTable, toColumn, constraintName})
			}
			fkRows.Close()
		}

		return relationships, nil

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relationships [][]string
	for rows.Next() {
		var fromTable, fromColumn, toTable, toColumn, constraintName string

		err := rows.Scan(&fromTable, &fromColumn, &toTable, &toColumn, &constraintName)
		if err != nil {
			return nil, err
		}

		relationships = append(relationships, []string{fromTable, fromColumn, toTable, toColumn, constraintName})
	}

	return relationships, nil
}
