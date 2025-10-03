package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dancaldera/dbx/internal/models"
	"github.com/dancaldera/dbx/internal/styles"
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

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("Enter") + ": execute query • " +
			styles.KeyStyle.Render("Tab") + ": switch focus • " +
			styles.KeyStyle.Render("↑/↓") + ": navigate results • " +
			styles.KeyStyle.Render("Ctrl+E") + ": export CSV • " +
			styles.KeyStyle.Render("Ctrl+J") + ": export JSON • " +
			styles.KeyStyle.Render("Esc") + ": back to tables",
	)

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
	content := styles.TitleStyle.Render(title)

	// Show status messages with improved styling
	if m.IsExporting {
		content += "\n" + styles.LoadingStyle.Render("⏳ Exporting data...")
	} else if m.Err != nil {
		content += "\n" + styles.ErrorStyle.Render("❌ Error: "+m.Err.Error())
	} else if m.QueryResult != "" {
		content += "\n" + styles.SuccessStyle.Render(m.QueryResult)
	}

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

		content += "\n" + styles.InfoStyle.Render(statusLine.String())

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

		content += "\n" + styles.SubtitleStyle.Render(columnLine.String())

		// Enhanced filter input with better styling
		if m.DataPreviewFilterActive {
			filterLabel := styles.SubtitleStyle.Render("🔍 Filter:")
			var filterField string
			if m.DataPreviewFilterInput.Focused() {
				filterField = styles.InputFocusedStyle.Render(m.DataPreviewFilterInput.View())
			} else {
				filterField = styles.InputStyle.Render(m.DataPreviewFilterInput.View())
			}
			content += "\n" + filterLabel + " " + filterField
		}

		// Enhanced sort mode indicator with column highlighting
		if m.DataPreviewSortMode {
			var sortModeInfo string
			if m.DataPreviewSortColumn != "" {
				currentDirection := "off"
				nextDirection := "ascending"
				switch m.DataPreviewSortDirection {
				case models.SortOff:
					currentDirection = "off"
					nextDirection = "ascending"
				case models.SortAsc:
					currentDirection = "ascending 🔼"
					nextDirection = "descending"
				case models.SortDesc:
					currentDirection = "descending 🔽"
					nextDirection = "off"
				}
				sortModeInfo = fmt.Sprintf("🎯 Sort Mode: '%s' (%s) → Press ENTER for %s",
					m.DataPreviewSortColumn, currentDirection, nextDirection)
			} else {
				sortModeInfo = "🎯 Sort Mode: Use ↑/↓ to select column, ENTER to sort"
			}
			content += "\n" + styles.WarningStyle.Render(sortModeInfo)
		}

		// Add visual separator before table
		content += "\n" + strings.Repeat("─", 80)
		content += "\n" + m.DataPreviewTable.View()

		// Add visual separator after table for better separation
		content += "\n" + strings.Repeat("─", 80)

	} else if m.Err == nil && m.QueryResult == "" && !m.IsExporting {
		content += "\n" + styles.InfoStyle.Render("📭 No data to display")
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
		// Group help text by function for better readability
		navigationHelp := styles.KeyStyle.Render("hjkl/↑↓←→") + ": navigate"
		actionHelp := styles.KeyStyle.Render("ENTER") + ": row details"
		pageHelp := styles.KeyStyle.Render("←→") + ": pages"
		featureHelp := styles.KeyStyle.Render("/") + ": filter • " + styles.KeyStyle.Render("s") + ": sort"
		utilityHelp := styles.KeyStyle.Render("ctrl+r") + ": reload • " + styles.KeyStyle.Render("ESC") + ": back"

		helpText = styles.HelpStyle.Render(
			navigationHelp + " • " + actionHelp + " • " + pageHelp + "\n" +
				featureHelp + " • " + utilityHelp)
	}

	content += "\n" + helpText
	return styles.DocStyle.Render(content)
}

