package views

import (
	"fmt"

	"github.com/dancaldera/dbx/internal/models"
	"github.com/dancaldera/dbx/internal/styles"
)

// DBTypeView renders the database type selection screen
func DBTypeView(m models.Model) string {
	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": select • " +
			styles.KeyStyle.Render("s") + ": saved connections • " +
			styles.KeyStyle.Render("q") + ": quit",
	)

	return NewViewBuilder().
		WithContent(m.DBTypeList.View()).
		WithHelp(helpText).
		Render()
}

// SavedConnectionsView renders the saved connections screen
func SavedConnectionsView(m models.Model) string {
	builder := NewViewBuilder().WithTitle("📋 Saved Connections")

	// Determine status message and type
	if m.IsConnecting {
		statusMsg := "⏳ Connecting..."
		if selectedItem, ok := m.SavedConnectionsList.SelectedItem().(models.Item); ok {
			statusMsg = fmt.Sprintf("⏳ Connecting to %s...", selectedItem.ItemTitle)
		}
		builder.WithStatus(statusMsg, StatusLoading)
	} else if m.Err != nil {
		builder.WithStatus("🚨 "+m.Err.Error(), StatusError)
	} else if m.QueryResult != "" {
		builder.WithStatus(m.QueryResult, StatusSuccess)
	}

	// Handle empty state
	if len(m.SavedConnections) == 0 && !m.IsConnecting && m.Err == nil && m.QueryResult == "" {
		emptyState := RenderEmptyState("📝", "No saved connections yet.\n\nGo back and create your first connection!")
		builder.WithContent(emptyState)
	} else {
		builder.WithContent(m.SavedConnectionsList.View())
	}

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": connect • " +
			styles.KeyStyle.Render("c") + ": copy to clipboard • " +
			styles.KeyStyle.Render("d") + ": delete • " +
			styles.KeyStyle.Render("esc") + ": back",
	)

	return builder.WithHelp(helpText).Render()
}

// ConnectionView renders the database connection configuration screen
func ConnectionView(m models.Model) string {
	// Determine database icon
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

	title := fmt.Sprintf("%s  Connect to %s", dbIcon, m.SelectedDB.Name)
	builder := NewViewBuilder().WithTitle(title)

	// Add status messages
	if m.IsTestingConnection {
		builder.WithStatus("⏳ Testing connection...", StatusLoading)
	} else if m.IsConnecting {
		builder.WithStatus("⏳ Connecting to database...", StatusLoading)
	} else if m.Err != nil {
		builder.WithStatus("❌ "+m.Err.Error(), StatusError)
	} else if m.QueryResult != "" {
		builder.WithStatus(m.QueryResult, StatusSuccess)
	}

	// Input fields
	nameField := RenderInputField("Connection Name:", m.NameInput.View(), m.NameInput.Focused())
	connField := RenderInputField("Connection String:", m.TextInput.View(), m.TextInput.Focused())

	// Examples based on database type
	var exampleText string
	switch m.SelectedDB.Driver {
	case "postgres":
		exampleText = "postgres://user:password@localhost/dbname?sslmode=disable"
	case "mysql":
		exampleText = "user:password@tcp(localhost:3306)/dbname"
	case "sqlite3":
		exampleText = "./database.db or /path/to/database.db"
	}
	examples := RenderInfoBox(styles.SubtitleStyle.Render("Examples:") + "\n" + exampleText)

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("Enter") + ": save and connect • " +
			styles.KeyStyle.Render("F1") + ": test connection • " +
			styles.KeyStyle.Render("Tab") + ": switch fields • " +
			styles.KeyStyle.Render("Esc") + ": back",
	)

	return builder.
		WithContent(nameField, connField, examples).
		WithHelp(helpText).
		Render()
}

// SaveConnectionView renders the connection saving screen
func SaveConnectionView(m models.Model) string {
	nameField := RenderInputField("Name for this connection:", m.NameInput.View(), m.NameInput.Focused())

	connectionInfo := styles.SubtitleStyle.Render("Connection to save:") + "\n" +
		styles.HelpStyle.Render(fmt.Sprintf("%s: %s", m.SelectedDB.Name, m.ConnectionStr))

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": save • " +
			styles.KeyStyle.Render("esc") + ": cancel",
	)

	return NewViewBuilder().
		WithTitle("Save Connection").
		WithContent(nameField, connectionInfo).
		WithHelp(helpText).
		Render()
}

// EditConnectionView renders the connection editing screen
func EditConnectionView(m models.Model) string {
	builder := NewViewBuilder().WithTitle("Edit Connection")

	// Add error status if present
	if m.Err != nil {
		builder.WithStatus("❌ Error: "+m.Err.Error(), StatusError)
	}

	// Connection name field
	nameField := RenderInputField("Connection name:", m.NameInput.View(), m.NameInput.Focused())

	// Database type
	dbType := fmt.Sprintf("Database type: %s", m.SelectedDB.Name)

	// Connection string field
	connField := RenderInputField("Connection string:", m.TextInput.View(), m.TextInput.Focused())

	// Examples
	var exampleText string
	switch m.SelectedDB.Driver {
	case "postgres":
		exampleText = "postgres://user:password@localhost/dbname?sslmode=disable"
	case "mysql":
		exampleText = "user:password@tcp(localhost:3306)/dbname"
	case "sqlite3":
		exampleText = "./database.db or /path/to/database.db"
	}
	examples := RenderInfoBox(styles.SubtitleStyle.Render("Examples:") + "\n" + exampleText)

	// Help text
	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": save changes • " +
			styles.KeyStyle.Render("tab") + ": switch fields • " +
			styles.KeyStyle.Render("esc") + ": cancel",
	)

	return builder.
		WithContent(nameField, dbType, connField, examples).
		WithHelp(helpText).
		Render()
}
