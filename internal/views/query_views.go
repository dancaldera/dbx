package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dancaldera/dbx/internal/models"
	"github.com/dancaldera/dbx/internal/styles"
	"github.com/dancaldera/dbx/internal/utils"
)

// QueryView renders the SQL query execution screen
func QueryView(m models.Model) string {
	builder := NewViewBuilder().WithTitle("⚡  SQL Query Runner")

	// Add status messages
	if m.IsExecutingQuery {
		builder.WithStatus("⏳ Executing query...", StatusLoading)
	} else if m.IsExporting {
		builder.WithStatus("⏳ Exporting data...", StatusLoading)
	} else if m.Err != nil {
		builder.WithStatus("❌ "+m.Err.Error(), StatusError)
	}

	// Query input field
	queryField := RenderInputField("💻 Enter SQL Query:", m.QueryInput.View(), m.QueryInput.Focused())

	// Assemble content elements
	var contentElements []string
	contentElements = append(contentElements, queryField)

	// Add query results if present
	if m.QueryResult != "" {
		resultLabel := RenderSectionTitle("Query Result:")
		resultText := styles.SuccessStyle.Render(m.QueryResult)

		// Only show the table if it has both columns and rows
		if len(m.QueryResultsTable.Columns()) > 0 && len(m.QueryResultsTable.Rows()) > 0 {
			tableContent := styles.CardStyle.Render(m.QueryResultsTable.View())
			resultContent := lipgloss.JoinVertical(lipgloss.Left, resultLabel, resultText, tableContent)
			contentElements = append(contentElements, resultContent)
		} else {
			resultContent := lipgloss.JoinVertical(lipgloss.Left, resultLabel, resultText)
			contentElements = append(contentElements, resultContent)
		}
	}

	// Examples box
	examples := RenderInfoBox(
		styles.SubtitleStyle.Render("💡 Examples:") + "\n" +
			styles.KeyStyle.Render("SELECT") + " * FROM users LIMIT 10;\n" +
			styles.KeyStyle.Render("INSERT") + " INTO users (name, email) VALUES ('John', 'john@example.com');\n" +
			styles.KeyStyle.Render("UPDATE") + " users SET email = 'new@example.com' WHERE id = 1;\n" +
			styles.KeyStyle.Render("DELETE") + " FROM users WHERE id = 1;",
	)
	contentElements = append(contentElements, examples)

	baseHelp := styles.KeyStyle.Render("?") + ": help • " +
		styles.KeyStyle.Render("Enter") + ": execute • " +
		styles.KeyStyle.Render("Tab") + ": switch focus • " +
		styles.KeyStyle.Render("Esc") + ": back"

	fullHelp := styles.KeyStyle.Render("Enter") + ": execute query • " +
		styles.KeyStyle.Render("Tab") + ": switch focus • " +
		styles.KeyStyle.Render("↑/↓") + ": navigate results • " +
		styles.KeyStyle.Render("Ctrl+E") + ": export CSV • " +
		styles.KeyStyle.Render("Ctrl+J") + ": export JSON • " +
		styles.KeyStyle.Render("Esc") + ": back to tables • " +
		styles.KeyStyle.Render("?") + ": hide help"

	helpText := RenderContextualHelp(baseHelp, fullHelp, m.ShowFullHelp)

	return builder.
		WithContent(contentElements...).
		WithHelp(helpText).
		Render()
}

// QueryHistoryView renders the query history screen
func QueryHistoryView(m models.Model) string {
	builder := NewViewBuilder()

	if len(m.QueryHistory) == 0 {
		emptyState := RenderEmptyState("📝", "No query history yet.\n\nExecute some queries to see them here!")
		builder.WithContent(m.QueryHistoryList.View(), emptyState)
	} else {
		builder.WithContent(m.QueryHistoryList.View())
	}

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": use query • " +
			styles.KeyStyle.Render("d") + ": delete • " +
			styles.KeyStyle.Render("esc") + ": back",
	)

	return builder.WithHelp(helpText).Render()
}

