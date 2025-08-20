package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/danielcaldera/dbx/internal/config"
	"github.com/danielcaldera/dbx/internal/models"
	"github.com/danielcaldera/dbx/internal/styles"
)

// Available database types
var DBTypes = []models.DBType{
	{"PostgreSQL", "postgres"},
	{"MySQL", "mysql"},
	{"SQLite", "sqlite3"},
}

// InitialModel creates and initializes the application model
func InitialModel() models.Model {
	// Database types list
	items := make([]list.Item, len(DBTypes))
	for i, db := range DBTypes {
		items[i] = models.Item{
			ItemTitle: db.Name,
			ItemDesc:  fmt.Sprintf("Connect to %s database", db.Name),
		}
	}

	dbList := list.New(items, list.NewDefaultDelegate(), 0, 0)
    dbList.Title = "DBX — Database Explorer"
	// Remove any default title background and apply our title style
	ls := list.DefaultStyles()
	ls.Title = styles.ListTitleStyle
	ls.TitleBar = lipgloss.NewStyle()
	dbList.Styles = ls
	dbList.SetShowStatusBar(false)
	dbList.SetFilteringEnabled(false)
	dbList.SetShowHelp(false)

	// Load saved connections
	savedConnections, _ := config.LoadSavedConnections()

	// Load query history
	queryHistory, _ := config.LoadQueryHistory()

	// Saved connections list
	savedConnectionsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 20)
    savedConnectionsList.Title = "Saved Connections"
	// Transparent title bar for saved connections
	scLS := list.DefaultStyles()
	scLS.Title = styles.ListTitleStyle
	scLS.TitleBar = lipgloss.NewStyle()
	savedConnectionsList.Styles = scLS
	savedConnectionsList.SetShowStatusBar(false)
	savedConnectionsList.SetFilteringEnabled(false)
	savedConnectionsList.SetShowHelp(false)

	// Populate the list with saved connections
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
	savedConnectionsList.SetItems(savedItems)

	// Connection input
	ti := textinput.New()
	ti.Placeholder = "Enter connection string..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 80

	// Connection name input
	ni := textinput.New()
	ni.Placeholder = "Name for this connection..."
	ni.CharLimit = 100
	ni.Width = 80

	// Query input
	qi := textinput.New()
	qi.Placeholder = "Enter SQL query (e.g., SELECT * FROM table_name LIMIT 10)..."
	qi.CharLimit = 1000
	qi.Width = 80

	// Search input
	si := textinput.New()
	si.Placeholder = "Type to search..."
	si.CharLimit = 100
	si.Width = 80

	// Tables list (compact: names only, no extra spacing)
	tblDelegate := list.NewDefaultDelegate()
	tblDelegate.ShowDescription = false
	tblDelegate.SetSpacing(0)
	tablesList := list.New([]list.Item{}, tblDelegate, 0, 0)
	tablesList.Title = "Available Tables"
	// Transparent title bar for tables
	tblLS := list.DefaultStyles()
	tblLS.Title = styles.ListTitleStyle
	tblLS.TitleBar = lipgloss.NewStyle()
	tablesList.Styles = tblLS
	tablesList.SetShowStatusBar(false)
	tablesList.SetFilteringEnabled(false)
	tablesList.SetShowHelp(false)

	// Columns table
	columns := []table.Column{
		{Title: "Column", Width: 20},
		{Title: "Type", Width: 15},
		{Title: "Null", Width: 8},
		{Title: "Default", Width: 15},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	t.SetStyles(styles.GetMagentaTableStyles())

	// Query results table
	queryResultsTable := table.New(
		table.WithColumns([]table.Column{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	queryResultsTable.SetStyles(styles.GetMagentaTableStyles())

	// Data preview table
	dataPreviewTable := table.New(
		table.WithColumns([]table.Column{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	dataPreviewTable.SetStyles(styles.GetMagentaTableStyles())

	// Indexes table
	indexesTable := table.New(
		table.WithColumns([]table.Column{
			{Title: "Index Name", Width: 20},
			{Title: "Type", Width: 12},
			{Title: "Columns", Width: 25},
			{Title: "Definition", Width: 40},
		}),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	indexesTable.SetStyles(styles.GetMagentaTableStyles())

	// Relationships table
	relationshipsTable := table.New(
		table.WithColumns([]table.Column{
			{Title: "From Table", Width: 20},
			{Title: "From Column", Width: 20},
			{Title: "To Table", Width: 20},
			{Title: "To Column", Width: 20},
			{Title: "Constraint Name", Width: 25},
		}),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	relationshipsTable.SetStyles(styles.GetMagentaTableStyles())

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.AccentMagenta)

	// Initialize textarea for field editing
	ta := textarea.New()
	ta.Placeholder = "Enter field content..."
	ta.SetWidth(80)
	ta.SetHeight(20)

	// Query history list
	queryHistoryList := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 20)
    queryHistoryList.Title = "Query History"
	// Transparent title bar for query history
	qhLS := list.DefaultStyles()
	qhLS.Title = styles.ListTitleStyle
	qhLS.TitleBar = lipgloss.NewStyle()
	queryHistoryList.Styles = qhLS
	queryHistoryList.SetShowStatusBar(false)
	queryHistoryList.SetFilteringEnabled(false)
	queryHistoryList.SetShowHelp(false)

	// Schemas list
	schemasList := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 20)
    schemasList.Title = "Database Schemas"
	// Transparent title bar for schemas
	schLS := list.DefaultStyles()
	schLS.Title = styles.ListTitleStyle
	schLS.TitleBar = lipgloss.NewStyle()
	schemasList.Styles = schLS
	schemasList.SetShowStatusBar(false)
	schemasList.SetFilteringEnabled(false)
	schemasList.SetShowHelp(false)

	m := models.Model{
		State:                   models.DBTypeView,
		DBTypeList:              dbList,
		SavedConnectionsList:    savedConnectionsList,
		TextInput:               ti,
		NameInput:               ni,
		QueryInput:              qi,
		SearchInput:             si,
		TablesList:              tablesList,
		ColumnsTable:            t,
		QueryResultsTable:       queryResultsTable,
		DataPreviewTable:        dataPreviewTable,
		IndexesTable:            indexesTable,
		RelationshipsTable:      relationshipsTable,
		QueryHistoryList:        queryHistoryList,
		SchemasList:             schemasList,
		SelectedSchema:          "public", // Default to public schema for PostgreSQL
		SavedConnections:        savedConnections,
		QueryHistory:            queryHistory,
		EditingConnectionIdx:    -1,
		Spinner:                 s,
		FullTextItemsPerPage:    5,   // Show 5 fields per page in full text view
		FieldDetailLinesPerPage: 25,  // Show 25 lines per page in field detail view
		FieldDetailCharsPerLine: 120, // Show 120 characters per line in field detail view
		FieldTextarea:           ta,  // Initialize textarea for field editing
	}

	// Initialize query history list with loaded data
	updateQueryHistoryList(&m)

	return m
}

// Helper function to update query history list
func updateQueryHistoryList(m *models.Model) {
	if len(m.QueryHistory) == 0 {
		m.QueryHistoryList.SetItems([]list.Item{})
		return
	}

	// Show newest queries first
	items := make([]list.Item, len(m.QueryHistory))
	for i := range m.QueryHistory {
		// Reverse order - newest first
		reverseIndex := len(m.QueryHistory) - 1 - i
		reversedEntry := m.QueryHistory[reverseIndex]

		// Truncate long queries for display
		displayQuery := reversedEntry.Query
		if len(displayQuery) > 80 {
			displayQuery = displayQuery[:80] + "..."
		}

		// Create status indicator
		statusIcon := "✅"
		if !reversedEntry.Success {
			statusIcon = "❌"
		}

		// Format timestamp
		timeStr := reversedEntry.Timestamp.Format("2006-01-02 15:04:05")

		var description string
		if reversedEntry.Success && reversedEntry.RowCount > 0 {
			description = fmt.Sprintf("%s %s • %d rows • %s", statusIcon, timeStr, reversedEntry.RowCount, reversedEntry.Database)
		} else {
			description = fmt.Sprintf("%s %s • %s", statusIcon, timeStr, reversedEntry.Database)
		}

		items[i] = models.Item{
			ItemTitle: displayQuery,
			ItemDesc:  description,
		}
	}

	m.QueryHistoryList.SetItems(items)
}
