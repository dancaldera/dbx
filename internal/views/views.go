package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dancaldera/dbx/internal/models"
	"github.com/dancaldera/dbx/internal/styles"
)

// DBTypeView renders the database type selection screen
func DBTypeView(m models.Model) string {
	content := m.DBTypeList.View()

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": select • " +
			styles.KeyStyle.Render("s") + ": saved connections • " +
			styles.KeyStyle.Render("q") + ": quit",
	)

	return styles.DocStyle.Render(content + "\n" + helpText)
}

// SavedConnectionsView renders the saved connections screen
func SavedConnectionsView(m models.Model) string {
	// Create the main content area
	listContent := m.SavedConnectionsList.View()
	
	// Build layout with elements
	var elements []string
	
	// Build title with inline status/error
	var titleLine string
	baseTitle := "📋 Saved Connections"
	
	if m.IsConnecting {
		// Show loading status inline with title using horizontal join
		titlePart := styles.TitleStyle.Render(baseTitle)
		if selectedItem, ok := m.SavedConnectionsList.SelectedItem().(models.Item); ok {
			loadingText := fmt.Sprintf("⏳ Connecting to %s...", selectedItem.ItemTitle)
			loadingPart := styles.LoadingStyle.Render(loadingText)
			titleLine = lipgloss.JoinHorizontal(lipgloss.Left, titlePart, "  ", loadingPart)
		} else {
			loadingText := "⏳ Connecting..."
			loadingPart := styles.LoadingStyle.Render(loadingText)
			titleLine = lipgloss.JoinHorizontal(lipgloss.Left, titlePart, "  ", loadingPart)
		}
	} else if m.Err != nil {
		// Show error inline with title - clean and seamless
		titlePart := styles.TitleStyle.Render(baseTitle)
		errorText := fmt.Sprintf("🚨 %s", m.Err.Error())
		errorPart := styles.ErrorStyle.Render(errorText)
		titleLine = lipgloss.JoinHorizontal(lipgloss.Left, titlePart, "  ", errorPart)
	} else if m.QueryResult != "" {
		// Show success message inline with title
		titlePart := styles.TitleStyle.Render(baseTitle)
		successText := m.QueryResult
		successPart := styles.SuccessStyle.Render(successText)
		titleLine = lipgloss.JoinHorizontal(lipgloss.Left, titlePart, "  ", successPart)
	} else {
		// Just the title
		titleLine = styles.TitleStyle.Render(baseTitle)
	}
	
	elements = append(elements, titleLine)
	
	// Add spacing before the list
	elements = append(elements, "")
	
	// Add the list
	elements = append(elements, listContent)
	
	// Add empty state message if needed (but don't show list)
	if len(m.SavedConnections) == 0 && !m.IsConnecting && m.Err == nil && m.QueryResult == "" {
		elements = []string{titleLine, "", styles.InfoStyle.Render("📝 No saved connections yet.\n\nGo back and create your first connection!")}
	}

	// Join all elements
	content := lipgloss.JoinVertical(lipgloss.Left, elements...)

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": connect • " +
			styles.KeyStyle.Render("d") + ": delete • " +
			styles.KeyStyle.Render("esc") + ": back",
	)

	return styles.DocStyle.Render(content + "\n" + helpText)
}