// DataPreviewView renders the enhanced table data preview screen
func DataPreviewView(m models.Model) string {
	// Enhanced title with table name
	title := fmt.Sprintf("%s", m.SelectedTable)
	builder := NewViewBuilder().WithTitle(title)

	// Show status messages with improved styling
	if m.IsExporting {
		builder.WithStatus("⏳ Exporting data...", StatusLoading)
	} else if m.Err != nil {
		builder.WithStatus("❌ Error: "+m.Err.Error(), StatusError)
	} else if m.QueryResult != "" {
		builder.WithStatus(m.QueryResult, StatusSuccess)
	}

	// Build content sections
	var contentElements []string

	// Only show the table if it has both columns and rows
	if len(m.DataPreviewTable.Columns()) > 0 && len(m.DataPreviewTable.Rows()) > 0 {
		// Calculate pagination info with better formatting
		totalPages := (m.DataPreviewTotalRows + m.DataPreviewItemsPerPage - 1) / m.DataPreviewItemsPerPage
		if totalPages == 0 {
			totalPages = 1
		}
		currentPage := m.DataPreviewCurrentPage + 1

		// Calculate current row range
		startRow := (m.DataPreviewCurrentPage * m.DataPreviewItemsPerPage) + 1
		endRow := startRow + len(m.DataPreviewTable.Rows()) - 1

		// Enhanced table metadata with better visual hierarchy
		var statusLine strings.Builder

		// Table name and filter status
		if m.DataPreviewFilterValue != "" {
			statusLine.WriteString(fmt.Sprintf("📋 %s (filtered: '%s')", m.SelectedTable, m.DataPreviewFilterValue))
		} else {
			statusLine.WriteString(fmt.Sprintf("📋 %s", m.SelectedTable))
		}

		// Row information with visual indicators
		statusLine.WriteString(fmt.Sprintf(" • 📄 Rows %d-%d of %d", startRow, endRow, m.DataPreviewTotalRows))

		// Page navigation with arrows
		if totalPages > 1 {
			var pageIndicator string
			if currentPage > 1 {
				pageIndicator += "← "
			}
			pageIndicator += fmt.Sprintf("Page %d/%d", currentPage, totalPages)
			if currentPage < totalPages {
				pageIndicator += " →"
			}
			statusLine.WriteString(" • " + pageIndicator)
		}

		// Sort status with enhanced indicators
		if m.DataPreviewSortColumn != "" {
			var sortIcon string
			switch m.DataPreviewSortDirection {
			case models.SortAsc:
				sortIcon = "🔼"
			case models.SortDesc:
				sortIcon = "🔽"
			}
			statusLine.WriteString(fmt.Sprintf(" • %s %s", sortIcon, m.DataPreviewSortColumn))
		}

		contentElements = append(contentElements, styles.InfoStyle.Render(statusLine.String()))

		// Enhanced column scroll indicator with scroll arrows
		totalCols := len(m.DataPreviewAllColumns)
		startCol := m.DataPreviewScrollOffset + 1
		endCol := m.DataPreviewScrollOffset + m.DataPreviewVisibleCols
		if endCol > totalCols {
			endCol = totalCols
		}

		var columnLine strings.Builder
		if m.DataPreviewScrollOffset > 0 {
			columnLine.WriteString("← ")
		}
		columnLine.WriteString(fmt.Sprintf("Columns %d-%d of %d", startCol, endCol, totalCols))
		if endCol < totalCols {
			columnLine.WriteString(" →")
		}

		visibleRows := len(m.DataPreviewTable.Rows())
		columnLine.WriteString(fmt.Sprintf(" • %d rows visible", visibleRows))

		contentElements = append(contentElements, styles.SubtitleStyle.Render(columnLine.String()))

		// Enhanced filter input with better styling
		if m.DataPreviewFilterActive {
			filterLabel := styles.SubtitleStyle.Render("🔍 Filter:")
			var filterField string
			if m.DataPreviewFilterInput.Focused() {
				filterField = styles.InputFocusedStyle.Render(m.DataPreviewFilterInput.View())
			} else {
				filterField = styles.InputStyle.Render(m.DataPreviewFilterInput.View())
			}
			contentElements = append(contentElements, filterLabel+" "+filterField)
		}

		// Enhanced sort mode indicator with clear navigation and state messaging
		if m.DataPreviewSortMode {
			var sortModeInfo string
			if m.DataPreviewSortColumn != "" {
				// A column is selected - show its current state and next action
				switch m.DataPreviewSortDirection {
				case models.SortOff:
					// Column selected but not sorted yet
					sortModeInfo = fmt.Sprintf("🎯 Sort Mode: '%s' → Press ENTER to sort ascending (↑/↓ to change column)",
						m.DataPreviewSortColumn)
				case models.SortAsc:
					// Currently sorted ascending
					sortModeInfo = fmt.Sprintf("🎯 Sort Mode: '%s' 🔼 ascending → Press ENTER for descending (↑/↓ to change column)",
						m.DataPreviewSortColumn)
				case models.SortDesc:
					// Currently sorted descending
					sortModeInfo = fmt.Sprintf("🎯 Sort Mode: '%s' 🔽 descending → Press ENTER to clear sort (↑/↓ to change column)",
						m.DataPreviewSortColumn)
				}
			} else {
				// No column selected yet - emphasize navigation
				sortModeInfo = "🎯 Sort Mode: Use ↑/↓ to select column, then ENTER to sort"
			}
			contentElements = append(contentElements, styles.WarningStyle.Render(sortModeInfo))
		}

		// Add visual separator before table
		contentElements = append(contentElements, strings.Repeat("─", 80))
		contentElements = append(contentElements, m.DataPreviewTable.View())

		// Add visual separator after table for better separation
		contentElements = append(contentElements, strings.Repeat("─", 80))

	} else if m.Err == nil && m.QueryResult == "" && !m.IsExporting {
		contentElements = append(contentElements, styles.InfoStyle.Render("📭 No data to display"))
	}

	// Enhanced help text with better grouping and visual hierarchy
	var helpText string
	if m.DataPreviewFilterActive {
		helpText = styles.HelpStyle.Render(
			styles.KeyStyle.Render("ENTER") + ": apply filter • " +
				styles.KeyStyle.Render("ESC") + ": cancel filter")
	} else if m.DataPreviewSortMode {
		helpText = styles.HelpStyle.Render(
			styles.KeyStyle.Render("↑↓") + ": select column • " +
				styles.KeyStyle.Render("ENTER") + ": cycle sort (off→asc→desc) • " +
				styles.KeyStyle.Render("ESC") + ": exit sort")
	} else {
		// Compact help for normal mode
		baseHelp := styles.KeyStyle.Render("?") + ": help • " +
			styles.KeyStyle.Render("↑↓←→") + ": navigate • " +
			styles.KeyStyle.Render("ENTER") + ": details • " +
			styles.KeyStyle.Render("/") + ": filter • " +
			styles.KeyStyle.Render("s") + ": sort • " +
			styles.KeyStyle.Render("ESC") + ": back"

		// Full help with all options
		fullHelp := styles.KeyStyle.Render("hjkl/↑↓←→") + ": navigate • " +
			styles.KeyStyle.Render("ENTER") + ": row details • " +
			styles.KeyStyle.Render("←→") + ": pages • " +
			styles.KeyStyle.Render("/") + ": filter • " +
			styles.KeyStyle.Render("s") + ": sort • " +
			styles.KeyStyle.Render("ctrl+r") + ": reload • " +
			styles.KeyStyle.Render("ESC") + ": back • " +
			styles.KeyStyle.Render("?") + ": hide help"

		helpText = RenderContextualHelp(baseHelp, fullHelp, m.ShowFullHelp)
	}

	return builder.WithContent(contentElements...).WithHelp(helpText).Render()
}

