package utils

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancaldera/dbx/internal/models"
	"github.com/dancaldera/dbx/internal/styles"
)

// DetermineSortParameters converts internal sort parameters to SQL format
func DetermineSortParameters(sortDirection models.SortDirection, sortColumn string) (string, string) {
	if sortDirection == models.SortOff || sortColumn == "" {
		return "", ""
	}

	switch sortDirection {
	case models.SortAsc:
		return sortColumn, "ASC"
	case models.SortDesc:
		return sortColumn, "DESC"
	default:
		return "", ""
	}
}

// CreateTableInfos generates TableInfo objects from table names
func CreateTableInfos(tables []string, schema string) []models.TableInfo {
	tableInfos := make([]models.TableInfo, len(tables))
	for i, tableName := range tables {
		tableInfos[i] = models.TableInfo{
			Name:        tableName,
			Schema:      schema,
			TableType:   "BASE TABLE",
			Description: "Table",
		}
	}
	return tableInfos
}

// UpdateRowDetailList creates list items for row detail view
func UpdateRowDetailList(columns []string, rowData []string) []list.Item {
	items := make([]list.Item, len(columns))
	for i, col := range columns {
		var value string
		if i < len(rowData) {
			value = rowData[i]
		} else {
			value = "NULL"
		}
		items[i] = models.FieldItem{Name: col, Value: value}
	}
	return items
}

// UpdateSavedConnectionsItems creates list items for saved connections
func UpdateSavedConnectionsItems(savedConnections []models.SavedConnection) []list.Item {
	savedItems := make([]list.Item, len(savedConnections))
	for i, conn := range savedConnections {
		connStr := conn.ConnectionStr
		if len(connStr) > 50 {
			connStr = connStr[:50] + "..."
		}
		savedItems[i] = models.Item{
			ItemTitle: conn.Name,
			ItemDesc:  fmt.Sprintf("%s - %s", conn.Driver, connStr),
		}
	}
	return savedItems
}

// CreateTableListItems generates list items from table infos for display
func CreateTableListItems(tableInfos []models.TableInfo) []list.Item {
	items := make([]list.Item, len(tableInfos))
	for i, info := range tableInfos {
		items[i] = models.Item{
			ItemTitle: info.Name,
			// omit description to avoid redundant "Table" line per item
		}
	}
	return items
}

// HandleTestConnectionResult processes test connection result
func HandleTestConnectionResult(m models.Model, msg models.TestConnectionResult) (models.Model, tea.Cmd) {
	updatedModel := m
	updatedModel.IsTestingConnection = false
	if msg.Success {
		updatedModel.QueryResult = "✅ Connection successful!"
	} else {
		updatedModel.Err = fmt.Errorf("connection failed: %v", msg.Err)
	}

	return updatedModel, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return models.ClearResultMsg{}
	})
}

// HandleColumnsResult processes columns result and updates model
func HandleColumnsResult(m models.Model, msg models.ColumnsResult) models.Model {
	updatedModel := m
	updatedModel.IsLoadingColumns = false

	if msg.Err != nil {
		updatedModel.Err = msg.Err
		return updatedModel
	}

	// Convert to table rows
	rows := make([]table.Row, len(msg.Columns))
	for i, col := range msg.Columns {
		rows[i] = table.Row{col[0], col[1], col[2], col[3]}
	}

	updatedModel.ColumnsTable.SetRows(rows)
	updatedModel.State = models.ColumnsView

	return updatedModel
}

// HandleDataPreviewResult processes data preview result
func HandleDataPreviewResult(m models.Model, msg models.DataPreviewResult) models.Model {
	updatedModel := m
	updatedModel.IsLoadingPreview = false
	if msg.Err != nil {
		updatedModel.Err = msg.Err
		return updatedModel
	}

	// Update total rows count
	if msg.TotalRows > 0 {
		updatedModel.DataPreviewTotalRows = msg.TotalRows
	}

	// Store all columns and rows for horizontal scrolling
	updatedModel.DataPreviewAllColumns = msg.Columns
	updatedModel.DataPreviewAllRows = msg.Rows
	updatedModel.DataPreviewScrollOffset = 0 // Reset scroll position

	// Create the initial table view
	updatedModel = CreateDataPreviewTable(updatedModel)
	updatedModel.State = models.DataPreviewView
	return updatedModel
}

// HandleRelationshipsResult processes relationships result
func HandleRelationshipsResult(m models.Model, msg models.RelationshipsResult) models.Model {
	updatedModel := m
	if msg.Err != nil {
		updatedModel.Err = msg.Err
		return updatedModel
	}
	// columns: From Table, From Column, To Table, To Column, Constraint Name
	cols := []table.Column{
		{Title: "From Table", Width: 20},
		{Title: "From Column", Width: 20},
		{Title: "To Table", Width: 20},
		{Title: "To Column", Width: 20},
		{Title: "Constraint Name", Width: 25},
	}
	rows := make([]table.Row, len(msg.Relationships))
	for i, rel := range msg.Relationships {
		rows[i] = table.Row(rel)
	}
	updatedModel.RelationshipsTable = table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	updatedModel.RelationshipsTable.SetStyles(styles.GetBlueTableStyles())
	updatedModel.State = models.RelationshipsView
	return updatedModel
}

// HandleFieldUpdateResult processes field update result
func HandleFieldUpdateResult(m models.Model, msg models.FieldUpdateResult) (models.Model, tea.Cmd) {
	updatedModel := m
	if msg.ExitEdit {
		updatedModel.IsEditingField = false
		updatedModel.FieldTextarea.Blur()
	}

	if msg.Success {
		// Update the row data with the new value
		if updatedModel.EditingFieldIndex >= 0 && updatedModel.EditingFieldIndex < len(updatedModel.SelectedRowData) {
			updatedModel.SelectedRowData[updatedModel.EditingFieldIndex] = msg.NewValue
		}

		// Update the row detail list with the new value
		items := UpdateRowDetailList(updatedModel.DataPreviewAllColumns, updatedModel.SelectedRowData)
		updatedModel.RowDetailList.SetItems(items)

		updatedModel.QueryResult = "✅ Field updated successfully!"
		updatedModel.Err = nil

		// Clear the editing state
		updatedModel.EditingFieldName = ""
		updatedModel.OriginalFieldValue = ""

		return updatedModel, ClearResultAfterTimeout()
	} else {
		updatedModel.Err = msg.Err
		return updatedModel, nil
	}
}
