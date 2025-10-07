package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancaldera/mirador/internal/models"
	"github.com/dancaldera/mirador/internal/styles"
)

// CalculateColumnWidths computes optimal column widths with improved distribution
func CalculateColumnWidths(columns []string, rows [][]string) []int {
	colWidths := make([]int, len(columns))

	// Track content type and lengths for better width allocation
	columnTypes := make([]string, len(columns))
	maxLengths := make([]int, len(columns))
	avgLengths := make([]float64, len(columns))

	// Initialize with header lengths (add space for sort indicators)
	for i, col := range columns {
		colWidths[i] = len(col) + 2 // Extra space for sort arrows
		maxLengths[i] = len(col)
	}

	// Analyze column content to determine optimal widths
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				cellLength := len(cell)

				// Infer column type for better width allocation
				if i < len(columnTypes) && columnTypes[i] == "" {
					if cellLength == 0 {
						columnTypes[i] = "empty"
					} else if IsNumeric(cell) {
						columnTypes[i] = "numeric"
					} else if IsDateLike(cell) {
						columnTypes[i] = "date"
					} else if IsBooleanLike(cell) {
						columnTypes[i] = "boolean"
					} else if cellLength > 50 {
						columnTypes[i] = "text"
					} else {
						columnTypes[i] = "string"
					}
				}

				// Track statistics
				if cellLength > maxLengths[i] {
					maxLengths[i] = cellLength
				}
				avgLengths[i] = (avgLengths[i] + float64(cellLength)) / 2
			}
		}
	}

	// Apply intelligent width allocation based on content type
	for i := range colWidths {
		contentType := columnTypes[i]
		maxLen := maxLengths[i]
		avgLen := int(avgLengths[i])

		switch contentType {
		case "boolean":
			colWidths[i] = Min(Max(8, len(columns[i])+2), 10)
		case "numeric":
			colWidths[i] = Min(Max(10, maxLen+1), 15)
		case "date":
			colWidths[i] = Min(Max(12, maxLen), 20)
		case "empty":
			colWidths[i] = Max(8, len(columns[i])+2)
		case "string":
			// Use average length with some padding, but cap reasonably
			target := Max(avgLen+3, len(columns[i])+2)
			colWidths[i] = Min(Max(target, 12), 35)
		case "text":
			// Long text gets more space but still capped
			target := Max(avgLen/2+10, len(columns[i])+2)
			colWidths[i] = Min(Max(target, 20), 45)
		default:
			// Fallback to original logic
			colWidths[i] = Min(Max(maxLen, len(columns[i])+2), 40)
		}

		// Ensure minimum and maximum bounds
		colWidths[i] = min(max(colWidths[i], 6), 60)
	}

	return colWidths
}

// IsNumeric checks if a string represents a number
func IsNumeric(s string) bool {
	if s == "" || s == "NULL" {
		return false
	}
	// Simple check for numeric content
	for _, char := range s {
		if !((char >= '0' && char <= '9') || char == '.' || char == '-' || char == '+' || char == 'e' || char == 'E') {
			return false
		}
	}
	return true
}

// IsDateLike checks if a string looks like a date/timestamp
func IsDateLike(s string) bool {
	if len(s) < 8 || s == "NULL" {
		return false
	}
	// Look for common date patterns
	return strings.Contains(s, "-") && (strings.Contains(s, ":") || len(s) >= 10)
}

// IsBooleanLike checks if a string represents a boolean
func IsBooleanLike(s string) bool {
	lower := strings.ToLower(s)
	return lower == "true" || lower == "false" || lower == "t" || lower == "f" ||
		lower == "yes" || lower == "no" || lower == "y" || lower == "n" ||
		lower == "1" || lower == "0"
}