// ConnectionView renders the database connection configuration screen
func ConnectionView(m models.Model) string {
	var dbIcon string
	switch m.SelectedDB.Driver {
	case "postgres":
		dbIcon = "🐘"
	case "mysql":
		dbIcon = "🐬"
	case "sqlite3":
		dbIcon = "📁"
	default:
		dbIcon = "🗄️"
	}

	title := styles.TitleStyle.Render(fmt.Sprintf("%s  Connect to %s", dbIcon, m.SelectedDB.Name))

	var messageContent string
	if m.IsTestingConnection {
		messageContent = "⏳ Testing connection..."
	} else if m.IsConnecting {
		messageContent = "⏳ Connecting to database..."
	} else if m.Err != nil {
		messageContent = styles.ErrorStyle.Render("❌ " + m.Err.Error())
	} else if m.QueryResult != "" {
		messageContent = styles.SuccessStyle.Render(m.QueryResult)
	}

	nameLabel := styles.SubtitleStyle.Render("Connection Name:")
	var nameField string
	if m.NameInput.Focused() {
		nameField = styles.InputFocusedStyle.Render(m.NameInput.View())
	} else {
		nameField = styles.InputStyle.Render(m.NameInput.View())
	}

	connLabel := styles.SubtitleStyle.Render("Connection String:")
	var connField string
	if m.TextInput.Focused() {
		connField = styles.InputFocusedStyle.Render(m.TextInput.View())
	} else {
		connField = styles.InputStyle.Render(m.TextInput.View())
	}

	var exampleText string
	switch m.SelectedDB.Driver {
	case "postgres":
		exampleText = "postgres://user:password@localhost/dbname?sslmode=disable"
	case "mysql":
		exampleText = "user:password@tcp(localhost:3306)/dbname"
	case "sqlite3":
		exampleText = "./database.db or /path/to/database.db"
	}

	examples := styles.InfoStyle.Render(
		styles.SubtitleStyle.Render("Examples:") + "\n" + exampleText,
	)

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("Enter") + ": save and connect • " +
			styles.KeyStyle.Render("F1") + ": test connection • " +
			styles.KeyStyle.Render("Tab") + ": switch fields • " +
			styles.KeyStyle.Render("Esc") + ": back",
	)

	var elements []string
	elements = append(elements, title)

	if messageContent != "" {
		elements = append(elements, messageContent)
	}

	elements = append(elements,
		nameLabel,
		nameField,
		connLabel,
		connField,
		examples,
		helpText,
	)

	content := lipgloss.JoinVertical(lipgloss.Left, elements...)
	return styles.DocStyle.Render(content)
}

// SaveConnectionView renders the connection saving screen
func SaveConnectionView(m models.Model) string {
	content := styles.TitleStyle.Render("Save Connection") + "\n\n"
	content += "Name for this connection:\n"
	content += m.NameInput.View() + "\n\n"
	content += "Connection to save:\n"
	content += styles.HelpStyle.Render(fmt.Sprintf("%s: %s", m.SelectedDB.Name, m.ConnectionStr))
	content += "\n\n" + styles.HelpStyle.Render("enter: save • esc: cancel")
	return styles.DocStyle.Render(content)
}

// EditConnectionView renders the connection editing screen
func EditConnectionView(m models.Model) string {
	content := styles.TitleStyle.Render("Edit Connection") + "\n\n"

	if m.Err != nil {
		content += styles.ErrorStyle.Render("❌ Error: "+m.Err.Error()) + "\n\n"
	}

	content += "Connection name:\n"
	content += m.NameInput.View() + "\n\n"

	content += fmt.Sprintf("Database type: %s\n", m.SelectedDB.Name)
	content += "Connection string:\n"
	content += m.TextInput.View() + "\n\n"

	content += "Examples:\n"
	switch m.SelectedDB.Driver {
	case "postgres":
		content += styles.HelpStyle.Render("postgres://user:password@localhost/dbname?sslmode=disable")
	case "mysql":
		content += styles.HelpStyle.Render("user:password@tcp(localhost:3306)/dbname")
	case "sqlite3":
		content += styles.HelpStyle.Render("./database.db or /path/to/database.db")
	}

	content += "\n\n" + styles.HelpStyle.Render("enter: save changes • tab: switch fields • esc: cancel")
	return styles.DocStyle.Render(content)
}

// SchemaView renders the schema selection screen
func SchemaView(m models.Model) string {
	var content string

	if m.IsLoadingSchemas {
		loadingMsg := "⏳ Loading schemas..."
		content = m.SchemasList.View() + "\n" + loadingMsg
	} else if len(m.Schemas) == 0 {
		emptyMsg := styles.InfoStyle.Render("🗂️ No additional schemas found.\n\nUsing default schema.")
		content = m.SchemasList.View() + "\n" + emptyMsg
	} else {
		content = m.SchemasList.View()
	}

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": select schema • " +
			styles.KeyStyle.Render("esc") + ": back",
	)

	return styles.DocStyle.Render(content + "\n" + helpText)
}

