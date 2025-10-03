package utils

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancaldera/dbx/internal/database"
	"github.com/dancaldera/dbx/internal/models"
)

// GetDefaultSchema returns the default schema name for a database driver
func GetDefaultSchema(driver string) string {
	switch driver {
	case "mysql":
		return "mysql"
	case "sqlite3":
		return "main"
	default: // postgres
		return "public"
	}
}

// BuildUpdateSQL generates database-specific UPDATE SQL statement
func BuildUpdateSQL(driver, schema, table, field, primaryKey string) string {
	switch driver {
	case "postgres":
		return fmt.Sprintf(`UPDATE "%s"."%s" SET "%s" = $1 WHERE "%s" = $2`,
			schema, table, field, primaryKey)
	case "mysql":
		return fmt.Sprintf("UPDATE `%s`.`%s` SET `%s` = ? WHERE `%s` = ?",
			schema, table, field, primaryKey)
	case "sqlite3":
		return fmt.Sprintf(`UPDATE "%s" SET "%s" = ? WHERE "%s" = ?`,
			table, field, primaryKey)
	default:
		return fmt.Sprintf(`UPDATE "%s"."%s" SET "%s" = $1 WHERE "%s" = $2`,
			schema, table, field, primaryKey)
	}
}

// DetermineSortParameters converts sort direction and column to database parameters
func DetermineSortParameters(sortDirection models.SortDirection, sortColumn string) (string, string) {
	switch sortDirection {
	case models.SortAsc:
		return sortColumn, "ASC"
	case models.SortDesc:
		return sortColumn, "DESC"
	default:
		return "", ""
	}
}

// FindPrimaryKeyColumn locates primary key column and value from row data
func FindPrimaryKeyColumn(columns []string, rowData []string) (string, string, error) {
	// Look for common primary key patterns
	for i, col := range columns {
		if col == "id" || col == "Id" || col == "ID" {
			if i < len(rowData) {
				return col, rowData[i], nil
			}
		}
	}

	// Try secondary patterns
	for i, col := range columns {
		colLower := strings.ToLower(col)
		if strings.HasSuffix(colLower, "_id") || strings.HasSuffix(colLower, "id") {
			if i < len(rowData) {
				return col, rowData[i], nil
			}
		}
	}

	return "", "", fmt.Errorf("no primary key column found in %d columns", len(columns))
}

// ConnectToDB establishes database connection and loads tables
func ConnectToDB(selectedDB models.DBType, connectionStr string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		db, err := sql.Open(selectedDB.Driver, connectionStr)
		if err != nil {
			return models.ConnectResult{Err: err}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = db.PingContext(ctx)
		if err != nil {
			db.Close()
			return models.ConnectResult{Err: err}
		}

		tables, err := database.GetTables(db, selectedDB.Driver)
		if err != nil {
			db.Close()
			return models.ConnectResult{Err: err}
		}

		schema := GetDefaultSchema(selectedDB.Driver)

		return models.ConnectResult{
			DB:     db,
			Driver: selectedDB.Driver,
			Tables: tables,
			Schema: schema,
		}
	})
}

// LoadColumns loads column information for a table
func LoadColumns(db *sql.DB, selectedDB models.DBType, selectedTable, selectedSchema string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		columns, err := database.GetColumns(db, selectedDB.Driver, selectedTable, selectedSchema)
		return models.ColumnsResult{
			Columns: columns,
			Err:     err,
		}
	})
}

// LoadDataPreview loads table data preview with pagination and sorting
func LoadDataPreview(db *sql.DB, selectedDB models.DBType, selectedTable, selectedSchema string, itemsPerPage int, sortDirection models.SortDirection, sortColumn string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Reset pagination and load first page
		totalRows, err := database.GetTableRowCount(db, selectedDB.Driver, selectedTable, selectedSchema)
		if err != nil {
			return models.DataPreviewResult{Columns: nil, Rows: nil, Err: err}
		}

		// Determine sort parameters
		sortCol, sortDir := DetermineSortParameters(sortDirection, sortColumn)

		cols, rows, err := database.GetTablePreviewPaginatedWithSort(db, selectedDB.Driver, selectedTable, selectedSchema, itemsPerPage, 0, sortCol, sortDir)
		return models.DataPreviewResult{Columns: cols, Rows: rows, Err: err, TotalRows: totalRows}
	})
}

