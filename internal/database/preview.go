package database

import (
	"database/sql"
	"fmt"
	"strings"
)

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

// GetTableRowCount returns the total number of rows in a table
func GetTableRowCount(db *sql.DB, driver, tableName, schema string) (int, error) {
	var query string
	switch driver {
	case "postgres":
		if schema == "" {
			schema = "public"
		}
		query = fmt.Sprintf("SELECT COUNT(*) FROM \"%s\".\"%s\"", schema, tableName)
	case "mysql":
		query = fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	case "sqlite3":
		query = fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", tableName)
	default:
		return 0, fmt.Errorf("unsupported driver: %s", driver)
	}

	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetTablePreviewPaginated returns paginated rows from a table/view with column names
func GetTablePreviewPaginated(db *sql.DB, driver, tableName, schema string, limit, offset int) ([]string, [][]string, error) {
	return GetTablePreviewPaginatedWithSort(db, driver, tableName, schema, limit, offset, "", "")
}

// GetTablePreviewPaginatedWithSort returns paginated rows from a table/view with column names and optional sorting
func GetTablePreviewPaginatedWithSort(db *sql.DB, driver, tableName, schema string, limit, offset int, sortColumn, sortDirection string) ([]string, [][]string, error) {
	if limit <= 0 {
		limit = 10
	}
	offset = max(offset, 0)

	var query string
	var orderBy string
	if sortColumn != "" && sortDirection != "" {
		switch driver {
		case "postgres":
			orderBy = fmt.Sprintf(" ORDER BY \"%s\" %s", sortColumn, sortDirection)
		case "mysql":
			orderBy = fmt.Sprintf(" ORDER BY `%s` %s", sortColumn, sortDirection)
		case "sqlite3":
			orderBy = fmt.Sprintf(" ORDER BY \"%s\" %s", sortColumn, sortDirection)
		}
	}

	switch driver {
	case "postgres":
		if schema == "" {
			schema = "public"
		}
		query = fmt.Sprintf("SELECT * FROM \"%s\".\"%s\"%s LIMIT %d OFFSET %d", schema, tableName, orderBy, limit, offset)
	case "mysql":
		query = fmt.Sprintf("SELECT * FROM `%s`%s LIMIT %d OFFSET %d", tableName, orderBy, limit, offset)
	case "sqlite3":
		query = fmt.Sprintf("SELECT * FROM \"%s\"%s LIMIT %d OFFSET %d", tableName, orderBy, limit, offset)
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

// GetTableRowCountWithFilter returns the total number of rows in a table with filter applied
func GetTableRowCountWithFilter(db *sql.DB, driver, tableName, schema, filterValue string, columns []string) (int, error) {
	if filterValue == "" {
		return GetTableRowCount(db, driver, tableName, schema)
	}

	var query string
	switch driver {
	case "postgres":
		if schema == "" {
			schema = "public"
		}
		// Build WHERE clause with OR conditions for each column
		whereConditions := make([]string, len(columns))
		for i, col := range columns {
			whereConditions[i] = fmt.Sprintf("(\"%s\"::TEXT ILIKE '%%%s%%')", col, filterValue)
		}
		whereClause := strings.Join(whereConditions, " OR ")
		query = fmt.Sprintf("SELECT COUNT(*) FROM \"%s\".\"%s\" WHERE %s", schema, tableName, whereClause)
	case "mysql":
		// Build WHERE clause with OR conditions for each column
		whereConditions := make([]string, len(columns))
		for i, col := range columns {
			whereConditions[i] = fmt.Sprintf("(CAST(`%s` AS CHAR) LIKE '%%%s%%')", col, filterValue)
		}
		whereClause := strings.Join(whereConditions, " OR ")
		query = fmt.Sprintf("SELECT COUNT(*) FROM `%s` WHERE %s", tableName, whereClause)
	case "sqlite3":
		// Build WHERE clause with OR conditions for each column
		whereConditions := make([]string, len(columns))
		for i, col := range columns {
			whereConditions[i] = fmt.Sprintf("(CAST(\"%s\" AS TEXT) LIKE '%%%s%%')", col, filterValue)
		}
		whereClause := strings.Join(whereConditions, " OR ")
		query = fmt.Sprintf("SELECT COUNT(*) FROM \"%s\" WHERE %s", tableName, whereClause)
	default:
		return 0, fmt.Errorf("unsupported driver: %s", driver)
	}

	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetTablePreviewPaginatedWithFilter returns paginated rows from a table/view with filter applied
func GetTablePreviewPaginatedWithFilter(db *sql.DB, driver, tableName, schema string, limit, offset int, filterValue string, columns []string) ([]string, [][]string, error) {
	return GetTablePreviewPaginatedWithFilterAndSort(db, driver, tableName, schema, limit, offset, filterValue, columns, "", "")
}

// GetTablePreviewPaginatedWithFilterAndSort returns paginated rows from a table/view with filter and sort applied
func GetTablePreviewPaginatedWithFilterAndSort(db *sql.DB, driver, tableName, schema string, limit, offset int, filterValue string, columns []string, sortColumn, sortDirection string) ([]string, [][]string, error) {
	if filterValue == "" {
		return GetTablePreviewPaginatedWithSort(db, driver, tableName, schema, limit, offset, sortColumn, sortDirection)
	}

	if limit <= 0 {
		limit = 25
	}
	offset = max(offset, 0)

	var orderBy string
	if sortColumn != "" && sortDirection != "" {
		switch driver {
		case "postgres":
			orderBy = fmt.Sprintf(" ORDER BY \"%s\" %s", sortColumn, sortDirection)
		case "mysql":
			orderBy = fmt.Sprintf(" ORDER BY `%s` %s", sortColumn, sortDirection)
		case "sqlite3":
			orderBy = fmt.Sprintf(" ORDER BY \"%s\" %s", sortColumn, sortDirection)
		}
	}

	var query string
	switch driver {
	case "postgres":
		if schema == "" {
			schema = "public"
		}
		// Build WHERE clause with OR conditions for each column
		whereConditions := make([]string, len(columns))
		for i, col := range columns {
			whereConditions[i] = fmt.Sprintf("(\"%s\"::TEXT ILIKE '%%%s%%')", col, filterValue)
		}
		whereClause := strings.Join(whereConditions, " OR ")
		query = fmt.Sprintf("SELECT * FROM \"%s\".\"%s\" WHERE %s%s LIMIT %d OFFSET %d", schema, tableName, whereClause, orderBy, limit, offset)
	case "mysql":
		// Build WHERE clause with OR conditions for each column
		whereConditions := make([]string, len(columns))
		for i, col := range columns {
			whereConditions[i] = fmt.Sprintf("(CAST(`%s` AS CHAR) LIKE '%%%s%%')", col, filterValue)
		}
		whereClause := strings.Join(whereConditions, " OR ")
		query = fmt.Sprintf("SELECT * FROM `%s` WHERE %s%s LIMIT %d OFFSET %d", tableName, whereClause, orderBy, limit, offset)
	case "sqlite3":
		// Build WHERE clause with OR conditions for each column
		whereConditions := make([]string, len(columns))
		for i, col := range columns {
			whereConditions[i] = fmt.Sprintf("(CAST(\"%s\" AS TEXT) LIKE '%%%s%%')", col, filterValue)
		}
		whereClause := strings.Join(whereConditions, " OR ")
		query = fmt.Sprintf("SELECT * FROM \"%s\" WHERE %s%s LIMIT %d OFFSET %d", tableName, whereClause, orderBy, limit, offset)
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