// RowDetailView renders the detailed view of a selected row using a simple list
func RowDetailView(m models.Model) string {
	if m.IsViewingFieldDetail {
		// Show full field detail view with scrolling
		title := styles.TitleStyle.Render(fmt.Sprintf("Field: %s", m.SelectedFieldForDetail))

		// Find the selected field value
		var fieldValue string
		for i, col := range m.DataPreviewAllColumns {
			if col == m.SelectedFieldForDetail && i < len(m.SelectedRowData) {
				fieldValue = m.SelectedRowData[i]
				break
			}
		}

		// Show empty string as-is when value is empty

		// Try to format JSON for better readability
		if strings.HasPrefix(strings.TrimSpace(fieldValue), "{") || strings.HasPrefix(strings.TrimSpace(fieldValue), "[") {
			// Attempt to pretty-print JSON
			var formatted strings.Builder
			indent := 0
			inString := false
			escaped := false

			for i, char := range fieldValue {
				if escaped {
					formatted.WriteRune(char)
					escaped = false
					continue
				}

				if char == '\\' && inString {
					formatted.WriteRune(char)
					escaped = true
					continue
				}

				if char == '"' {
					inString = !inString
					formatted.WriteRune(char)
					continue
				}

				if inString {
					formatted.WriteRune(char)
					continue
				}

				switch char {
				case '{', '[':
					formatted.WriteRune(char)
					formatted.WriteRune('\n')
					indent++
					for j := 0; j < indent*2; j++ {
						formatted.WriteRune(' ')
					}
				case '}', ']':
					if i > 0 && fieldValue[i-1] != '\n' {
						formatted.WriteRune('\n')
					}
					indent--
					for j := 0; j < indent*2; j++ {
						formatted.WriteRune(' ')
					}
					formatted.WriteRune(char)
					if i < len(fieldValue)-1 {
						formatted.WriteRune('\n')
						for j := 0; j < indent*2; j++ {
							formatted.WriteRune(' ')
						}
					}
				case ',':
					formatted.WriteRune(char)
					formatted.WriteRune('\n')
					for j := 0; j < indent*2; j++ {
						formatted.WriteRune(' ')
					}
				case ':':
					formatted.WriteRune(char)
					formatted.WriteRune(' ')
				default:
					if char != ' ' || formatted.Len() == 0 || formatted.String()[formatted.Len()-1] != ' ' {
						formatted.WriteRune(char)
					}
				}
			}
			fieldValue = formatted.String()
		}

		// Split content into lines for scrolling
		lines := strings.Split(fieldValue, "\n")

		// Calculate dynamic height (use window height minus padding for title and help text)
		availableHeight := max(m.Height-10, 5) // Reserve space for title, help text, and margins

		// Calculate visible range
		startLine := m.FieldDetailScrollOffset
		endLine := min(startLine+availableHeight, len(lines))

		// Calculate dynamic width (use window width minus padding)
		availableWidth := min(max(m.Width-10, 40), 200) // Reserve space for margins

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

		titleWithScroll := title
		if scrollInfo != "" {
			titleWithScroll += styles.InfoStyle.Render(scrollInfo)
		}

		// Render with dynamic dimensions
		contentBox := styles.InputStyle.Width(availableWidth).Height(availableHeight).Render(displayContent)

		helpText := styles.HelpStyle.Render(
			styles.KeyStyle.Render("↑↓/jk") + ": scroll vertical • " +
				styles.KeyStyle.Render("←→/hl") + ": scroll horizontal • " +
				styles.KeyStyle.Render("esc") + ": back to field list",
		)

		content := titleWithScroll + "\n\n" + contentBox + "\n\n" + helpText

		return styles.DocStyle.Render(content)
	}

	// Show field list view or edit mode
	if m.IsEditingField {
		// Show simplified field editing interface
		title := fmt.Sprintf("Edit Field: %s", m.EditingFieldName)
		content := styles.TitleStyle.Render(title) + "\n\n"

		// Show status messages
		if m.Err != nil {
			content += styles.ErrorStyle.Render("❌ "+m.Err.Error()) + "\n\n"
		} else if m.QueryResult != "" {
			content += styles.SuccessStyle.Render(m.QueryResult) + "\n\n"
		}

		// Only show the textarea for editing
		content += m.FieldTextarea.View() + "\n\n"

		helpText := styles.HelpStyle.Render(
			styles.KeyStyle.Render("Ctrl+S") + ": save changes • " +
				styles.KeyStyle.Render("Ctrl+K") + ": clear • " +
				styles.KeyStyle.Render("Esc") + ": cancel",
		)
		content += helpText

		return styles.DocStyle.Render(content)
	}

	title := fmt.Sprintf("Row Details - %s", m.SelectedTable)
	content := styles.TitleStyle.Render(title) + "\n"

	if len(m.SelectedRowData) == 0 || len(m.DataPreviewAllColumns) == 0 {
		content += styles.ErrorStyle.Render("❌ No row data available") + "\n\n"
		helpText := styles.HelpStyle.Render(styles.KeyStyle.Render("esc") + ": back to table")
		content += helpText
		return styles.DocStyle.Render(content)
	}

	// Show status messages
	if m.Err != nil {
		content += styles.ErrorStyle.Render("❌ "+m.Err.Error()) + "\n\n"
	} else if m.QueryResult != "" {
		content += styles.SuccessStyle.Render(m.QueryResult) + "\n\n"
	}

	// Show the list of fields
	content += m.RowDetailList.View()

	// Add help text
	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("↑↓") + ": navigate fields • " +
			styles.KeyStyle.Render("enter") + ": view field detail • " +
			styles.KeyStyle.Render("e") + ": edit field • " +
			styles.KeyStyle.Render("esc") + ": back to table",
	)
	content += "\n" + helpText

	return styles.DocStyle.Render(content)
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