// LoadDataPreviewWithPagination loads data with pagination support
func LoadDataPreviewWithPagination(db *sql.DB, selectedDB models.DBType, selectedTable, selectedSchema string, itemsPerPage, currentPage int, sortDirection models.SortDirection, sortColumn, filterValue string, allColumns []string, totalRows int) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Determine sort parameters
		sortCol, sortDir := DetermineSortParameters(sortDirection, sortColumn)

		offset := currentPage * itemsPerPage
		if filterValue != "" {
			cols, rows, err := database.GetTablePreviewPaginatedWithFilterAndSort(db, selectedDB.Driver, selectedTable, selectedSchema, itemsPerPage, offset, filterValue, allColumns, sortCol, sortDir)
			return models.DataPreviewResult{Columns: cols, Rows: rows, Err: err, TotalRows: totalRows}
		}
		cols, rows, err := database.GetTablePreviewPaginatedWithSort(db, selectedDB.Driver, selectedTable, selectedSchema, itemsPerPage, offset, sortCol, sortDir)
		return models.DataPreviewResult{Columns: cols, Rows: rows, Err: err, TotalRows: totalRows}
	})
}

// LoadDataPreviewWithFilter loads data with filter applied
func LoadDataPreviewWithFilter(db *sql.DB, selectedDB models.DBType, selectedTable, selectedSchema string, itemsPerPage int, filterValue string, allColumns []string, sortDirection models.SortDirection, sortColumn string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Get total rows with filter
		totalRows, err := database.GetTableRowCountWithFilter(db, selectedDB.Driver, selectedTable, selectedSchema, filterValue, allColumns)
		if err != nil {
			return models.DataPreviewResult{Columns: nil, Rows: nil, Err: err}
		}

		// Determine sort parameters
		sortCol, sortDir := DetermineSortParameters(sortDirection, sortColumn)

		// Get filtered and sorted data
		cols, rows, err := database.GetTablePreviewPaginatedWithFilterAndSort(db, selectedDB.Driver, selectedTable, selectedSchema, itemsPerPage, 0, filterValue, allColumns, sortCol, sortDir)
		return models.DataPreviewResult{Columns: cols, Rows: rows, Err: err, TotalRows: totalRows}
	})
}

// LoadDataPreviewWithSort loads data with sorting applied
func LoadDataPreviewWithSort(db *sql.DB, selectedDB models.DBType, selectedTable, selectedSchema string, itemsPerPage, currentPage int, sortDirection models.SortDirection, sortColumn, filterValue string, allColumns []string, totalRows int) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Determine sort parameters
		sortCol, sortDir := DetermineSortParameters(sortDirection, sortColumn)

		offset := currentPage * itemsPerPage

		// Use appropriate function based on whether filter is active
		if filterValue != "" {
			cols, rows, err := database.GetTablePreviewPaginatedWithFilterAndSort(db, selectedDB.Driver, selectedTable, selectedSchema, itemsPerPage, offset, filterValue, allColumns, sortCol, sortDir)
			return models.DataPreviewResult{Columns: cols, Rows: rows, Err: err, TotalRows: totalRows}
		} else {
			cols, rows, err := database.GetTablePreviewPaginatedWithSort(db, selectedDB.Driver, selectedTable, selectedSchema, itemsPerPage, offset, sortCol, sortDir)
			return models.DataPreviewResult{Columns: cols, Rows: rows, Err: err, TotalRows: totalRows}
		}
	})
}

// LoadRelationships loads foreign key relationships
func LoadRelationships(db *sql.DB, selectedDB models.DBType, selectedSchema string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		rels, err := database.GetForeignKeyRelationships(db, selectedDB.Driver, selectedSchema)
		return models.RelationshipsResult{Relationships: rels, Err: err}
	})
}

// SaveFieldEdit creates and executes an UPDATE statement for the edited field
func SaveFieldEdit(db *sql.DB, selectedDB models.DBType, selectedSchema, selectedTable, editingFieldName string, allColumns, selectedRowData []string, editingFieldIndex int, newValue string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Find the primary key column and value for the WHERE clause
		primaryKeyColumn, primaryKeyValue, err := FindPrimaryKeyColumn(allColumns, selectedRowData)
		if err != nil {
			return models.FieldUpdateResult{
				Success:  false,
				Err:      err,
				ExitEdit: false,
			}
		}

		// Build UPDATE SQL statement
		updateSQL := BuildUpdateSQL(selectedDB.Driver, selectedSchema, selectedTable, editingFieldName, primaryKeyColumn)

		// Execute the UPDATE statement
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := db.ExecContext(ctx, updateSQL, newValue, primaryKeyValue)
		if err != nil {
			return models.FieldUpdateResult{
				Success:  false,
				Err:      fmt.Errorf("failed to update field: %v", err),
				ExitEdit: false,
			}
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return models.FieldUpdateResult{
				Success:  false,
				Err:      fmt.Errorf("failed to get affected rows: %v", err),
				ExitEdit: false,
			}
		}

		if rowsAffected == 0 {
			return models.FieldUpdateResult{
				Success:  false,
				Err:      fmt.Errorf("no rows were updated - record may not exist"),
				ExitEdit: false,
			}
		}

		return models.FieldUpdateResult{
			Success:  true,
			ExitEdit: true,
			NewValue: newValue,
		}
	})
}

