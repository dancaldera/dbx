package views

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/danielcaldera/dbx/internal/models"
	"github.com/danielcaldera/dbx/internal/styles"
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
	var content string

	if m.IsConnecting {
		loadingMsg := "⏳ Connecting to saved connection..."
		content = m.SavedConnectionsList.View() + "\n" + loadingMsg
	} else if len(m.SavedConnections) == 0 {
		emptyMsg := styles.InfoStyle.Render("📝 No saved connections yet.\n\nGo back and create your first connection!")
		content = m.SavedConnectionsList.View() + "\n" + emptyMsg
	} else {
		content = m.SavedConnectionsList.View()
	}

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": connect • " +
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
		statusText := styles.SuccessStyle.Render(fmt.Sprintf("✅ Connected successfully (%d tables found)", len(m.Tables)))
		elements = append(elements, statusText)
		elements = append(elements, "")
		elements = append(elements, m.TablesList.View())
	}

	content := lipgloss.JoinVertical(lipgloss.Left, elements...)

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": view columns • " +
			styles.KeyStyle.Render("s") + ": save connection • " +
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
		resultLabel := styles.SubtitleStyle.Render("📊 Query Result:")
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

// DataPreviewView renders the table data preview screen
func DataPreviewView(m models.Model) string {
	content := styles.TitleStyle.Render(fmt.Sprintf("Data Preview: %s", m.SelectedTable)) + "\n\n"

	// Show status messages (loading, success, error)
	if m.IsExporting {
		content += "⏳ Exporting data..." + "\n\n"
	} else if m.Err != nil {
		content += styles.ErrorStyle.Render("❌ Error: "+m.Err.Error()) + "\n\n"
	} else if m.QueryResult != "" {
		content += styles.SuccessStyle.Render(m.QueryResult) + "\n\n"
	}

	// Only show the table if it has both columns and rows
	if len(m.DataPreviewTable.Columns()) > 0 && len(m.DataPreviewTable.Rows()) > 0 {
		content += styles.InfoStyle.Render(fmt.Sprintf("Showing first 10 rows from %s", m.SelectedTable)) + "\n\n"
		content += m.DataPreviewTable.View()
	} else if m.Err == nil && m.QueryResult == "" && !m.IsExporting {
		content += styles.InfoStyle.Render("No data to display")
	}

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": view row details • " +
			styles.KeyStyle.Render("r") + ": reload • " +
			styles.KeyStyle.Render("Ctrl+E") + ": export CSV • " +
			styles.KeyStyle.Render("Ctrl+J") + ": export JSON • " +
			styles.KeyStyle.Render("esc") + ": back to tables")

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