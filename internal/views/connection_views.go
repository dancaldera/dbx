package views

import (
	"fmt"
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