// HandleConnectResult processes database connection result and updates model
func HandleConnectResult(m models.Model, msg models.ConnectResult) (models.Model, tea.Cmd) {
	updatedModel := m
	updatedModel.IsConnecting = false

	if msg.Err != nil {
		// Ensure we stay in SavedConnectionsView to display the error
		updatedModel.State = models.SavedConnectionsView
		return SetErrorWithTimeout(updatedModel, msg.Err, 3*time.Second)
	}

	updatedModel.DB = msg.DB
	updatedModel.Tables = msg.Tables
	updatedModel.SelectedSchema = msg.Schema

	// Sort tables alphabetically
	sort.Strings(updatedModel.Tables)

	// Create simple table infos
	updatedModel.TableInfos = CreateTableInfos(updatedModel.Tables, updatedModel.SelectedSchema)

	// Update tables list (show only table names)
	items := CreateTableListItems(updatedModel.TableInfos)
	updatedModel.TablesList.SetItems(items)

	updatedModel.State = models.TablesView
	return updatedModel, nil
}

// CreateTableInfos creates TableInfo objects from table names
func CreateTableInfos(tables []string, schema string) []models.TableInfo {
	infos := make([]models.TableInfo, len(tables))
	for i, table := range tables {
		infos[i] = models.TableInfo{
			Name:   table,
			Schema: schema,
		}
	}
	return infos
}

// CreateTableListItems creates list items from table infos
func CreateTableListItems(infos []models.TableInfo) []list.Item {
	items := make([]list.Item, len(infos))
	for i, info := range infos {
		items[i] = models.Item{
			ItemTitle: info.Name,
			ItemDesc:  fmt.Sprintf("Table in %s schema", info.Schema),
		}
	}
	return items
}

// HandleTestConnectionResult processes test connection result and updates model
func HandleTestConnectionResult(m models.Model, msg models.TestConnectionResult) (models.Model, tea.Cmd) {
	updatedModel := m
	updatedModel.IsTestingConnection = false

	if msg.Err != nil {
		return SetErrorWithTimeout(updatedModel, msg.Err, 3*time.Second)
	}

	updatedModel.QueryResult = "Connection successful!"
	return updatedModel, ClearResultAfterTimeout()
}

// HandleColumnsResult processes columns result and updates model
func HandleColumnsResult(m models.Model, msg models.ColumnsResult) (models.Model, tea.Cmd) {
	updatedModel := m
	updatedModel.IsLoadingColumns = false

	if msg.Err != nil {
		return SetErrorWithTimeout(updatedModel, msg.Err, 3*time.Second)
	}

	// Convert columns to table rows (msg.Columns is [][]string)
	rows := make([]table.Row, len(msg.Columns))
	for i, col := range msg.Columns {
		if len(col) >= 4 {
			rows[i] = table.Row{col[0], col[1], col[2], col[3]}
		} else if len(col) >= 3 {
			rows[i] = table.Row{col[0], col[1], col[2], ""}
		} else if len(col) >= 2 {
			rows[i] = table.Row{col[0], col[1], "", ""}
		} else if len(col) >= 1 {
			rows[i] = table.Row{col[0], "", "", ""}
		} else {
			rows[i] = table.Row{"", "", "", ""}
		}
	}

	// Update columns table
	updatedModel.ColumnsTable.SetRows(rows)
	updatedModel.State = models.ColumnsView
	return updatedModel, nil
}

// HandleDataPreviewResult processes data preview result and updates model
func HandleDataPreviewResult(m models.Model, msg models.DataPreviewResult) (models.Model, tea.Cmd) {
	updatedModel := m
	updatedModel.IsLoadingPreview = false

	if msg.Err != nil {
		return SetErrorWithTimeout(updatedModel, msg.Err, 3*time.Second)
	}

	updatedModel.DataPreviewAllColumns = msg.Columns
	updatedModel.DataPreviewAllRows = msg.Rows
	updatedModel.DataPreviewTotalRows = msg.TotalRows

	// Create the data preview table
	updatedModel = CreateDataPreviewTable(updatedModel)

	// Switch to data preview view to show the table
	updatedModel.State = models.DataPreviewView
	return updatedModel, nil
}