// TablesView renders the tables listing screen
func TablesView(m models.Model) string {
	var elements []string

	if m.IsLoadingColumns {
		loadingMsg := "⏳ Loading table columns..."
		elements = append(elements, m.TablesList.View())
		elements = append(elements, loadingMsg)
	} else if len(m.Tables) == 0 {
		emptyMsg := styles.InfoStyle.Render("📋 No tables found in this database.")
		elements = append(elements, m.TablesList.View())
		elements = append(elements, emptyMsg)
	} else {
		// Show tables list without success banner
		elements = append(elements, m.TablesList.View())
	}

	content := lipgloss.JoinVertical(lipgloss.Left, elements...)

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": preview data • " +
			styles.KeyStyle.Render("v") + ": view columns • " +
			styles.KeyStyle.Render("f") + ": relationships • " +
			styles.KeyStyle.Render("esc") + ": disconnect")

	return styles.DocStyle.Render(content + "\n" + helpText)
}

// ColumnsView renders the table columns display screen
func ColumnsView(m models.Model) string {
	content := styles.TitleStyle.Render(fmt.Sprintf("Columns of table: %s", m.SelectedTable)) + "\n\n"
	content += m.ColumnsTable.View()

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("↑/↓") + ": navigate • " +
			styles.KeyStyle.Render("esc") + ": back to tables")

	content += "\n" + helpText
	return styles.DocStyle.Render(content)
}

// QueryView renders the SQL query execution screen
func QueryView(m models.Model) string {
	title := styles.TitleStyle.Render("⚡  SQL Query Runner")

	var messageContent string
	if m.IsExecutingQuery {
		messageContent = "⏳ Executing query..."
	} else if m.IsExporting {
		messageContent = "⏳ Exporting data..."
	} else if m.Err != nil {
		messageContent = styles.ErrorStyle.Render("❌ " + m.Err.Error())
	}

	// Query input with enhanced styling
	queryLabel := styles.SubtitleStyle.Render("💻 Enter SQL Query:")
	var queryField string
	if m.QueryInput.Focused() {
		queryField = styles.InputFocusedStyle.Render(m.QueryInput.View())
	} else {
		queryField = styles.InputStyle.Render(m.QueryInput.View())
	}

	var resultContent string
	if m.QueryResult != "" {
		resultLabel := styles.SubtitleStyle.Render("Query Result:")
		resultText := styles.SuccessStyle.Render(m.QueryResult)

		// Only show the table if it has both columns and rows, and they match
		if len(m.QueryResultsTable.Columns()) > 0 && len(m.QueryResultsTable.Rows()) > 0 {
			tableContent := styles.CardStyle.Render(m.QueryResultsTable.View())
			resultContent = lipgloss.JoinVertical(lipgloss.Left, resultLabel, resultText, tableContent)
		} else {
			resultContent = lipgloss.JoinVertical(lipgloss.Left, resultLabel, resultText)
		}
	}

	// Examples in an info box
	examples := styles.InfoStyle.Render(
		styles.SubtitleStyle.Render("💡 Examples:") + "\n" +
			styles.KeyStyle.Render("SELECT") + " * FROM users LIMIT 10;\n" +
			styles.KeyStyle.Render("INSERT") + " INTO users (name, email) VALUES ('John', 'john@example.com');\n" +
			styles.KeyStyle.Render("UPDATE") + " users SET email = 'new@example.com' WHERE id = 1;\n" +
			styles.KeyStyle.Render("DELETE") + " FROM users WHERE id = 1;",
	)

	// Help text with enhanced key styling
	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("Enter") + ": execute query • " +
			styles.KeyStyle.Render("Tab") + ": switch focus • " +
			styles.KeyStyle.Render("↑/↓") + ": navigate results • " +
			styles.KeyStyle.Render("Ctrl+E") + ": export CSV • " +
			styles.KeyStyle.Render("Ctrl+J") + ": export JSON • " +
			styles.KeyStyle.Render("Esc") + ": back to tables",
	)

	// Assemble content with proper spacing
	var elements []string
	elements = append(elements, title)

	if messageContent != "" {
		elements = append(elements, messageContent)
	}

	elements = append(elements, queryLabel, queryField)

	if resultContent != "" {
		elements = append(elements, resultContent)
	}

	elements = append(elements, examples, helpText)

	content := lipgloss.JoinVertical(lipgloss.Left, elements...)
	return styles.DocStyle.Render(content)
}

