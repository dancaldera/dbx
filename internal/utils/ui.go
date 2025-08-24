package utils

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/dancaldera/dbx/internal/models"
	"github.com/dancaldera/dbx/internal/styles"
)

// CalculateColumnWidths computes optimal column widths based on content
func CalculateColumnWidths(columns []string, rows [][]string) []int {
	colWidths := make([]int, len(columns))

	// Initialize with header lengths
	for i, col := range columns {
		colWidths[i] = len(col)
	}

	// Single pass through all data
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				cellLength := len(cell)
				if cellLength > 40 {
					cellLength = 40 // Cap at maximum display width
				}
				if cellLength > colWidths[i] {
					colWidths[i] = cellLength
				}
			}
		}
	}

	// Apply minimum width constraints
	for i := range colWidths {
		if colWidths[i] < 8 {
			colWidths[i] = 8
		} else if colWidths[i] > 40 {
			colWidths[i] = 40
		}
	}

	return colWidths
}

// CreateVisibleColumnsAndRows handles horizontal scrolling for tables
func CreateVisibleColumnsAndRows(columns []string, rows [][]string, scrollOffset, visibleCols int, colWidths []int) ([]table.Column, []table.Row) {
	if len(columns) == 0 || scrollOffset >= len(columns) {
		return []table.Column{}, []table.Row{}
	}

	// Calculate end column index
	endCol := scrollOffset + visibleCols
	if endCol > len(columns) {
		endCol = len(columns)
	}

	// Build visible columns
	visibleColumns := columns[scrollOffset:endCol]
	cols := make([]table.Column, len(visibleColumns))
	for i, c := range visibleColumns {
		cols[i] = table.Column{Title: c, Width: colWidths[scrollOffset+i]}
	}

	// Build visible rows with content truncation per computed column width
	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		visibleCells := make(table.Row, len(visibleColumns))
		for j := 0; j < len(visibleColumns); j++ {
			colIndex := scrollOffset + j
			if colIndex < len(r) {
				cell := r[colIndex]
				maxW := colWidths[colIndex]
				if len(cell) > maxW {
					ellipsis := 0
					if maxW >= 3 {
						ellipsis = 3
					}
					trim := Max(0, maxW-ellipsis)
					if trim <= 0 {
						visibleCells[j] = ""
					} else if ellipsis > 0 {
						visibleCells[j] = cell[:trim] + "..."
					} else {
						visibleCells[j] = cell[:trim]
					}
				} else {
					visibleCells[j] = cell
				}
			} else {
				visibleCells[j] = ""
			}
		}
		tableRows[i] = visibleCells
	}

	return cols, tableRows
}

// CreateDataPreviewTable builds a data preview table with horizontal scrolling support
func CreateDataPreviewTable(m models.Model) models.Model {
	if len(m.DataPreviewAllColumns) == 0 {
		return m
	}

	// Determine available width for table content within the document frame
	h, v := styles.DocStyle.GetFrameSize()
	availableWidth := m.Width - h - 4
	if availableWidth < 20 {
		availableWidth = 20
	}

	// Calculate column widths
	colWidths := CalculateColumnWidths(m.DataPreviewAllColumns, m.DataPreviewAllRows)

	// Compute how many columns fit starting from the current scroll offset
	startCol := m.DataPreviewScrollOffset
	sum := 0
	endCol := startCol
	for endCol < len(colWidths) {
		// Rough allowance for padding/separators per column
		next := colWidths[endCol] + 3
		if sum+next > availableWidth {
			break
		}
		sum += next
		endCol++
	}
	if endCol == startCol {
		// Ensure at least one column is visible
		endCol = Min(startCol+1, len(colWidths))
	}
	visibleCount := endCol - startCol
	if visibleCount < 0 {
		visibleCount = 0
	}

	// Create visible columns and rows
	cols, rows := CreateVisibleColumnsAndRows(m.DataPreviewAllColumns, m.DataPreviewAllRows, startCol, visibleCount, colWidths)

	// Compute dynamic height to use remaining vertical space
	reserved := 10 // Title + info + help, approximate
	availableHeight := m.Height - v - reserved
	if availableHeight < 5 {
		availableHeight = 5
	}

	// Create updated model with new table
	updatedModel := m
	updatedModel.DataPreviewVisibleCols = visibleCount
	updatedModel.DataPreviewTable = table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(availableHeight),
	)
	updatedModel.DataPreviewTable.SetStyles(styles.GetBlueTableStyles())

	return updatedModel
}

// UpdateSavedConnectionsList refreshes the saved connections list items
func UpdateSavedConnectionsList(m models.Model) models.Model {
	savedItems := UpdateSavedConnectionsItems(m.SavedConnections)
	updatedModel := m
	updatedModel.SavedConnectionsList.SetItems(savedItems)
	return updatedModel
}