// HandleRelationshipsResult processes relationships result and updates model
func HandleRelationshipsResult(m models.Model, msg models.RelationshipsResult) (models.Model, tea.Cmd) {
	updatedModel := m

	if msg.Err != nil {
		return SetErrorWithTimeout(updatedModel, msg.Err, 3*time.Second)
	}

	// Convert relationships to table rows
	rows := make([]table.Row, len(msg.Relationships))
	for i, rel := range msg.Relationships {
		if len(rel) >= 4 {
			rows[i] = table.Row{rel[0], rel[1], rel[2], rel[3]}
		} else if len(rel) >= 3 {
			rows[i] = table.Row{rel[0], rel[1], rel[2], ""}
		} else if len(rel) >= 2 {
			rows[i] = table.Row{rel[0], rel[1], "", ""}
		} else if len(rel) >= 1 {
			rows[i] = table.Row{rel[0], "", "", ""}
		} else {
			rows[i] = table.Row{"", "", "", ""}
		}
	}

	// Update relationships table
	updatedModel.RelationshipsTable.SetRows(rows)
	updatedModel.State = models.RelationshipsView
	return updatedModel, nil
}

// HandleFieldUpdateResult processes field update result and updates model
func HandleFieldUpdateResult(m models.Model, msg models.FieldUpdateResult) (models.Model, tea.Cmd) {
	updatedModel := m

	if msg.Err != nil {
		return SetErrorWithTimeout(updatedModel, msg.Err, 3*time.Second)
	}

	if msg.Success {
		updatedModel.OriginalFieldValue = msg.NewValue
		if msg.ExitEdit {
			updatedModel.IsEditingField = false
			updatedModel.FieldTextarea.Blur()
			updatedModel.EditingFieldName = ""
			// Refresh data preview to show updated value
			return updatedModel, LoadDataPreview(m.DB, m.SelectedDB, m.SelectedTable, m.SelectedSchema, m.DataPreviewItemsPerPage, m.DataPreviewSortDirection, m.DataPreviewSortColumn)
		}
	}

	return updatedModel, nil
}

// ExecuteQuery executes a user-provided SQL query and returns results
func ExecuteQuery(db *sql.DB, selectedDB models.DBType, query string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Trim whitespace from query
		query = strings.TrimSpace(query)
		if query == "" {
			return models.QueryResultMsg{
				Result: "",
				Err:    fmt.Errorf("empty query"),
			}
		}

		// Check if it's a SELECT query (for read-only operations)
		isSelect := strings.HasPrefix(strings.ToUpper(query), "SELECT")

		if isSelect {
			// Execute SELECT query
			rows, err := db.Query(query)
			if err != nil {
				return models.QueryResultMsg{
					Result: "",
					Err:    err,
				}
			}
			defer rows.Close()

			// Get column names
			columns, err := rows.Columns()
			if err != nil {
				return models.QueryResultMsg{
					Result: "",
					Err:    err,
				}
			}

			// Prepare result variables
			values := make([]interface{}, len(columns))
			scanArgs := make([]interface{}, len(columns))
			for i := range values {
				scanArgs[i] = &values[i]
			}

			// Collect all rows
			var allRows [][]string
			rowCount := 0
			const maxRows = 1000 // Limit results to prevent memory issues

			for rows.Next() && rowCount < maxRows {
				err = rows.Scan(scanArgs...)
				if err != nil {
					return models.QueryResultMsg{
						Result: "",
						Err:    err,
					}
				}

				row := make([]string, len(columns))
				for i, val := range values {
					if val != nil {
						row[i] = fmt.Sprintf("%v", val)
					} else {
						row[i] = "NULL"
					}
				}
				allRows = append(allRows, row)
				rowCount++
			}

			if err = rows.Err(); err != nil {
				return models.QueryResultMsg{
					Result: "",
					Err:    err,
				}
			}

			// Create result message
			var result string
			if len(allRows) == 0 {
				result = "Query executed successfully. No rows returned."
			} else {
				if rowCount >= maxRows {
					result = fmt.Sprintf("Query executed successfully. Showing first %d rows out of more results.", maxRows)
				} else {
					result = fmt.Sprintf("Query executed successfully. Returned %d rows.", len(allRows))
				}
			}

			return models.QueryResultMsg{
				Result:  result,
				Columns: columns,
				Rows:    allRows,
				Err:     nil,
			}

		} else {
			// Execute non-SELECT query (INSERT, UPDATE, DELETE)
			result, err := db.Exec(query)
			if err != nil {
				return models.QueryResultMsg{
					Result: "",
					Err:    err,
				}
			}

			// Get affected rows count
			rowsAffected, _ := result.RowsAffected()

			return models.QueryResultMsg{
				Result: fmt.Sprintf("Query executed successfully. %d rows affected.", rowsAffected),
				Err:    nil,
			}
		}
	})
}