// QueryHistoryView renders the query history screen
func QueryHistoryView(m models.Model) string {
	var content string

	if len(m.QueryHistory) == 0 {
		emptyMsg := styles.InfoStyle.Render("📝 No query history yet.\n\nExecute some queries to see them here!")
		content = m.QueryHistoryList.View() + "\n" + emptyMsg
	} else {
		content = m.QueryHistoryList.View()
	}

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": use query • " +
			styles.KeyStyle.Render("d") + ": delete • " +
			styles.KeyStyle.Render("esc") + ": back",
	)

	return styles.DocStyle.Render(content + "\n" + helpText)
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
			"🔍 " + styles.KeyStyle.Render("ENTER") + ": apply filter • " +
				styles.KeyStyle.Render("ESC") + ": cancel filter")
	} else if m.DataPreviewSortMode {
		helpText = styles.HelpStyle.Render(
			"🎯 " + styles.KeyStyle.Render("↑↓") + ": select column • " +
				styles.KeyStyle.Render("ENTER") + ": cycle sort (off→asc→desc) • " +
				styles.KeyStyle.Render("ESC") + ": exit sort")
	} else {
		// Group help text by function for better readability
		navigationHelp := styles.KeyStyle.Render("hjkl/↑↓←→") + ": navigate"
		actionHelp := styles.KeyStyle.Render("ENTER") + ": row details"
		pageHelp := styles.KeyStyle.Render("←→") + ": pages"
		featureHelp := styles.KeyStyle.Render("/") + ": filter • " + styles.KeyStyle.Render("s") + ": sort"
		utilityHelp := styles.KeyStyle.Render("r") + ": reload • " + styles.KeyStyle.Render("ESC") + ": back"

		helpText = styles.HelpStyle.Render(
			"📍 " + navigationHelp + " • " + actionHelp + " • " + pageHelp + "\n" +
				"⚡ " + featureHelp + " • " + utilityHelp)
	}

	content += "\n" + helpText
	return styles.DocStyle.Render(content)
}

// IndexesView renders the table indexes and constraints screen
func IndexesView(m models.Model) string {
	content := styles.TitleStyle.Render(fmt.Sprintf("🔑 Indexes & Constraints: %s", m.SelectedTable)) + "\n\n"

	if m.Err != nil {
		content += styles.ErrorStyle.Render("❌ "+m.Err.Error()) + "\n\n"
	}

	// Show the indexes table
	content += m.IndexesTable.View() + "\n\n"

	// Help text
	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("↑/↓") + ": navigate • " +
			styles.KeyStyle.Render("enter") + ": view details • " +
			styles.KeyStyle.Render("esc") + ": back to columns")

	content += "\n" + helpText
	return styles.DocStyle.Render(content)
}

// RelationshipsView renders the foreign key relationships screen
func RelationshipsView(m models.Model) string {
	content := styles.TitleStyle.Render("🔗 Foreign Key Relationships") + "\n\n"

	if m.Err != nil {
		content += styles.ErrorStyle.Render("❌ "+m.Err.Error()) + "\n\n"
	}

	// Show the relationships table
	content += m.RelationshipsTable.View() + "\n\n"

	// Help text
	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("↑/↓") + ": navigate • " +
			styles.KeyStyle.Render("esc") + ": back to tables")

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
		availableHeight := m.Height - 10 // Reserve space for title, help text, and margins
		if availableHeight < 5 {
			availableHeight = 5 // Minimum height
		}

		// Calculate visible range
		startLine := m.FieldDetailScrollOffset
		endLine := startLine + availableHeight
		if endLine > len(lines) {
			endLine = len(lines)
		}

		// Calculate dynamic width (use window width minus padding)
		availableWidth := m.Width - 10 // Reserve space for margins
		if availableWidth < 40 {
			availableWidth = 40 // Minimum width
		}
		if availableWidth > 200 {
			availableWidth = 200 // Maximum width for readability
		}

		// Build visible content with horizontal scrolling
		var visibleLines []string
		for i := startLine; i < endLine; i++ {
			line := lines[i]
			// Apply horizontal scrolling
			if m.FieldDetailHorizontalOffset < len(line) {
				endChar := m.FieldDetailHorizontalOffset + availableWidth
				if endChar > len(line) {
					endChar = len(line)
				}
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
		endDisplayLine := m.FieldDetailScrollOffset + len(visibleLines)
		if endDisplayLine > len(lines) {
			endDisplayLine = len(lines)
		}

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