// CreateVisibleColumnsAndRows handles horizontal scrolling for tables with enhanced UX
func CreateVisibleColumnsAndRows(columns []string, rows [][]string, scrollOffset, visibleCols int, colWidths []int, sortColumn string, sortDirection models.SortDirection) ([]table.Column, []table.Row) {
	if len(columns) == 0 || scrollOffset >= len(columns) {
		return []table.Column{}, []table.Row{}
	}

	// Calculate end column index
	endCol := scrollOffset + visibleCols
	if endCol > len(columns) {
		endCol = len(columns)
	}

	// Build visible columns with enhanced headers
	visibleColumns := columns[scrollOffset:endCol]
	cols := make([]table.Column, len(visibleColumns))
	for i, c := range visibleColumns {
		columnTitle := c

		// Add sorting indicators to column headers
		if c == sortColumn {
			switch sortDirection {
			case models.SortAsc:
				columnTitle = c + " ↑"
			case models.SortDesc:
				columnTitle = c + " ↓"
			}
		}

		cols[i] = table.Column{Title: columnTitle, Width: colWidths[scrollOffset+i]}
	}

	// Build visible rows with smarter content truncation
	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		visibleCells := make(table.Row, len(visibleColumns))
		for j := 0; j < len(visibleColumns); j++ {
			colIndex := scrollOffset + j
			if colIndex < len(r) {
				cell := r[colIndex]
				maxW := colWidths[colIndex]

				// Enhanced truncation logic for better readability
				if len(cell) > maxW {
					if maxW <= 8 {
						// Very narrow columns: show first few chars
						visibleCells[j] = cell[:Max(1, maxW-1)] + "…"
					} else if maxW <= 15 {
						// Narrow columns: smart truncation
						if len(cell) <= maxW+3 {
							visibleCells[j] = cell // Don't truncate if just slightly over
						} else {
							visibleCells[j] = cell[:maxW-2] + "…"
						}
					} else {
						// Wider columns: show more content with better ellipsis
						if strings.Contains(cell, " ") && len(cell) > maxW {
							// Try to break at word boundaries
							truncated := cell[:maxW-1]
							if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxW/2 {
								visibleCells[j] = truncated[:lastSpace] + "…"
							} else {
								visibleCells[j] = truncated + "…"
							}
						} else {
							visibleCells[j] = cell[:maxW-1] + "…"
						}
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
	availableWidth = max(availableWidth, 20)

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
	visibleCount = max(visibleCount, 0)

	// Create visible columns and rows with sorting indicators
	cols, rows := CreateVisibleColumnsAndRows(m.DataPreviewAllColumns, m.DataPreviewAllRows, startCol, visibleCount, colWidths, m.DataPreviewSortColumn, m.DataPreviewSortDirection)

	// Compute dynamic height to use remaining vertical space
	reserved := 10 // Title + info + help, approximate
	availableHeight := m.Height - v - reserved
	availableHeight = max(availableHeight, 5)

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

// UpdateRowDetailList creates field items for row detail view
func UpdateRowDetailList(columns []string, rowData []string) []list.Item {
	items := make([]list.Item, len(columns))
	for i, col := range columns {
		if i < len(rowData) {
			items[i] = models.FieldItem{
				Name:  col,
				Value: rowData[i],
			}
		} else {
			items[i] = models.FieldItem{
				Name:  col,
				Value: "",
			}
		}
	}
	return items
}

// SetErrorWithTimeout sets an error on the model that will auto-clear after the specified duration
func SetErrorWithTimeout(m models.Model, err error, duration time.Duration) (models.Model, tea.Cmd) {
	updatedModel := m
	updatedModel.Err = err
	if err != nil {
		timeout := time.Now().Add(duration)
		updatedModel.ErrorTimeout = &timeout
		return updatedModel, tea.Tick(duration, func(t time.Time) tea.Msg {
			return models.ErrorTimeoutMsg{}
		})
	} else {
		updatedModel.ErrorTimeout = nil
	}
	return updatedModel, nil
}

// ClearErrorTimeout clears error timeout and error if the timeout has expired
func ClearErrorTimeout(m models.Model) models.Model {
	updatedModel := m
	if m.ErrorTimeout != nil && time.Now().After(*m.ErrorTimeout) {
		updatedModel.Err = nil
		updatedModel.ErrorTimeout = nil
	}
	return updatedModel
}

// UpdateSavedConnectionsItems creates list items from saved connections
func UpdateSavedConnectionsItems(connections []models.SavedConnection) []list.Item {
	items := make([]list.Item, len(connections))
	for i, conn := range connections {
		connStr := conn.ConnectionStr
		if len(connStr) > 50 {
			connStr = connStr[:50] + "..."
		}
		items[i] = models.Item{
			ItemTitle: conn.Name,
			ItemDesc:  fmt.Sprintf("%s - %s", conn.Driver, connStr),
		}
	}
	return items
}

// CalculateListViewportHeight calculates the appropriate height for list components
// accounting for ViewBuilder-applied margins, title, status, and help text
func CalculateListViewportHeight(totalHeight int, hasTitle bool, hasStatus bool) int {
	_, v := styles.DocStyle.GetFrameSize()

	// Start with total height minus DocStyle frame
	availableHeight := totalHeight - v

	// Account for title (1-2 lines with margin)
	if hasTitle {
		availableHeight -= 3
	}

	// Account for status message (1-2 lines if present)
	if hasStatus {
		availableHeight -= 2
	}

	// Account for help text (1 line with margin)
	availableHeight -= 3

	// Account for additional spacing and margins
	availableHeight -= 2

	// Ensure minimum height
	if availableHeight < 5 {
		availableHeight = 5
	}

	return availableHeight
}