// RowDetailView renders the detailed view of a selected row using a simple list
func RowDetailView(m models.Model) string {
	if m.IsViewingFieldDetail {
		// Show full field detail view with scrolling
		title := fmt.Sprintf("Field: %s", m.SelectedFieldForDetail)

		// Find the selected field value
		var fieldValue string
		for i, col := range m.DataPreviewAllColumns {
			if col == m.SelectedFieldForDetail && i < len(m.SelectedRowData) {
				fieldValue = m.SelectedRowData[i]
				break
			}
		}

		// Format field value (handles JSON pretty-printing)
		fieldValue = utils.FormatFieldValue(fieldValue)

		// Split content into lines for scrolling
		lines := strings.Split(fieldValue, "\n")

		// Calculate dynamic height accounting for ViewBuilder elements
		// Title (2-3 lines), status (1-2 lines), help (1 line), margins
		h, v := styles.DocStyle.GetFrameSize()
		availableHeight := m.Height - v - 12 // Account for all UI elements
		if availableHeight < 5 {
			availableHeight = 5
		}

		// Calculate visible range
		startLine := m.FieldDetailScrollOffset
		endLine := min(startLine+availableHeight, len(lines))

		// Calculate dynamic width (use window width minus padding)
		availableWidth := m.Width - h - 8 // Account for frame and padding
		if availableWidth < 40 {
			availableWidth = 40
		}
		if availableWidth > 200 {
			availableWidth = 200
		}

		// Build visible content with horizontal scrolling
		var visibleLines []string
		for i := startLine; i < endLine; i++ {
			line := lines[i]
			// Apply horizontal scrolling
			if m.FieldDetailHorizontalOffset < len(line) {
				endChar := min(m.FieldDetailHorizontalOffset+availableWidth, len(line))
				line = line[m.FieldDetailHorizontalOffset:endChar]
			} else {
				line = ""
			}
			visibleLines = append(visibleLines, line)
		}

		// Join the visible lines
		displayContent := strings.Join(visibleLines, "\n")

		// Create scroll indicators
		scrollInfo := ""

		// Show line information
		startDisplayLine := m.FieldDetailScrollOffset + 1
		endDisplayLine := min(m.FieldDetailScrollOffset+len(visibleLines), len(lines))

		if len(lines) > 1 {
			scrollInfo = fmt.Sprintf(" • Lines %d-%d of %d", startDisplayLine, endDisplayLine, len(lines))
		}

		if m.FieldDetailHorizontalOffset > 0 {
			scrollInfo += fmt.Sprintf(" • Column offset: %d", m.FieldDetailHorizontalOffset)
		}

		// Build with ViewBuilder
		builder := NewViewBuilder().WithTitle(title)

		if scrollInfo != "" {
			builder.WithStatus(scrollInfo, StatusInfo)
		}

		// Render with dynamic dimensions
		contentBox := styles.InputStyle.Width(availableWidth).Height(availableHeight).Render(displayContent)

		helpText := styles.HelpStyle.Render(
			styles.KeyStyle.Render("↑↓/jk") + ": scroll vertical • " +
				styles.KeyStyle.Render("←→/hl") + ": scroll horizontal • " +
				styles.KeyStyle.Render("esc") + ": back to field list",
		)

		return builder.WithContent(contentBox).WithHelp(helpText).Render()
	}

	// Show field list view or edit mode
	if m.IsEditingField {
		// Show simplified field editing interface
		title := fmt.Sprintf("Edit Field: %s", m.EditingFieldName)
		builder := NewViewBuilder().WithTitle(title)

		// Show status messages
		if m.Err != nil {
			builder.WithStatus("❌ "+m.Err.Error(), StatusError)
		} else if m.QueryResult != "" {
			builder.WithStatus(m.QueryResult, StatusSuccess)
		}

		// Help text
		helpText := styles.HelpStyle.Render(
			styles.KeyStyle.Render("Ctrl+S") + ": save changes • " +
				styles.KeyStyle.Render("Ctrl+K") + ": clear • " +
				styles.KeyStyle.Render("Esc") + ": cancel",
		)

		return builder.WithContent(m.FieldTextarea.View()).WithHelp(helpText).Render()
	}

	// Default view: field list
	fieldCount := len(m.DataPreviewAllColumns)
	title := fmt.Sprintf("Row Details - %s (%d fields)", m.SelectedTable, fieldCount)
	builder := NewViewBuilder().WithTitle(title)

	if len(m.SelectedRowData) == 0 || len(m.DataPreviewAllColumns) == 0 {
		builder.WithStatus("❌ No row data available", StatusError)
		helpText := styles.HelpStyle.Render(styles.KeyStyle.Render("esc") + ": back to table")
		return builder.WithHelp(helpText).Render()
	}

	// Show status messages
	if m.Err != nil {
		builder.WithStatus("❌ "+m.Err.Error(), StatusError)
	} else if m.QueryResult != "" {
		builder.WithStatus(m.QueryResult, StatusSuccess)
	}

	// Add help text
	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("↑↓") + ": navigate fields • " +
			styles.KeyStyle.Render("enter") + ": view field detail • " +
			styles.KeyStyle.Render("e") + ": edit field • " +
			styles.KeyStyle.Render("esc") + ": back to table",
	)

	return builder.WithContent(m.RowDetailList.View()).WithHelp(helpText).Render()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
