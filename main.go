package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Global styles with magenta theme
var (
	// Primary magenta colors
	primaryMagenta = lipgloss.Color("#D946EF") // Main magenta
	lightMagenta   = lipgloss.Color("#F3E8FF") // Light magenta background
	darkMagenta    = lipgloss.Color("#7C2D91") // Dark magenta
	accentMagenta  = lipgloss.Color("#A855F7") // Purple accent

	// Supporting colors
	darkGray      = lipgloss.Color("#374151")
	lightGray     = lipgloss.Color("#9CA3AF")
	white         = lipgloss.Color("#FFFFFF")
	successGreen  = lipgloss.Color("#10B981")
	errorRed      = lipgloss.Color("#EF4444")
	warningOrange = lipgloss.Color("#F59E0B")

	// Main title style with gradient-like magenta
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryMagenta).
			Padding(0, 1).
			Margin(0, 0, 1, 0).
			Bold(true)

	// Subtitle for sections
	subtitleStyle = lipgloss.NewStyle().
			Foreground(darkMagenta).
			Bold(true).
			Margin(0, 0, 1, 0)

	// Focused/selected item style
	focusedStyle = lipgloss.NewStyle().
			Foreground(accentMagenta).
			Padding(0, 1).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryMagenta)

	// Input field styling
	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryMagenta).
			Padding(0, 1).
			Margin(0, 0, 1, 0)

	// Input field when focused
	inputFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(primaryMagenta).
				Padding(0, 1).
				Margin(0, 0, 1, 0)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(lightGray).
			Italic(true).
			Margin(1, 0).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lightGray).
			Padding(0, 1)

	// Key binding help style
	keyStyle = lipgloss.NewStyle().
			Foreground(accentMagenta).
			Bold(true)

	// Error messages
	errorStyle = lipgloss.NewStyle().
			Foreground(errorRed).
			Padding(0, 1).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorRed)

	// Success messages
	successStyle = lipgloss.NewStyle().
			Foreground(successGreen).
			Padding(0, 1).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(successGreen)

	// Warning messages
	warningStyle = lipgloss.NewStyle().
			Foreground(warningOrange).
			Padding(0, 1).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(warningOrange)

	// Information boxes
	infoStyle = lipgloss.NewStyle().
			Foreground(darkMagenta).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryMagenta).
			Margin(0, 0, 1, 0)

	// Table header style
	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(darkMagenta).
				Bold(true).
				Padding(0, 1).
				Align(lipgloss.Center)

	// Main document container
	docStyle = lipgloss.NewStyle().
			Margin(2, 2).
			Padding(1)

	// Card-like container for sections
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentMagenta).
			Padding(1, 2).
			Margin(0, 0, 1, 0)

	// Loading indicator style
	loadingStyle = lipgloss.NewStyle().
			Foreground(accentMagenta).
			Bold(true).
			Italic(true)
)

// Helper function to get loading text with spinner
func (m model) getLoadingText(message string) string {
	return loadingStyle.Render(m.spinner.View() + " " + message)
}

// Helper function to get magenta-themed table styles
func getMagentaTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		Foreground(darkMagenta).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryMagenta).
		BorderBottom(true).
		Bold(true).
		Align(lipgloss.Center)
	s.Selected = s.Selected.
		Foreground(accentMagenta).
		Bold(true)
	s.Cell = s.Cell.
		Padding(0, 1)
	return s
}

// Application states
type viewState int

const (
	dbTypeView viewState = iota
	savedConnectionsView
	connectionView
	saveConnectionView
	editConnectionView
	schemaView
	tablesView
	columnsView
	queryView
	queryHistoryView
	dataPreviewView
	rowDetailView
	fieldDetailView
	indexesView
	indexDetailView
)

// Saved connection
type SavedConnection struct {
	Name          string `json:"name"`
	Driver        string `json:"driver"`
	ConnectionStr string `json:"connection_str"`
}

// Query history entry
type QueryHistoryEntry struct {
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database,omitempty"`
	Success   bool      `json:"success"`
	RowCount  int       `json:"row_count,omitempty"`
}

// Database types
type dbType struct {
	name   string
	driver string
}

// Schema information
type schemaInfo struct {
	name        string
	description string
}

// Table information
type tableInfo struct {
	name        string
	schema      string
	tableType   string
	rowCount    int64
	description string
}

var dbTypes = []dbType{
	{"PostgreSQL", "postgres"},
	{"MySQL", "mysql"},
	{"SQLite", "sqlite3"},
}

// List item
type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// Main model
type model struct {
	state                viewState
	dbTypeList           list.Model
	savedConnectionsList list.Model
	textInput            textinput.Model
	nameInput            textinput.Model
	queryInput           textinput.Model
	tablesList           list.Model
	columnsTable         table.Model
	queryResultsTable    table.Model
	dataPreviewTable     table.Model
	indexesTable         table.Model
	selectedDB           dbType
	connectionStr        string
	db                   *sql.DB
	err                  error
	tables               []string
	tableInfos           []tableInfo
	selectedTable        string
	schemas              []schemaInfo
	selectedSchema       string
	schemasList          list.Model
	isLoadingSchemas     bool
	savedConnections     []SavedConnection
	editingConnectionIdx int
	queryResult          string
	width                int
	height               int
	// Loading states
	isTestingConnection bool
	isConnecting        bool
	isSavingConnection  bool
	isLoadingTables     bool
	isLoadingColumns    bool
	isExecutingQuery    bool
	isLoadingPreview    bool
	// Export states
	isExporting        bool
	lastQueryColumns   []string
	lastQueryRows      [][]string
	lastPreviewColumns []string
	lastPreviewRows    [][]string
	// Spinner for animations
	spinner spinner.Model
	// Search functionality
	searchInput        textinput.Model
	isSearchingTables  bool
	isSearchingColumns bool
	originalTableItems []list.Item
	originalTableRows  []table.Row
	searchTerm         string
	// Query history functionality
	queryHistory     []QueryHistoryEntry
	queryHistoryList list.Model
	isViewingHistory bool
	// Row detail functionality
	selectedRowData        []string
	selectedRowIndex       int
	rowDetailCurrentPage   int
	rowDetailItemsPerPage  int
	rowDetailSelectedField int
	isViewingFullText      bool
	fullTextScrollOffset   int
	fullTextLinesPerPage   int
	// Full text view pagination
	fullTextCurrentPage   int
	fullTextItemsPerPage  int
	fullTextSelectedField int
	// Individual field detail view
	selectedFieldName           string
	selectedFieldValue          string
	selectedFieldIndex          int
	fieldDetailScrollOffset     int
	fieldDetailHorizontalOffset int
	fieldDetailLinesPerPage     int
	fieldDetailCharsPerLine     int
	// Index detail view
	selectedIndexName       string
	selectedIndexType       string
	selectedIndexColumns    string
	selectedIndexDefinition string
}

func initialModel() model {
	// Database types list
	items := make([]list.Item, len(dbTypes))
	for i, db := range dbTypes {
		items[i] = item{
			title: db.name,
			desc:  fmt.Sprintf("Connect to %s database", db.name),
		}
	}

	dbList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	dbList.Title = "🗄️  DBX - Database Explorer"
	dbList.SetShowStatusBar(false)
	dbList.SetFilteringEnabled(false)
	dbList.SetShowHelp(false)

	// Load saved connections
	savedConnections, _ := loadSavedConnections()

	// Load query history
	queryHistory, _ := loadQueryHistory()

	// Saved connections list
	savedConnectionsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 20)
	savedConnectionsList.Title = "💾 Saved Connections"
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
		savedItems[i] = item{
			title: conn.Name,
			desc:  fmt.Sprintf("%s - %s", conn.Driver, connStr),
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

	// Tables list
	tablesList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	tablesList.Title = "📊 Available Tables"
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

	t.SetStyles(getMagentaTableStyles())

	// Query results table
	queryResultsTable := table.New(
		table.WithColumns([]table.Column{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	queryResultsTable.SetStyles(getMagentaTableStyles())

	// Data preview table
	dataPreviewTable := table.New(
		table.WithColumns([]table.Column{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	dataPreviewTable.SetStyles(getMagentaTableStyles())

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
	indexesTable.SetStyles(getMagentaTableStyles())

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(accentMagenta)

	// Query history list
	queryHistoryList := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 20)
	queryHistoryList.Title = "📝 Query History"
	queryHistoryList.SetShowStatusBar(false)
	queryHistoryList.SetFilteringEnabled(false)
	queryHistoryList.SetShowHelp(false)

	// Schemas list
	schemasList := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 20)
	schemasList.Title = "🗂️ Database Schemas"
	schemasList.SetShowStatusBar(false)
	schemasList.SetFilteringEnabled(false)
	schemasList.SetShowHelp(false)

	m := model{
		state:                   dbTypeView,
		dbTypeList:              dbList,
		savedConnectionsList:    savedConnectionsList,
		textInput:               ti,
		nameInput:               ni,
		queryInput:              qi,
		searchInput:             si,
		tablesList:              tablesList,
		columnsTable:            t,
		queryResultsTable:       queryResultsTable,
		dataPreviewTable:        dataPreviewTable,
		indexesTable:            indexesTable,
		queryHistoryList:        queryHistoryList,
		schemasList:             schemasList,
		selectedSchema:          "public", // Default to public schema for PostgreSQL
		savedConnections:        savedConnections,
		queryHistory:            queryHistory,
		editingConnectionIdx:    -1,
		spinner:                 s,
		rowDetailItemsPerPage:   8,   // Show 8 fields per page
		fullTextLinesPerPage:    20,  // Show 20 lines per page in full text view
		fullTextItemsPerPage:    5,   // Show 5 fields per page in full text view
		fieldDetailLinesPerPage: 25,  // Show 25 lines per page in field detail view
		fieldDetailCharsPerLine: 120, // Show 120 characters per line in field detail view
	}

	// Initialize query history list with loaded data
	m.updateQueryHistoryList()

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle connection and column result messages first
	switch msg := msg.(type) {
	case connectResult:
		return m.handleConnectResult(msg)
	case testConnectionResult:
		return m.handleTestConnectionResult(msg)
	case columnsResult:
		return m.handleColumnsResult(msg)
	case queryResult:
		return m.handleQueryResult(msg)
	case dataPreviewResult:
		return m.handleDataPreviewResult(msg)
	case indexesResult:
		return m.handleIndexesResult(msg)
	case exportResult:
		return m.handleExportResult(msg)
	case testAndSaveResult:
		return m.handleTestAndSaveResult(msg)
	case fieldValueResult:
		return m.handleFieldValueResult(msg)
	case clipboardResult:
		return m.handleClipboardResult(msg)
	case clearResultMsg:
		m.queryResult = ""
		return m, nil
	case clearErrorMsg:
		m.err = nil
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.dbTypeList.SetSize(msg.Width-h, msg.Height-v-5)
		m.savedConnectionsList.SetSize(msg.Width-h, msg.Height-v-5)
		m.tablesList.SetSize(msg.Width-h, msg.Height-v-5)
		m.queryHistoryList.SetSize(msg.Width-h, msg.Height-v-5)
		m.schemasList.SetSize(msg.Width-h, msg.Height-v-5)
		m.textInput.Width = msg.Width - h - 4
		m.nameInput.Width = msg.Width - h - 4
		m.queryInput.Width = msg.Width - h - 4
		m.searchInput.Width = msg.Width - h - 4

		// Update query results table height
		tableHeight := msg.Height - v - 15 // Reserve space for query input and help text
		if tableHeight < 5 {
			tableHeight = 5
		}
		// Simply adjust the height for the existing table
		if len(m.queryResultsTable.Columns()) > 0 {
			cols := m.queryResultsTable.Columns()
			rows := m.queryResultsTable.Rows()
			m.queryResultsTable = table.New(
				table.WithColumns(cols),
				table.WithRows(rows),
				table.WithFocused(false),
				table.WithHeight(tableHeight),
			)
			// Apply magenta theme styles
			m.queryResultsTable.SetStyles(getMagentaTableStyles())
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.db != nil {
				m.db.Close()
			}
			return m, tea.Quit

		case "q":
			// Only allow quit from main database type view
			if m.state == dbTypeView {
				if m.db != nil {
					m.db.Close()
				}
				return m, tea.Quit
			}

		case "esc":
			switch m.state {
			case savedConnectionsView:
				m.state = dbTypeView
				m.err = nil
			case connectionView:
				m.state = dbTypeView
				m.err = nil
			case saveConnectionView:
				m.state = connectionView
				m.err = nil
			case editConnectionView:
				m.state = savedConnectionsView
				m.err = nil
				m.editingConnectionIdx = -1
			case schemaView:
				// Close database connection and go back to connection view
				if m.db != nil {
					m.db.Close()
					m.db = nil
				}
				m.state = connectionView
				m.err = nil
			case tablesView:
				// Exit search mode if active
				if m.isSearchingTables {
					m.isSearchingTables = false
					m.searchInput.Blur()
					m.tablesList.SetItems(m.originalTableItems)
					m.searchTerm = ""
					return m, nil
				}
				// Close database connection and go back to main menu
				if m.db != nil {
					m.db.Close()
					m.db = nil
				}
				m.state = dbTypeView
				m.connectionStr = ""
				m.tables = nil
				m.tableInfos = nil
				m.selectedTable = ""
				m.err = nil
			case columnsView:
				// Exit search mode if active
				if m.isSearchingColumns {
					m.isSearchingColumns = false
					m.searchInput.Blur()
					m.columnsTable.SetRows(m.originalTableRows)
					m.searchTerm = ""
					return m, nil
				}
				m.state = tablesView
			case queryView:
				m.state = tablesView
			case dataPreviewView:
				m.state = tablesView
			case rowDetailView:
				m.state = dataPreviewView
			case fieldDetailView:
				m.state = rowDetailView
			case indexesView:
				m.state = columnsView
			case indexDetailView:
				m.state = indexesView
			case queryHistoryView:
				m.state = queryView
			}
			return m, nil

		case "s":
			if m.state == dbTypeView {
				m.state = savedConnectionsView
				// Reload connections from file
				if connections, err := loadSavedConnections(); err == nil {
					m.savedConnections = connections
				}
				m = m.updateSavedConnectionsList()
				return m, nil
			}
			if m.state == columnsView && m.connectionStr != "" {
				// Go to connection naming view
				m.state = saveConnectionView
				m.nameInput.SetValue("")
				m.nameInput.Focus()
				return m, nil
			}

		case "f1":
			if m.state == connectionView && !m.isTestingConnection {
				// Test connection
				m.connectionStr = m.textInput.Value()
				if m.connectionStr != "" {
					m.isTestingConnection = true
					m.err = nil
					m.queryResult = ""
					return m, m.testConnection()
				}
			}

		case "f2":
			if m.state == connectionView && !m.isSavingConnection && !m.isConnecting && !m.isTestingConnection {
				// Test connection before saving
				name := m.nameInput.Value()
				connectionStr := m.textInput.Value()
				if name != "" && connectionStr != "" {
					m.isSavingConnection = true
					m.err = nil
					m.queryResult = ""
					m.connectionStr = connectionStr
					return m, m.testAndSaveConnection(name, connectionStr)
				}
			}

		case "r":
			// Don't trigger query view if we're in search mode
			if (m.state == tablesView || m.state == columnsView) && m.db != nil && !m.isSearchingTables && !m.isSearchingColumns {
				// Go to query view
				m.state = queryView
				m.queryInput.SetValue("")
				m.queryInput.Focus()
				m.queryResultsTable.Blur()
				return m, nil
			}

		case "ctrl+f", "/":
			// Toggle search mode for tables
			if m.state == tablesView && !m.isSearchingTables {
				m.isSearchingTables = true
				m.originalTableItems = m.tablesList.Items() // Backup original items
				m.searchInput.SetValue("")
				m.searchInput.Focus()
				m.searchTerm = ""
				return m, nil
			}
			// Toggle search mode for columns
			if m.state == columnsView && !m.isSearchingColumns {
				m.isSearchingColumns = true
				m.originalTableRows = m.columnsTable.Rows() // Backup original rows
				m.searchInput.SetValue("")
				m.searchInput.Focus()
				m.searchTerm = ""
				return m, nil
			}

		case "ctrl+e":
			// Export query results to CSV
			if m.state == queryView && !m.isExporting {
				if len(m.lastQueryColumns) == 0 || len(m.lastQueryRows) == 0 {
					m.queryResult = "⚠️ No data to export. Execute a query first!"
					return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
						return clearResultMsg{}
					})
				}
				m.isExporting = true
				m.err = nil
				m.queryResult = ""
				return m, m.exportDataToCSV(m.lastQueryColumns, m.lastQueryRows, "")
			}
			// Export preview data to CSV
			if m.state == dataPreviewView && !m.isExporting {
				if len(m.lastPreviewColumns) == 0 || len(m.lastPreviewRows) == 0 {
					m.queryResult = "⚠️ No data to export. Load table preview first!"
					return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
						return clearResultMsg{}
					})
				}
				m.isExporting = true
				m.err = nil
				m.queryResult = ""
				return m, m.exportDataToCSV(m.lastPreviewColumns, m.lastPreviewRows, m.selectedTable)
			}

		case "ctrl+j":
			// Export query results to JSON
			if m.state == queryView && !m.isExporting {
				if len(m.lastQueryColumns) == 0 || len(m.lastQueryRows) == 0 {
					m.queryResult = "⚠️ No data to export. Execute a query first!"
					return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
						return clearResultMsg{}
					})
				}
				m.isExporting = true
				m.err = nil
				m.queryResult = ""
				return m, m.exportDataToJSON(m.lastQueryColumns, m.lastQueryRows, "")
			}
			// Export preview data to JSON
			if m.state == dataPreviewView && !m.isExporting {
				if len(m.lastPreviewColumns) == 0 || len(m.lastPreviewRows) == 0 {
					m.queryResult = "⚠️ No data to export. Load table preview first!"
					return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
						return clearResultMsg{}
					})
				}
				m.isExporting = true
				m.err = nil
				m.queryResult = ""
				return m, m.exportDataToJSON(m.lastPreviewColumns, m.lastPreviewRows, m.selectedTable)
			}

		case "ctrl+h":
			// Access query history
			if m.state == queryView && m.db != nil {
				m.state = queryHistoryView
				m.updateQueryHistoryList()
				return m, nil
			}

		case "p":
			// Don't trigger preview if we're in search mode
			if m.state == tablesView && m.db != nil && !m.isLoadingPreview && !m.isSearchingTables {
				// Preview table data
				if i, ok := m.tablesList.SelectedItem().(item); ok {
					m.selectedTable = i.title
					m.isLoadingPreview = true
					m.err = nil
					return m, m.loadDataPreview()
				}
			}

		case "i":
			// Show indexes and constraints view from columns view
			if m.state == columnsView && m.db != nil && m.selectedTable != "" && !m.isSearchingColumns {
				m.state = indexesView
				return m, m.loadIndexes()
			}

		case "e":
			if m.state == savedConnectionsView && len(m.savedConnections) > 0 {
				// Edit selected connection
				if i, ok := m.savedConnectionsList.SelectedItem().(item); ok {
					// Find connection by name
					for idx, conn := range m.savedConnections {
						if conn.Name == i.title {
							m.editingConnectionIdx = idx
							// Set up edit form with existing values
							conn := m.savedConnections[idx]
							m.nameInput.SetValue(conn.Name)
							m.textInput.SetValue(conn.ConnectionStr)
							// Set selected DB type
							for _, db := range dbTypes {
								if db.driver == conn.Driver {
									m.selectedDB = db
									break
								}
							}
							m.state = editConnectionView
							m.nameInput.Focus()
							return m, nil
						}
					}
				}
			}

		case "d":
			if m.state == savedConnectionsView && len(m.savedConnections) > 0 {
				// Delete selected connection
				if i, ok := m.savedConnectionsList.SelectedItem().(item); ok {
					// Find and remove connection by name
					for idx, conn := range m.savedConnections {
						if conn.Name == i.title {
							// Remove connection from slice
							m.savedConnections = append(m.savedConnections[:idx], m.savedConnections[idx+1:]...)
							// Save updated connections
							saveConnections(m.savedConnections)
							// Update the list
							m = m.updateSavedConnectionsList()
							return m, nil
						}
					}
				}
			}
			if m.state == queryHistoryView && len(m.queryHistory) > 0 {
				// Delete selected query from history
				selectedIndex := m.queryHistoryList.Index()
				if selectedIndex < len(m.queryHistory) {
					// Since we display newest first, reverse the index
					historyIndex := len(m.queryHistory) - 1 - selectedIndex

					// Remove query from slice
					m.queryHistory = append(m.queryHistory[:historyIndex], m.queryHistory[historyIndex+1:]...)

					// Save updated history
					go saveQueryHistory(m.queryHistory)

					// Update the list display
					m.updateQueryHistoryList()

					// Adjust cursor position if needed
					if selectedIndex >= len(m.queryHistory) && len(m.queryHistory) > 0 {
						// If we deleted the last item, move cursor to new last item
						m.queryHistoryList.Select(len(m.queryHistory) - 1)
					}

					return m, nil
				}
			}

		case "enter":
			switch m.state {
			case dbTypeView:
				if i, ok := m.dbTypeList.SelectedItem().(item); ok {
					for _, db := range dbTypes {
						if db.name == i.title {
							m.selectedDB = db
							break
						}
					}
					m.state = connectionView
					m.nameInput.SetValue("")
					m.textInput.SetValue("")
					m.textInput.Blur()
					m.nameInput.Focus()

					// Set placeholder according to DB type
					switch m.selectedDB.driver {
					case "postgres":
						m.textInput.Placeholder = "postgres://user:password@localhost/dbname?sslmode=disable"
					case "mysql":
						m.textInput.Placeholder = "user:password@tcp(localhost:3306)/dbname"
					case "sqlite3":
						m.textInput.Placeholder = "/path/to/database.db"
					}
				}

			case savedConnectionsView:
				if i, ok := m.savedConnectionsList.SelectedItem().(item); ok && !m.isConnecting {
					// Find saved connection by name
					for _, conn := range m.savedConnections {
						if conn.Name == i.title {
							// Set DB type based on driver
							for _, db := range dbTypes {
								if db.driver == conn.Driver {
									m.selectedDB = db
									break
								}
							}
							m.connectionStr = conn.ConnectionStr
							m.isConnecting = true
							m.err = nil
							return m, m.connectDB()
						}
					}
				}

			case saveConnectionView:
				name := m.nameInput.Value()
				if name != "" {
					// Save connection with provided name
					newConnection := SavedConnection{
						Name:          name,
						Driver:        m.selectedDB.driver,
						ConnectionStr: m.connectionStr,
					}
					m.savedConnections = append(m.savedConnections, newConnection)
					saveConnections(m.savedConnections)
					m.state = connectionView
					return m, nil
				}

			case editConnectionView:
				name := m.nameInput.Value()
				connectionStr := m.textInput.Value()
				if name != "" && connectionStr != "" && m.editingConnectionIdx >= 0 {
					// Update the existing connection
					m.savedConnections[m.editingConnectionIdx] = SavedConnection{
						Name:          name,
						Driver:        m.selectedDB.driver,
						ConnectionStr: connectionStr,
					}
					saveConnections(m.savedConnections)
					m = m.updateSavedConnectionsList()
					m.state = savedConnectionsView
					m.editingConnectionIdx = -1
					return m, nil
				}

			case schemaView:
				if i, ok := m.schemasList.SelectedItem().(item); ok {
					// Set selected schema and load tables for that schema
					m.selectedSchema = i.title
					m.isLoadingTables = true
					m.err = nil
					return m, m.loadTablesForSchema()
				}

			case tablesView:
				// Don't trigger table selection if we're in search mode
				if i, ok := m.tablesList.SelectedItem().(item); ok && !m.isLoadingColumns && !m.isSearchingTables {
					m.selectedTable = i.title
					m.isLoadingColumns = true
					m.err = nil
					return m, m.loadColumns()
				}

			case queryView:
				query := m.queryInput.Value()
				if query != "" && !m.isExecutingQuery {
					m.isExecutingQuery = true
					m.err = nil
					m.queryResult = ""
					return m, m.executeQuery(query)
				}

			case queryHistoryView:
				if _, ok := m.queryHistoryList.SelectedItem().(item); ok {
					// Get the original query from history
					selectedIndex := m.queryHistoryList.Index()
					if selectedIndex < len(m.queryHistory) {
						// Since we display newest first, reverse the index
						historyIndex := len(m.queryHistory) - 1 - selectedIndex
						selectedQuery := m.queryHistory[historyIndex].Query

						// Go back to query view and populate the input
						m.state = queryView
						m.queryInput.SetValue(selectedQuery)
						m.queryInput.Focus()
					}
				}

			case dataPreviewView:
				// Get selected row data and show row detail view
				if len(m.dataPreviewTable.Rows()) > 0 {
					selectedIndex := m.dataPreviewTable.Cursor()
					if selectedIndex < len(m.dataPreviewTable.Rows()) {
						m.selectedRowIndex = selectedIndex
						m.selectedRowData = m.dataPreviewTable.Rows()[selectedIndex]
						m.rowDetailCurrentPage = 0   // Reset to first page
						m.rowDetailSelectedField = 0 // Reset field selection
						m.isViewingFullText = false  // Reset full text view
						m.fullTextScrollOffset = 0   // Reset scroll position
						m.fullTextCurrentPage = 0    // Reset full text page
						m.fullTextSelectedField = 0  // Reset full text field selection
						m.state = rowDetailView
						return m, nil // Important: return to prevent further processing of the same key event
					}
				}
			}
		}
	}

	// Update components according to state
	switch m.state {
	case dbTypeView:
		m.dbTypeList, cmd = m.dbTypeList.Update(msg)
	case savedConnectionsView:
		m.savedConnectionsList, cmd = m.savedConnectionsList.Update(msg)
	case connectionView:
		// Handle focus between name and connection string inputs
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "tab":
				if m.nameInput.Focused() {
					m.nameInput.Blur()
					m.textInput.Focus()
				} else {
					m.textInput.Blur()
					m.nameInput.Focus()
				}
				return m, nil
			}
		}

		// Update the focused input
		if m.nameInput.Focused() {
			m.nameInput, cmd = m.nameInput.Update(msg)
		} else {
			m.textInput, cmd = m.textInput.Update(msg)
		}
	case saveConnectionView:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case editConnectionView:
		// Handle focus between name and connection string inputs
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "tab":
				if m.nameInput.Focused() {
					m.nameInput.Blur()
					m.textInput.Focus()
				} else {
					m.textInput.Blur()
					m.nameInput.Focus()
				}
				return m, nil
			}
		}

		// Update the focused input
		if m.nameInput.Focused() {
			m.nameInput, cmd = m.nameInput.Update(msg)
		} else {
			m.textInput, cmd = m.textInput.Update(msg)
		}
	case queryView:
		// Handle focus between query input and results table
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "tab":
				// Toggle focus between input and table
				if m.queryInput.Focused() && len(m.queryResultsTable.Rows()) > 0 {
					m.queryInput.Blur()
					m.queryResultsTable.Focus()
				} else if len(m.queryResultsTable.Rows()) > 0 {
					m.queryResultsTable.Blur()
					m.queryInput.Focus()
				}
				return m, nil
			case "up", "down", "j", "k":
				// If table has focus, let it handle navigation
				if len(m.queryResultsTable.Rows()) > 0 && !m.queryInput.Focused() {
					m.queryResultsTable, cmd = m.queryResultsTable.Update(msg)
					return m, cmd
				}
			}
		}

		// Update query input if it has focus or no table results
		if m.queryInput.Focused() || len(m.queryResultsTable.Rows()) == 0 {
			m.queryInput, cmd = m.queryInput.Update(msg)
		}
	case tablesView:
		if m.isSearchingTables {
			// Handle search input updates
			if msg, ok := msg.(tea.KeyMsg); ok {
				switch msg.String() {
				case "enter":
					// Finish searching and focus on filtered list
					m.isSearchingTables = false
					m.searchInput.Blur()
					return m, nil
				case "esc":
					// This is handled in the main esc case above
					// No need to handle here as it would be duplicate
				}
			}

			// Update search input and filter results
			m.searchInput, cmd = m.searchInput.Update(msg)

			// Apply filter when search term changes
			newSearchTerm := m.searchInput.Value()
			if newSearchTerm != m.searchTerm {
				m.searchTerm = newSearchTerm
				filteredItems := m.filterTableItems(newSearchTerm)
				m.tablesList.SetItems(filteredItems)
			}
		} else {
			// Normal table list navigation
			m.tablesList, cmd = m.tablesList.Update(msg)
		}
	case columnsView:
		if m.isSearchingColumns {
			// Handle search input updates
			if msg, ok := msg.(tea.KeyMsg); ok {
				switch msg.String() {
				case "enter":
					// Finish searching and focus on filtered table
					m.isSearchingColumns = false
					m.searchInput.Blur()
					return m, nil
				case "esc":
					// This is handled in the main esc case above
				}
			}

			// Update search input and filter results
			m.searchInput, cmd = m.searchInput.Update(msg)

			// Apply filter when search term changes
			newSearchTerm := m.searchInput.Value()
			if newSearchTerm != m.searchTerm {
				m.searchTerm = newSearchTerm
				filteredRows := m.filterColumnRows(newSearchTerm)
				m.columnsTable.SetRows(filteredRows)
			}
		} else {
			// Normal column table navigation
			m.columnsTable, cmd = m.columnsTable.Update(msg)
		}
	case schemaView:
		m.schemasList, cmd = m.schemasList.Update(msg)
	case queryHistoryView:
		m.queryHistoryList, cmd = m.queryHistoryList.Update(msg)
	case dataPreviewView:
		m.dataPreviewTable, cmd = m.dataPreviewTable.Update(msg)
	case indexesView:
		// Handle Enter key for index detail view
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "enter":
				// Get selected index/constraint details
				if selectedRow := m.indexesTable.SelectedRow(); len(selectedRow) >= 4 {
					m.selectedIndexName = selectedRow[0]
					m.selectedIndexType = selectedRow[1]
					m.selectedIndexColumns = selectedRow[2]
					m.selectedIndexDefinition = selectedRow[3]
					m.state = indexDetailView
					return m, nil
				}
			}
		}
		m.indexesTable, cmd = m.indexesTable.Update(msg)
	case rowDetailView:
		// Handle pagination and navigation keys for row detail view
		if msg, ok := msg.(tea.KeyMsg); ok {
			columns := m.dataPreviewTable.Columns()
			totalPages := (len(columns) + m.rowDetailItemsPerPage - 1) / m.rowDetailItemsPerPage

			if m.isViewingFullText {
				// In full text view, handle pagination and field navigation
				columns := m.dataPreviewTable.Columns()
				totalFullTextPages := (len(columns) + m.fullTextItemsPerPage - 1) / m.fullTextItemsPerPage

				switch msg.String() {
				case "esc":
					m.isViewingFullText = false
					m.fullTextScrollOffset = 0
					m.fullTextCurrentPage = 0
					m.fullTextSelectedField = 0
				case "right", "n":
					if m.fullTextCurrentPage < totalFullTextPages-1 {
						m.fullTextCurrentPage++
						m.fullTextSelectedField = 0
					}
				case "left", "p":
					if m.fullTextCurrentPage > 0 {
						m.fullTextCurrentPage--
						m.fullTextSelectedField = 0
					}
				case "home":
					m.fullTextCurrentPage = 0
					m.fullTextSelectedField = 0
				case "end":
					m.fullTextCurrentPage = totalFullTextPages - 1
					m.fullTextSelectedField = 0
				case "up", "k":
					if m.fullTextSelectedField > 0 {
						m.fullTextSelectedField--
					}
				case "down", "j":
					fieldsOnPage := m.fullTextItemsPerPage
					if m.fullTextCurrentPage == totalFullTextPages-1 {
						fieldsOnPage = len(columns) - (m.fullTextCurrentPage * m.fullTextItemsPerPage)
					}
					if m.fullTextSelectedField < fieldsOnPage-1 {
						m.fullTextSelectedField++
					}
				case "pgup":
					m.fullTextScrollOffset -= m.fullTextLinesPerPage
					if m.fullTextScrollOffset < 0 {
						m.fullTextScrollOffset = 0
					}
				case "pgdown":
					m.fullTextScrollOffset += m.fullTextLinesPerPage
				}
			} else {
				// In normal row detail view
				switch msg.String() {
				case "right", "n", "pgdown":
					if m.rowDetailCurrentPage < totalPages-1 {
						m.rowDetailCurrentPage++
						m.rowDetailSelectedField = 0 // Reset field selection
					}
				case "left", "p", "pgup":
					if m.rowDetailCurrentPage > 0 {
						m.rowDetailCurrentPage--
						m.rowDetailSelectedField = 0 // Reset field selection
					}
				case "home":
					m.rowDetailCurrentPage = 0
					m.rowDetailSelectedField = 0
				case "end":
					m.rowDetailCurrentPage = totalPages - 1
					m.rowDetailSelectedField = 0
				case "up", "k":
					if m.rowDetailSelectedField > 0 {
						m.rowDetailSelectedField--
					}
				case "down", "j":
					fieldsOnPage := m.rowDetailItemsPerPage
					if m.rowDetailCurrentPage == totalPages-1 {
						// Last page might have fewer items
						fieldsOnPage = len(columns) - (m.rowDetailCurrentPage * m.rowDetailItemsPerPage)
					}
					if m.rowDetailSelectedField < fieldsOnPage-1 {
						m.rowDetailSelectedField++
					}
				case "enter", "space":
					// Navigate to field detail view - get complete field value from database
					columns := m.dataPreviewTable.Columns()
					actualFieldIndex := m.rowDetailCurrentPage*m.rowDetailItemsPerPage + m.rowDetailSelectedField
					if actualFieldIndex < len(columns) {
						m.selectedFieldName = columns[actualFieldIndex].Title
						m.selectedFieldIndex = actualFieldIndex
						m.fieldDetailScrollOffset = 0
						m.fieldDetailHorizontalOffset = 0
						m.state = fieldDetailView
						// Fetch complete field value from database
						return m, m.loadCompleteFieldValue(columns[actualFieldIndex].Title)
					}
				}
			}
		}
	case fieldDetailView:
		// Handle navigation in field detail view
		if msg, ok := msg.(tea.KeyMsg); ok {
			lines := strings.Split(m.selectedFieldValue, "\n")
			totalLines := len(lines)

			switch msg.String() {
			case "left", "h":
				// Scroll horizontally left
				if m.fieldDetailHorizontalOffset > 0 {
					m.fieldDetailHorizontalOffset -= 10 // Scroll by 10 characters
					if m.fieldDetailHorizontalOffset < 0 {
						m.fieldDetailHorizontalOffset = 0
					}
				}
			case "right", "l":
				// Scroll horizontally right
				m.fieldDetailHorizontalOffset += 10 // Scroll by 10 characters
			case "shift+left":
				// Fast scroll left
				if m.fieldDetailHorizontalOffset > 0 {
					m.fieldDetailHorizontalOffset -= 50 // Scroll by 50 characters
					if m.fieldDetailHorizontalOffset < 0 {
						m.fieldDetailHorizontalOffset = 0
					}
				}
			case "shift+right":
				// Fast scroll right
				m.fieldDetailHorizontalOffset += 50 // Scroll by 50 characters
			case "ctrl+home":
				// Jump to beginning of line horizontally
				m.fieldDetailHorizontalOffset = 0
			case "ctrl+end":
				// Jump to end of line horizontally (show last part)
				lines := strings.Split(m.selectedFieldValue, "\n")
				maxLineLength := 0
				for _, line := range lines {
					if len(line) > maxLineLength {
						maxLineLength = len(line)
					}
				}
				if maxLineLength > m.fieldDetailCharsPerLine {
					m.fieldDetailHorizontalOffset = maxLineLength - m.fieldDetailCharsPerLine
				} else {
					m.fieldDetailHorizontalOffset = 0
				}
			case "c":
				// Copy complete field value to clipboard
				return m, m.copyToClipboard(m.selectedFieldValue)
			}

			// Handle vertical scrolling for multi-line content
			if totalLines > 1 {
				switch msg.String() {
				case "up", "k":
					if m.fieldDetailScrollOffset > 0 {
						m.fieldDetailScrollOffset--
					}
				case "down", "j":
					maxOffset := totalLines - m.fieldDetailLinesPerPage
					if maxOffset < 0 {
						maxOffset = 0
					}
					if m.fieldDetailScrollOffset < maxOffset {
						m.fieldDetailScrollOffset++
					}
				case "pgup":
					m.fieldDetailScrollOffset -= m.fieldDetailLinesPerPage
					if m.fieldDetailScrollOffset < 0 {
						m.fieldDetailScrollOffset = 0
					}
				case "pgdown":
					maxOffset := totalLines - m.fieldDetailLinesPerPage
					if maxOffset < 0 {
						maxOffset = 0
					}
					m.fieldDetailScrollOffset += m.fieldDetailLinesPerPage
					if m.fieldDetailScrollOffset > maxOffset {
						m.fieldDetailScrollOffset = maxOffset
					}
				case "home":
					m.fieldDetailScrollOffset = 0
				case "end":
					maxOffset := totalLines - m.fieldDetailLinesPerPage
					if maxOffset < 0 {
						maxOffset = 0
					}
					m.fieldDetailScrollOffset = maxOffset
				}
			}
		}
	default:
		// Handle spinner messages
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, cmd
}

func (m model) View() string {
	switch m.state {
	case dbTypeView:
		return m.dbTypeView()
	case savedConnectionsView:
		return m.savedConnectionsView()
	case connectionView:
		return m.connectionView()
	case saveConnectionView:
		return m.saveConnectionView()
	case editConnectionView:
		return m.editConnectionView()
	case schemaView:
		return m.schemaView()
	case tablesView:
		return m.tablesView()
	case columnsView:
		return m.columnsView()
	case queryView:
		return m.queryView()
	case queryHistoryView:
		return m.queryHistoryView()
	case dataPreviewView:
		return m.dataPreviewView()
	case rowDetailView:
		return m.rowDetailView()
	case fieldDetailView:
		return m.fieldDetailView()
	case indexesView:
		return m.indexesView()
	case indexDetailView:
		return m.indexDetailView()
	}
	return ""
}

func (m model) dbTypeView() string {
	// Just use the list with its built-in title, add minimal help below
	content := m.dbTypeList.View()

	// Simple help text
	helpText := helpStyle.Render(
		keyStyle.Render("enter") + ": select • " +
			keyStyle.Render("s") + ": saved connections • " +
			keyStyle.Render("q") + ": quit",
	)

	return docStyle.Render(content + "\n" + helpText)
}

func (m model) savedConnectionsView() string {
	var content string

	if m.isConnecting {
		loadingMsg := m.getLoadingText("Connecting to saved connection...")
		content = m.savedConnectionsList.View() + "\n" + loadingMsg
	} else if len(m.savedConnections) == 0 {
		emptyMsg := infoStyle.Render("📝 No saved connections yet.\n\nGo back and create your first connection!")
		content = m.savedConnectionsList.View() + "\n" + emptyMsg
	} else {
		content = m.savedConnectionsList.View()
	}

	helpText := helpStyle.Render(
		keyStyle.Render("enter") + ": connect • " +
			keyStyle.Render("e") + ": edit • " +
			keyStyle.Render("d") + ": delete • " +
			keyStyle.Render("esc") + ": back",
	)

	return docStyle.Render(content + "\n" + helpText)
}

func (m model) saveConnectionView() string {
	content := titleStyle.Render("Save Connection") + "\n\n"
	content += "Name for this connection:\n"
	content += m.nameInput.View() + "\n\n"
	content += "Connection to save:\n"
	content += helpStyle.Render(fmt.Sprintf("%s: %s", m.selectedDB.name, m.connectionStr))
	content += "\n\n" + helpStyle.Render("enter: save • esc: cancel")
	return docStyle.Render(content)
}

func (m model) editConnectionView() string {
	content := titleStyle.Render("Edit Connection") + "\n\n"

	if m.err != nil {
		content += errorStyle.Render("❌ Error: "+m.err.Error()) + "\n\n"
	}

	content += "Connection name:\n"
	content += m.nameInput.View() + "\n\n"

	content += fmt.Sprintf("Database type: %s\n", m.selectedDB.name)
	content += "Connection string:\n"
	content += m.textInput.View() + "\n\n"

	content += "Examples:\n"
	switch m.selectedDB.driver {
	case "postgres":
		content += helpStyle.Render("postgres://user:password@localhost/dbname?sslmode=disable")
	case "mysql":
		content += helpStyle.Render("user:password@tcp(localhost:3306)/dbname")
	case "sqlite3":
		content += helpStyle.Render("./database.db or /path/to/database.db")
	}

	content += "\n\n" + helpStyle.Render("enter: save changes • tab: switch fields • esc: cancel")
	return docStyle.Render(content)
}

func (m model) schemaView() string {
	var content string

	if m.isLoadingSchemas {
		loadingMsg := m.getLoadingText("Loading schemas...")
		content = m.schemasList.View() + "\n" + loadingMsg
	} else if len(m.schemas) == 0 {
		emptyMsg := infoStyle.Render("🗂️ No additional schemas found.\n\nUsing default schema.")
		content = m.schemasList.View() + "\n" + emptyMsg
	} else {
		content = m.schemasList.View()
	}

	helpText := helpStyle.Render(
		keyStyle.Render("enter") + ": select schema • " +
			keyStyle.Render("esc") + ": back",
	)

	return docStyle.Render(content + "\n" + helpText)
}

func (m model) connectionView() string {
	// Database icon based on type
	var dbIcon string
	switch m.selectedDB.driver {
	case "postgres":
		dbIcon = "🐘"
	case "mysql":
		dbIcon = "🐬"
	case "sqlite3":
		dbIcon = "📁"
	default:
		dbIcon = "🗄️"
	}

	title := titleStyle.Render(fmt.Sprintf("%s  Connect to %s", dbIcon, m.selectedDB.name))

	var messageContent string
	if m.isTestingConnection {
		messageContent = m.getLoadingText("Testing connection...")
	} else if m.isSavingConnection {
		messageContent = m.getLoadingText("Validating and saving connection...")
	} else if m.isConnecting {
		messageContent = m.getLoadingText("Connecting to database...")
	} else if m.err != nil {
		messageContent = errorStyle.Render("❌ " + m.err.Error())
	} else if m.queryResult != "" {
		messageContent = successStyle.Render(m.queryResult)
	}

	// Input fields with enhanced styling
	nameLabel := subtitleStyle.Render("Connection Name:")
	var nameField string
	if m.nameInput.Focused() {
		nameField = inputFocusedStyle.Render(m.nameInput.View())
	} else {
		nameField = inputStyle.Render(m.nameInput.View())
	}

	connLabel := subtitleStyle.Render("Connection String:")
	var connField string
	if m.textInput.Focused() {
		connField = inputFocusedStyle.Render(m.textInput.View())
	} else {
		connField = inputStyle.Render(m.textInput.View())
	}

	// Examples in an info box
	var exampleText string
	switch m.selectedDB.driver {
	case "postgres":
		exampleText = "postgres://user:password@localhost/dbname?sslmode=disable"
	case "mysql":
		exampleText = "user:password@tcp(localhost:3306)/dbname"
	case "sqlite3":
		exampleText = "./database.db or /path/to/database.db"
	}

	examples := infoStyle.Render(
		subtitleStyle.Render("Examples:") + "\n" + exampleText,
	)

	// Help text with enhanced key styling
	helpText := helpStyle.Render(
		keyStyle.Render("F1") + ": test connection • " +
			keyStyle.Render("F2") + ": validate, save & connect • " +
			keyStyle.Render("Tab") + ": switch fields • " +
			keyStyle.Render("Esc") + ": back",
	)

	// Assemble content with proper spacing
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
	return docStyle.Render(content)
}

func (m model) tablesView() string {
	var content string

	if m.isSearchingTables {
		// Search mode - simplified with lots of spacing to ensure visibility
		searchLabel := "🔍 Search Tables:"
		var searchField string
		if m.searchInput.Focused() {
			searchField = inputFocusedStyle.Render(m.searchInput.View())
		} else {
			searchField = inputStyle.Render(m.searchInput.View())
		}

		// Show search results count
		searchCount := fmt.Sprintf("Showing %d of %d tables",
			len(m.tablesList.Items()), len(m.originalTableItems))

		// Simple string concatenation with lots of spacing
		content = "\n\n\n\n\n" + // 5 empty lines at top
			searchLabel + "\n" +
			searchField + "\n" +
			searchCount + "\n\n" +
			m.tablesList.View()
	} else {
		// Normal view logic - status messages + table list
		var elements []string

		if m.isLoadingColumns {
			loadingMsg := m.getLoadingText("Loading table columns...")
			elements = append(elements, m.tablesList.View())
			elements = append(elements, loadingMsg)
		} else if m.isLoadingPreview {
			loadingMsg := m.getLoadingText("Loading table data preview...")
			elements = append(elements, m.tablesList.View())
			elements = append(elements, loadingMsg)
		} else if m.queryResult != "" {
			// Show query result message (like view definitions)
			elements = append(elements, m.tablesList.View())
			elements = append(elements, successStyle.Render(m.queryResult))
		} else if len(m.tables) == 0 {
			emptyMsg := infoStyle.Render("📋 No tables found in this database.\n\nThe database might be empty or you might not have sufficient permissions.")
			elements = append(elements, m.tablesList.View())
			elements = append(elements, emptyMsg)
		} else {
			statusText := successStyle.Render(fmt.Sprintf("✅ Connected successfully (%d tables found)", len(m.tables)))
			elements = append(elements, statusText)
			elements = append(elements, "") // Empty line
			elements = append(elements, m.tablesList.View())
		}

		content = lipgloss.JoinVertical(lipgloss.Left, elements...)
	}

	// Update help text based on search mode
	var helpText string
	if m.isSearchingTables {
		helpText = helpStyle.Render(
			keyStyle.Render("enter") + ": finish search • " +
				keyStyle.Render("esc") + ": cancel search")
	} else {
		helpText = helpStyle.Render(
			keyStyle.Render("enter") + ": view columns • " +
				keyStyle.Render("p") + ": preview data • " +
				keyStyle.Render("r") + ": run query • " +
				keyStyle.Render("Ctrl+F") + ": search • " +
				keyStyle.Render("esc") + ": disconnect")
	}

	return docStyle.Render(content + "\n" + helpText)
}

func (m model) columnsView() string {
	content := titleStyle.Render(fmt.Sprintf("Columns of table: %s", m.selectedTable)) + "\n\n"

	// Add search input if in search mode
	var searchElements []string
	if m.isSearchingColumns {
		searchLabel := subtitleStyle.Render("🔍 Search Columns:")
		var searchField string
		if m.searchInput.Focused() {
			searchField = inputFocusedStyle.Render(m.searchInput.View())
		} else {
			searchField = inputStyle.Render(m.searchInput.View())
		}

		// Show search results count
		var searchInfo string
		if len(m.originalTableRows) > 0 {
			searchInfo = infoStyle.Render(fmt.Sprintf("Showing %d of %d columns",
				len(m.columnsTable.Rows()), len(m.originalTableRows)))
		}

		searchElements = append(searchElements, searchLabel)
		searchElements = append(searchElements, searchField)
		if searchInfo != "" {
			searchElements = append(searchElements, searchInfo)
		}
		searchElements = append(searchElements, "") // Empty line separator

		content += lipgloss.JoinVertical(lipgloss.Left, searchElements...)
	}

	content += m.columnsTable.View()

	// Update help text based on search mode
	var helpText string
	if m.isSearchingColumns {
		helpText = helpStyle.Render(
			keyStyle.Render("enter") + ": finish search • " +
				keyStyle.Render("esc") + ": cancel search")
	} else {
		helpText = helpStyle.Render(
			keyStyle.Render("↑/↓") + ": navigate • " +
				keyStyle.Render("r") + ": run query • " +
				keyStyle.Render("i") + ": indexes & constraints • " +
				keyStyle.Render("Ctrl+F") + ": search • " +
				keyStyle.Render("s") + ": save connection • " +
				keyStyle.Render("esc") + ": back to tables")
	}

	content += "\n" + helpText
	return docStyle.Render(content)
}

func (m model) indexesView() string {
	content := titleStyle.Render(fmt.Sprintf("🔑 Indexes & Constraints: %s", m.selectedTable)) + "\n\n"

	if m.err != nil {
		content += errorStyle.Render("❌ "+m.err.Error()) + "\n\n"
	}

	// Show the indexes table
	content += m.indexesTable.View() + "\n\n"

	// Help text
	helpText := helpStyle.Render(
		keyStyle.Render("↑/↓") + ": navigate • " +
			keyStyle.Render("enter") + ": view details • " +
			keyStyle.Render("esc") + ": back to columns")

	content += "\n" + helpText
	return docStyle.Render(content)
}

func (m model) indexDetailView() string {
	if m.selectedIndexName == "" {
		return "No index data available\n\nesc: back to indexes"
	}

	// Title with icon based on type
	icon := "🔑"
	if m.selectedIndexType == "PRIMARY" {
		icon = "🗝️"
	} else if m.selectedIndexType == "UNIQUE" {
		icon = "✨"
	} else if strings.Contains(m.selectedIndexType, "FOREIGN") {
		icon = "🔗"
	}

	title := titleStyle.Render(fmt.Sprintf("%s Index/Constraint Details", icon))
	content := title + "\n\n"

	// Plain text styling with colors but no borders or boxes
	labelStyle := lipgloss.NewStyle().Foreground(darkMagenta).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))

	// Simple, clean layout with plain text
	content += labelStyle.Render("Name: ") + valueStyle.Render(m.selectedIndexName) + "\n"
	content += labelStyle.Render("Type: ") + valueStyle.Render(m.selectedIndexType) + "\n"
	content += labelStyle.Render("Columns: ") + valueStyle.Render(m.selectedIndexColumns) + "\n"
	content += labelStyle.Render("Definition: ") + valueStyle.Render(m.selectedIndexDefinition) + "\n\n"

	// Simple help text without boxes
	content += labelStyle.Render("esc") + ": back to indexes"
	
	return docStyle.Render(content)
}

func (m model) queryView() string {
	title := titleStyle.Render("⚡  SQL Query Runner")

	var messageContent string
	if m.isExecutingQuery {
		messageContent = m.getLoadingText("Executing query...")
	} else if m.isExporting {
		messageContent = m.getLoadingText("Exporting data...")
	} else if m.err != nil {
		messageContent = errorStyle.Render("❌ " + m.err.Error())
	}

	// Query input with enhanced styling
	queryLabel := subtitleStyle.Render("💻 Enter SQL Query:")
	var queryField string
	if m.queryInput.Focused() {
		queryField = inputFocusedStyle.Render(m.queryInput.View())
	} else {
		queryField = inputStyle.Render(m.queryInput.View())
	}

	var resultContent string
	if m.queryResult != "" {
		resultLabel := subtitleStyle.Render("📊 Query Result:")
		resultText := successStyle.Render(m.queryResult)

		// Only show the table if it has both columns and rows, and they match
		if len(m.queryResultsTable.Columns()) > 0 && len(m.queryResultsTable.Rows()) > 0 {
			tableContent := cardStyle.Render(m.queryResultsTable.View())
			resultContent = lipgloss.JoinVertical(lipgloss.Left, resultLabel, resultText, tableContent)
		} else {
			resultContent = lipgloss.JoinVertical(lipgloss.Left, resultLabel, resultText)
		}
	}

	// Examples in an info box
	examples := infoStyle.Render(
		subtitleStyle.Render("💡 Examples:") + "\n" +
			keyStyle.Render("SELECT") + " * FROM users LIMIT 10;\n" +
			keyStyle.Render("INSERT") + " INTO users (name, email) VALUES ('John', 'john@example.com');\n" +
			keyStyle.Render("UPDATE") + " users SET email = 'new@example.com' WHERE id = 1;\n" +
			keyStyle.Render("DELETE") + " FROM users WHERE id = 1;",
	)

	// Help text with enhanced key styling
	helpText := helpStyle.Render(
		keyStyle.Render("Enter") + ": execute query • " +
			keyStyle.Render("Tab") + ": switch focus • " +
			keyStyle.Render("↑/↓") + ": navigate results • " +
			keyStyle.Render("Ctrl+E") + ": export CSV • " +
			keyStyle.Render("Ctrl+J") + ": export JSON • " +
			keyStyle.Render("Esc") + ": back to tables",
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
	return docStyle.Render(content)
}

func (m model) dataPreviewView() string {
	content := titleStyle.Render(fmt.Sprintf("Data Preview: %s", m.selectedTable)) + "\n\n"

	// Show status messages (loading, success, error)
	if m.isExporting {
		content += m.getLoadingText("Exporting data...") + "\n\n"
	} else if m.err != nil {
		content += errorStyle.Render("❌ Error: "+m.err.Error()) + "\n\n"
	} else if m.queryResult != "" {
		content += successStyle.Render(m.queryResult) + "\n\n"
	}

	// Only show the table if it has both columns and rows, and they match
	if len(m.dataPreviewTable.Columns()) > 0 && len(m.dataPreviewTable.Rows()) > 0 {
		content += infoStyle.Render(fmt.Sprintf("Showing first 10 rows from %s", m.selectedTable)) + "\n\n"
		content += m.dataPreviewTable.View()
	} else if m.err == nil && m.queryResult == "" && !m.isExporting {
		content += helpStyle.Render("Loading data preview...")
	}

	content += "\n\n" + helpStyle.Render("↑/↓: navigate rows • "+keyStyle.Render("enter")+": view row details • "+keyStyle.Render("Ctrl+E")+": export CSV • "+keyStyle.Render("Ctrl+J")+": export JSON • esc: back to tables")
	return docStyle.Render(content)
}

func (m model) queryHistoryView() string {
	var content string

	if len(m.queryHistory) == 0 {
		emptyMsg := infoStyle.Render("📝 No query history yet.\n\nExecute some queries first!")
		content = m.queryHistoryList.View() + "\n" + emptyMsg
	} else {
		content = m.queryHistoryList.View()
	}

	helpText := helpStyle.Render(
		keyStyle.Render("enter") + ": use query • " +
			keyStyle.Render("d") + ": delete • " +
			keyStyle.Render("esc") + ": back",
	)

	return docStyle.Render(content + "\n" + helpText)
}

func (m model) rowDetailView() string {
	if len(m.selectedRowData) == 0 {
		return docStyle.Render("No row data available")
	}

	// Get column names from the data preview table
	columns := m.dataPreviewTable.Columns()

	// If viewing full text of a field
	if m.isViewingFullText {
		return m.fullRowDataView()
	}

	// Calculate pagination
	totalFields := len(columns)
	totalPages := (totalFields + m.rowDetailItemsPerPage - 1) / m.rowDetailItemsPerPage
	startIndex := m.rowDetailCurrentPage * m.rowDetailItemsPerPage
	endIndex := startIndex + m.rowDetailItemsPerPage
	if endIndex > totalFields {
		endIndex = totalFields
	}

	// Build title with pagination info
	pageInfo := ""
	if totalPages > 1 {
		pageInfo = fmt.Sprintf(" (Page %d of %d)", m.rowDetailCurrentPage+1, totalPages)
	}
	title := titleStyle.Render(fmt.Sprintf("Row Details - %s (Row %d)%s", m.selectedTable, m.selectedRowIndex+1, pageInfo))

	// Build the row detail content for current page
	var details []string
	for i := startIndex; i < endIndex; i++ {
		if i < len(columns) && i < len(m.selectedRowData) {
			col := columns[i]
			value := m.selectedRowData[i]

			// Check if this field is selected
			isSelected := (i - startIndex) == m.rowDetailSelectedField

			// Handle empty values
			displayValue := value
			if value == "" {
				displayValue = lipgloss.NewStyle().Foreground(lightGray).Render("(empty)")
			} else if strings.TrimSpace(value) == "" {
				displayValue = lipgloss.NewStyle().Foreground(lightGray).Render("(whitespace)")
			}

			// Smart truncation for better UX
			maxValueLength := 80 // Reduced for better pagination view
			truncated := false
			originalLength := len(value)

			// Check if value is large (show preview only)
			if originalLength > maxValueLength {
				// For single line content, show preview
				if !strings.Contains(value, "\n") {
					displayValue = value[:maxValueLength] + "..."
					truncated = true
				} else {
					// For multi-line content, show first line + indicator
					lines := strings.Split(value, "\n")
					if len(lines[0]) > maxValueLength {
						displayValue = lines[0][:maxValueLength] + "..."
					} else {
						displayValue = lines[0]
					}
					if len(lines) > 1 {
						displayValue += "\n" + lipgloss.NewStyle().Foreground(lightGray).Render(fmt.Sprintf("... (+%d more lines)", len(lines)-1))
					}
					truncated = true
				}
			} else if strings.Contains(displayValue, "\n") {
				// Handle smaller multi-line values
				lines := strings.Split(displayValue, "\n")
				if len(lines) > 2 {
					// Show first 2 lines and indicate more
					displayValue = strings.Join(lines[:2], "\n") + "\n" +
						lipgloss.NewStyle().Foreground(lightGray).Render(fmt.Sprintf("... (+%d more lines)", len(lines)-2))
					truncated = true
				}
			}

			// Format field name and value with better styling
			fieldNameStyle := lipgloss.NewStyle().Foreground(accentMagenta).Bold(true)
			if isSelected {
				fieldNameStyle = fieldNameStyle.Background(primaryMagenta).Foreground(white)
			}
			fieldName := fieldNameStyle.Render(fmt.Sprintf("%-20s", col.Title))

			fieldValueStyle := lipgloss.NewStyle().Foreground(white)
			if isSelected {
				fieldValueStyle = fieldValueStyle.Background(lightMagenta).Foreground(darkGray)
			}
			fieldValue := fieldValueStyle.Render(displayValue)

			// Add size and truncation indicators
			sizeIndicator := ""
			if originalLength > 0 {
				if originalLength > 1024*1024 { // 1MB
					sizeIndicator = fmt.Sprintf(" %s", lipgloss.NewStyle().Foreground(warningOrange).Render(fmt.Sprintf("[%.1fMB]", float64(originalLength)/(1024*1024))))
				} else if originalLength > 1024 { // 1KB
					sizeIndicator = fmt.Sprintf(" %s", lipgloss.NewStyle().Foreground(accentMagenta).Render(fmt.Sprintf("[%.1fKB]", float64(originalLength)/1024)))
				} else if originalLength > maxValueLength {
					sizeIndicator = fmt.Sprintf(" %s", lipgloss.NewStyle().Foreground(lightGray).Render(fmt.Sprintf("[%d chars]", originalLength)))
				}
			}

			enterHint := ""
			if truncated || originalLength > 50 {
				enterHint = " " + lipgloss.NewStyle().Foreground(lightGray).Render("(press enter to view full)")
			}

			fieldContent := fmt.Sprintf("%s │ %s%s%s", fieldName, fieldValue, sizeIndicator, enterHint)
			details = append(details, fieldContent)
		}
	}

	// Add separator between fields
	detailContent := strings.Join(details, "\n"+strings.Repeat("─", 100)+"\n")

	// Create card style for better presentation
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryMagenta).
		Padding(1, 2).
		Margin(1, 0)

	// Wrap content in card
	contentCard := cardStyle.Render(detailContent)

	// Build navigation help text
	var navHelp []string
	navHelp = append(navHelp,
		keyStyle.Render("↑/↓")+" or "+keyStyle.Render("j/k")+": select field",
		keyStyle.Render("enter/space")+": view field details",
	)
	if totalPages > 1 {
		navHelp = append(navHelp,
			keyStyle.Render("←/→")+" or "+keyStyle.Render("p/n")+": navigate pages",
			keyStyle.Render("home/end")+": first/last page",
		)
	}
	navHelp = append(navHelp, keyStyle.Render("esc")+": back to data preview")

	helpText := helpStyle.Render(strings.Join(navHelp, " • "))

	// Field summary
	fieldSummary := infoStyle.Render(fmt.Sprintf("Showing fields %d-%d of %d", startIndex+1, endIndex, totalFields))

	// Combine all elements
	content := title + "\n\n" + fieldSummary + "\n\n" + contentCard

	return docStyle.Render(content + "\n\n" + helpText)
}

// Helper function to wrap text to specified width
func (m model) wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	currentLine := words[0]
	for i := 1; i < len(words); i++ {
		if len(currentLine)+len(words[i])+1 <= width {
			currentLine += " " + words[i]
		} else {
			result.WriteString(currentLine + "\n")
			currentLine = words[i]
		}
	}
	result.WriteString(currentLine)
	return result.String()
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to get maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Function to display complete row data by querying database directly
func (m model) fullRowDataView() string {
	// Build query to get the complete row data
	var query string
	switch m.selectedDB.driver {
	case "postgres":
		if m.selectedSchema != "" {
			query = fmt.Sprintf(`SELECT * FROM "%s"."%s" LIMIT %d OFFSET %d`,
				m.selectedSchema, m.selectedTable, 1, m.selectedRowIndex)
		} else {
			query = fmt.Sprintf(`SELECT * FROM "%s" LIMIT %d OFFSET %d`,
				m.selectedTable, 1, m.selectedRowIndex)
		}
	case "mysql":
		query = fmt.Sprintf("SELECT * FROM `%s` LIMIT %d OFFSET %d",
			m.selectedTable, 1, m.selectedRowIndex)
	case "sqlite3":
		query = fmt.Sprintf(`SELECT * FROM "%s" LIMIT %d OFFSET %d`,
			m.selectedTable, 1, m.selectedRowIndex)
	default:
		query = fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d",
			m.selectedTable, 1, m.selectedRowIndex)
	}

	// Execute query to get complete row data
	rows, err := m.db.Query(query)
	if err != nil {
		return fmt.Sprintf("Complete Row Data - %s (Row %d)\n\nError: %v\n\nesc: back to row details",
			m.selectedTable, m.selectedRowIndex+1, err)
	}
	defer rows.Close()

	// Get column names
	columnNames, err := rows.Columns()
	if err != nil {
		return fmt.Sprintf("Complete Row Data - %s (Row %d)\n\nError getting columns: %v\n\nesc: back to row details",
			m.selectedTable, m.selectedRowIndex+1, err)
	}

	// Get the row data
	if !rows.Next() {
		return fmt.Sprintf("Complete Row Data - %s (Row %d)\n\nNo data found\n\nesc: back to row details",
			m.selectedTable, m.selectedRowIndex+1)
	}

	// Create slice to hold values
	values := make([]interface{}, len(columnNames))
	valuePtrs := make([]interface{}, len(columnNames))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	err = rows.Scan(valuePtrs...)
	if err != nil {
		return fmt.Sprintf("Complete Row Data - %s (Row %d)\n\nError reading row: %v\n\nesc: back to row details",
			m.selectedTable, m.selectedRowIndex+1, err)
	}

	// Calculate pagination for full text view
	totalFields := len(columnNames)
	totalPages := (totalFields + m.fullTextItemsPerPage - 1) / m.fullTextItemsPerPage
	startIndex := m.fullTextCurrentPage * m.fullTextItemsPerPage
	endIndex := startIndex + m.fullTextItemsPerPage
	if endIndex > totalFields {
		endIndex = totalFields
	}

	// Build title with pagination info
	pageInfo := ""
	if totalPages > 1 {
		pageInfo = fmt.Sprintf(" (Page %d of %d)", m.fullTextCurrentPage+1, totalPages)
	}
	title := titleStyle.Render(fmt.Sprintf("Complete Row Data - %s (Row %d)%s",
		m.selectedTable, m.selectedRowIndex+1, pageInfo))

	// Build the display content for current page
	var fieldDetails []string
	for i := startIndex; i < endIndex; i++ {
		if i < len(columnNames) {
			columnName := columnNames[i]
			var valueStr string
			if values[i] == nil {
				valueStr = lipgloss.NewStyle().Foreground(lightGray).Render("NULL")
			} else {
				valueStr = fmt.Sprintf("%v", values[i])
			}

			if valueStr == "" {
				valueStr = lipgloss.NewStyle().Foreground(lightGray).Render("(empty)")
			} else if strings.TrimSpace(valueStr) == "" {
				valueStr = lipgloss.NewStyle().Foreground(lightGray).Render("(whitespace)")
			}

			// Check if this field is selected
			isSelected := (i - startIndex) == m.fullTextSelectedField

			// Format field name with better styling
			fieldNameStyle := lipgloss.NewStyle().Foreground(accentMagenta).Bold(true)
			if isSelected {
				fieldNameStyle = fieldNameStyle.Background(primaryMagenta).Foreground(white)
			}
			fieldName := fieldNameStyle.Render(fmt.Sprintf("%-20s", columnName))

			// Handle scrolling for selected field
			displayValue := valueStr
			if isSelected && len(valueStr) > 0 {
				// Apply scrolling to the selected field's value
				lines := strings.Split(valueStr, "\n")
				if m.fullTextScrollOffset < len(lines) {
					maxLines := min(m.fullTextLinesPerPage, len(lines)-m.fullTextScrollOffset)
					if maxLines > 0 {
						displayValue = strings.Join(lines[m.fullTextScrollOffset:m.fullTextScrollOffset+maxLines], "\n")
					}
				}

				// Add scroll indicators
				if m.fullTextScrollOffset > 0 {
					displayValue = lipgloss.NewStyle().Foreground(warningOrange).Render("↑ (scroll up)") + "\n" + displayValue
				}
				if m.fullTextScrollOffset+m.fullTextLinesPerPage < len(lines) {
					displayValue = displayValue + "\n" + lipgloss.NewStyle().Foreground(warningOrange).Render("↓ (scroll down)")
				}
			}

			// Format field value with styling
			fieldValueStyle := lipgloss.NewStyle().Foreground(white)
			if isSelected {
				fieldValueStyle = fieldValueStyle.Background(lightMagenta).Foreground(darkGray)
			}
			fieldValue := fieldValueStyle.Render(displayValue)

			// Combine field name and value
			fieldContent := fmt.Sprintf("%s │ %s", fieldName, fieldValue)
			fieldDetails = append(fieldDetails, fieldContent)
		}
	}

	// Join fields with separators
	detailContent := strings.Join(fieldDetails, "\n"+strings.Repeat("─", 120)+"\n")

	// Create card style for better presentation
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryMagenta).
		Padding(1, 2).
		Margin(1, 0)

	// Wrap content in card
	contentCard := cardStyle.Render(detailContent)

	// Build navigation help text
	var navHelp []string
	navHelp = append(navHelp,
		keyStyle.Render("↑/↓")+" or "+keyStyle.Render("j/k")+": select field",
		keyStyle.Render("pgup/pgdown")+": scroll field content",
	)
	if totalPages > 1 {
		navHelp = append(navHelp,
			keyStyle.Render("←/→")+" or "+keyStyle.Render("p/n")+": navigate pages",
			keyStyle.Render("home/end")+": first/last page",
		)
	}
	navHelp = append(navHelp, keyStyle.Render("esc")+": back to row details")

	helpText := helpStyle.Render(strings.Join(navHelp, " • "))

	// Field summary
	fieldSummary := infoStyle.Render(fmt.Sprintf("Showing fields %d-%d of %d",
		startIndex+1, endIndex, totalFields))

	// Selected field info
	selectedFieldInfo := ""
	if startIndex+m.fullTextSelectedField < len(columnNames) {
		selectedFieldName := columnNames[startIndex+m.fullTextSelectedField]
		selectedFieldInfo = infoStyle.Render(fmt.Sprintf("Selected: %s", selectedFieldName))
	}

	// Combine all elements
	content := title + "\n\n" + fieldSummary + "\n" + selectedFieldInfo + "\n\n" + contentCard

	return docStyle.Render(content + "\n\n" + helpText)
}

// fieldDetailView displays a single field's complete content with scrolling
func (m model) fieldDetailView() string {
	if m.selectedFieldName == "" || m.selectedFieldValue == "" {
		return "No field data available\n\nesc: back to row details"
	}

	// Simple title
	title := fmt.Sprintf("Field: %s", m.selectedFieldName)

	// Content statistics and position
	valueLength := len(m.selectedFieldValue)
	lines := strings.Split(m.selectedFieldValue, "\n")
	lineCount := len(lines)

	// Get max line length for horizontal scrolling
	maxLineLength := 0
	for _, line := range lines {
		if len(line) > maxLineLength {
			maxLineLength = len(line)
		}
	}

	// Handle content display with scrolling
	var visibleContent string

	if lineCount == 1 {
		// Single line - apply horizontal scrolling
		content := m.selectedFieldValue
		if len(content) > m.fieldDetailCharsPerLine {
			startChar := m.fieldDetailHorizontalOffset
			endChar := startChar + m.fieldDetailCharsPerLine
			if startChar >= len(content) {
				startChar = 0
				m.fieldDetailHorizontalOffset = 0
			}
			if endChar > len(content) {
				endChar = len(content)
			}
			visibleContent = content[startChar:endChar]
		} else {
			visibleContent = content
		}
	} else {
		// Multi-line - handle both vertical and horizontal scrolling
		visibleLines := m.fieldDetailLinesPerPage
		startLine := m.fieldDetailScrollOffset
		endLine := startLine + visibleLines

		if endLine > lineCount {
			endLine = lineCount
		}
		if startLine >= lineCount {
			startLine = 0
		}

		// Get visible lines and apply horizontal scrolling
		visibleLineSlice := lines[startLine:endLine]
		var processedLines []string
		for _, line := range visibleLineSlice {
			if len(line) > m.fieldDetailCharsPerLine && m.fieldDetailHorizontalOffset < len(line) {
				startChar := m.fieldDetailHorizontalOffset
				endChar := startChar + m.fieldDetailCharsPerLine
				if endChar > len(line) {
					endChar = len(line)
				}
				processedLines = append(processedLines, line[startChar:endChar])
			} else if m.fieldDetailHorizontalOffset < len(line) {
				processedLines = append(processedLines, line[m.fieldDetailHorizontalOffset:])
			} else {
				processedLines = append(processedLines, "")
			}
		}
		visibleContent = strings.Join(processedLines, "\n")
	}

	// Build status line
	var statusParts []string

	// Size
	if valueLength > 1024*1024 {
		statusParts = append(statusParts, fmt.Sprintf("%.1fMB", float64(valueLength)/(1024*1024)))
	} else if valueLength > 1024 {
		statusParts = append(statusParts, fmt.Sprintf("%.1fKB", float64(valueLength)/1024))
	} else {
		statusParts = append(statusParts, fmt.Sprintf("%d chars", valueLength))
	}

	// Lines
	if lineCount > 1 {
		statusParts = append(statusParts, fmt.Sprintf("%d lines", lineCount))
	}

	// Position indicators
	if lineCount > m.fieldDetailLinesPerPage {
		startLine := m.fieldDetailScrollOffset
		endLine := startLine + m.fieldDetailLinesPerPage
		if endLine > lineCount {
			endLine = lineCount
		}
		statusParts = append(statusParts, fmt.Sprintf("lines %d-%d", startLine+1, endLine))
	}

	if maxLineLength > m.fieldDetailCharsPerLine {
		statusParts = append(statusParts, fmt.Sprintf("chars %d-%d of %d",
			m.fieldDetailHorizontalOffset+1,
			min(m.fieldDetailHorizontalOffset+m.fieldDetailCharsPerLine, maxLineLength),
			maxLineLength))
	}

	statusLine := strings.Join(statusParts, " • ")

	// Simple help
	var helpParts []string
	if maxLineLength > m.fieldDetailCharsPerLine {
		helpParts = append(helpParts, "←/→: scroll horizontal")
	}
	if lineCount > m.fieldDetailLinesPerPage {
		helpParts = append(helpParts, "↑/↓: scroll vertical")
	}
	helpParts = append(helpParts, "c: copy to clipboard", "esc: back")

	helpText := strings.Join(helpParts, " • ")

	// Simple layout without excessive borders
	result := title + "\n"
	result += statusLine + "\n"
	result += strings.Repeat("-", 60) + "\n"
	result += visibleContent + "\n"
	result += strings.Repeat("-", 60) + "\n"
	result += helpText

	return result
}

// Command to load complete field value from database
func (m model) loadCompleteFieldValue(fieldName string) tea.Cmd {
	return func() tea.Msg {
		// Build query to get the specific field from the selected row
		var query string
		switch m.selectedDB.driver {
		case "postgres":
			if m.selectedSchema != "" {
				query = fmt.Sprintf(`SELECT "%s" FROM "%s"."%s" LIMIT 1 OFFSET %d`,
					fieldName, m.selectedSchema, m.selectedTable, m.selectedRowIndex)
			} else {
				query = fmt.Sprintf(`SELECT "%s" FROM "%s" LIMIT 1 OFFSET %d`,
					fieldName, m.selectedTable, m.selectedRowIndex)
			}
		case "mysql":
			query = fmt.Sprintf("SELECT `%s` FROM `%s` LIMIT 1 OFFSET %d",
				fieldName, m.selectedTable, m.selectedRowIndex)
		case "sqlite3":
			query = fmt.Sprintf(`SELECT "%s" FROM "%s" LIMIT 1 OFFSET %d`,
				fieldName, m.selectedTable, m.selectedRowIndex)
		default:
			query = fmt.Sprintf("SELECT %s FROM %s LIMIT 1 OFFSET %d",
				fieldName, m.selectedTable, m.selectedRowIndex)
		}

		// Execute query
		rows, err := m.db.Query(query)
		if err != nil {
			return fieldValueResult{
				fieldName: fieldName,
				value:     "",
				err:       err,
			}
		}
		defer rows.Close()

		if !rows.Next() {
			return fieldValueResult{
				fieldName: fieldName,
				value:     "",
				err:       fmt.Errorf("no data found"),
			}
		}

		var value interface{}
		err = rows.Scan(&value)
		if err != nil {
			return fieldValueResult{
				fieldName: fieldName,
				value:     "",
				err:       err,
			}
		}

		var valueStr string
		if value == nil {
			valueStr = "NULL"
		} else {
			valueStr = fmt.Sprintf("%v", value)
		}

		return fieldValueResult{
			fieldName: fieldName,
			value:     valueStr,
			err:       nil,
		}
	}
}

// Command to connect to database
func (m model) testConnection() tea.Cmd {
	return func() tea.Msg {
		return testConnectionWithTimeout(m.selectedDB.driver, m.connectionStr)
	}
}

// Test connection with timeout and better error messages
func testConnectionWithTimeout(driver, connectionStr string) testConnectionResult {
	// Basic validation
	if connectionStr == "" {
		return testConnectionResult{success: false, err: fmt.Errorf("connection string is empty")}
	}

	// Driver-specific validation
	if err := validateConnectionString(driver, connectionStr); err != nil {
		return testConnectionResult{success: false, err: err}
	}

	db, err := sql.Open(driver, connectionStr)
	if err != nil {
		return testConnectionResult{success: false, err: enhanceConnectionError(driver, err)}
	}
	defer db.Close()

	// Set connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return testConnectionResult{success: false, err: enhanceConnectionError(driver, err)}
	}

	return testConnectionResult{success: true, err: nil}
}

func (m model) connectDB() tea.Cmd {
	return func() tea.Msg {
		db, err := sql.Open(m.selectedDB.driver, m.connectionStr)
		if err != nil {
			return connectResult{err: err}
		}

		err = db.Ping()
		if err != nil {
			db.Close()
			return connectResult{err: err}
		}

		// Get schemas first for PostgreSQL
		schemas, err := getSchemas(db, m.selectedDB.driver)
		if err != nil {
			db.Close()
			return connectResult{err: err}
		}

		// If this is PostgreSQL and we have multiple schemas, return schemas for selection
		if m.selectedDB.driver == "postgres" && len(schemas) > 1 {
			return connectResult{db: db, schemas: schemas, requiresSchemaSelection: true}
		}

		// Otherwise, get tables using the default schema
		var defaultSchema string
		if len(schemas) > 0 {
			defaultSchema = schemas[0].name
		} else {
			switch m.selectedDB.driver {
			case "postgres":
				defaultSchema = "public"
			case "mysql":
				defaultSchema = "mysql"
			case "sqlite3":
				defaultSchema = "main"
			}
		}

		tableInfos, err := getTableInfos(db, m.selectedDB.driver, defaultSchema)
		if err != nil {
			db.Close()
			return connectResult{err: err}
		}

		// Extract table names for backward compatibility
		tables := make([]string, len(tableInfos))
		for i, info := range tableInfos {
			tables[i] = info.name
		}

		return connectResult{db: db, tables: tables, tableInfos: tableInfos, schemas: schemas, selectedSchema: defaultSchema}
	}
}

// Command to load columns
func (m model) loadColumns() tea.Cmd {
	return func() tea.Msg {
		columns, err := getColumns(m.db, m.selectedDB.driver, m.selectedTable, m.selectedSchema)
		if err != nil {
			return columnsResult{err: err}
		}
		return columnsResult{columns: columns}
	}
}

// Command to load tables for selected schema
func (m model) loadTablesForSchema() tea.Cmd {
	return func() tea.Msg {
		tableInfos, err := getTableInfos(m.db, m.selectedDB.driver, m.selectedSchema)
		if err != nil {
			return connectResult{err: err}
		}

		// Extract table names for backward compatibility
		tables := make([]string, len(tableInfos))
		for i, info := range tableInfos {
			tables[i] = info.name
		}

		return connectResult{tables: tables, tableInfos: tableInfos}
	}
}

// Command to load data preview
func (m model) loadDataPreview() tea.Cmd {
	return func() tea.Msg {
		// Build query with proper table name quoting based on database type
		var query string
		switch m.selectedDB.driver {
		case "postgres":
			// Include schema for PostgreSQL
			if m.selectedSchema != "" {
				query = fmt.Sprintf(`SELECT * FROM "%s"."%s" LIMIT 10`, m.selectedSchema, m.selectedTable)
			} else {
				query = fmt.Sprintf(`SELECT * FROM "%s" LIMIT 10`, m.selectedTable)
			}
		case "mysql":
			query = fmt.Sprintf("SELECT * FROM `%s` LIMIT 10", m.selectedTable)
		case "sqlite3":
			query = fmt.Sprintf(`SELECT * FROM "%s" LIMIT 10`, m.selectedTable)
		default:
			query = fmt.Sprintf("SELECT * FROM %s LIMIT 10", m.selectedTable)
		}

		// Set query timeout for data preview
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		rows, err := m.db.QueryContext(ctx, query)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return dataPreviewResult{err: fmt.Errorf("data preview timeout: operation took longer than 15 seconds")}
			}
			return dataPreviewResult{err: enhanceQueryError(err)}
		}
		defer rows.Close()

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			return dataPreviewResult{err: err}
		}

		// Get all rows
		var results [][]string
		for rows.Next() {
			// Create slice to hold column values
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))

			for i := range values {
				valuePtrs[i] = &values[i]
			}

			err := rows.Scan(valuePtrs...)
			if err != nil {
				return dataPreviewResult{err: err}
			}

			// Convert to strings
			strValues := make([]string, len(columns))
			for i, val := range values {
				if val != nil {
					strValues[i] = fmt.Sprintf("%v", val)
				} else {
					strValues[i] = "NULL"
				}
			}
			results = append(results, strValues)
		}

		return dataPreviewResult{columns: columns, rows: results}
	}
}

// Command to load indexes for a table
func (m model) loadIndexes() tea.Cmd {
	return func() tea.Msg {
		// Get indexes
		indexes, indexErr := getIndexes(m.db, m.selectedDB.driver, m.selectedTable, m.selectedSchema)

		// Get constraints
		constraints, constraintErr := getConstraints(m.db, m.selectedDB.driver, m.selectedTable, m.selectedSchema)

		// Combine both into a single result
		var combinedRows [][]string

		// Add indexes first
		if indexErr == nil {
			for _, index := range indexes {
				combinedRows = append(combinedRows, index)
			}
		}

		// Add constraints
		if constraintErr == nil {
			for _, constraint := range constraints {
				// Format constraint as index-like row: name, type, columns, definition
				name := constraint[0]
				constraintType := constraint[1]
				column := constraint[2]
				referenced := constraint[3]

				definition := fmt.Sprintf("CONSTRAINT %s", constraintType)
				if referenced != "" {
					definition += fmt.Sprintf(" REFERENCES %s", referenced)
				}

				combinedRows = append(combinedRows, []string{name, constraintType, column, definition})
			}
		}

		// Return combined result
		return indexesResult{
			indexes: combinedRows,
			err:     indexErr, // Return index error if it exists, otherwise constraint error
		}
	}
}

// Command to execute query
func (m model) executeQuery(query string) tea.Cmd {
	return func() tea.Msg {
		// Set query timeout to prevent hanging
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		rows, err := m.db.QueryContext(ctx, query)
		if err != nil {
			// Enhanced error handling for query execution
			if ctx.Err() == context.DeadlineExceeded {
				return queryResult{err: fmt.Errorf("query timeout: operation took longer than 30 seconds")}
			}
			return queryResult{err: enhanceQueryError(err)}
		}
		defer rows.Close()

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			return queryResult{err: err}
		}

		// Get all rows
		var results [][]string
		for rows.Next() {
			// Create slice to hold column values
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))

			for i := range values {
				valuePtrs[i] = &values[i]
			}

			err := rows.Scan(valuePtrs...)
			if err != nil {
				return queryResult{err: err}
			}

			// Convert to strings
			strValues := make([]string, len(columns))
			for i, val := range values {
				if val != nil {
					strValues[i] = fmt.Sprintf("%v", val)
				} else {
					strValues[i] = "NULL"
				}
			}
			results = append(results, strValues)
		}

		return queryResult{columns: columns, rows: results}
	}
}

// Result messages
type connectResult struct {
	db                      *sql.DB
	tables                  []string
	tableInfos              []tableInfo
	schemas                 []schemaInfo
	selectedSchema          string
	requiresSchemaSelection bool
	err                     error
}

type testConnectionResult struct {
	success bool
	err     error
}

type columnsResult struct {
	columns [][]string
	err     error
}

type queryResult struct {
	columns []string
	rows    [][]string
	err     error
}

type dataPreviewResult struct {
	columns []string
	rows    [][]string
	err     error
}

type indexesResult struct {
	indexes [][]string
	err     error
}

type clearResultMsg struct{}
type clearErrorMsg struct{}

type exportResult struct {
	filename string
	format   string
	rowCount int
	err      error
}

type testAndSaveResult struct {
	name       string
	connection SavedConnection
	success    bool
	err        error
}

type fieldValueResult struct {
	fieldName string
	value     string
	err       error
}

type clipboardResult struct {
	success bool
	err     error
}

// Command to clear query result after timeout
func clearResultAfterTimeout() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return clearResultMsg{}
	})
}

// Command to export data to CSV
func (m model) exportDataToCSV(columns []string, rows [][]string, tableName string) tea.Cmd {
	return func() tea.Msg {
		filename := generateExportFilename(tableName, "csv")
		err := exportToCSV(columns, rows, filename)
		return exportResult{
			filename: filename,
			format:   "CSV",
			rowCount: len(rows),
			err:      err,
		}
	}
}

// Command to export data to JSON
func (m model) exportDataToJSON(columns []string, rows [][]string, tableName string) tea.Cmd {
	return func() tea.Msg {
		filename := generateExportFilename(tableName, "json")
		err := exportToJSON(columns, rows, filename)
		return exportResult{
			filename: filename,
			format:   "JSON",
			rowCount: len(rows),
			err:      err,
		}
	}
}

// Command to test connection and save if successful
func (m model) testAndSaveConnection(name, connectionStr string) tea.Cmd {
	return func() tea.Msg {
		// Test connection first
		testResult := testConnectionWithTimeout(m.selectedDB.driver, connectionStr)
		if !testResult.success {
			return testAndSaveResult{
				name:    name,
				success: false,
				err:     testResult.err,
			}
		}

		// If test successful, save connection
		newConnection := SavedConnection{
			Name:          name,
			Driver:        m.selectedDB.driver,
			ConnectionStr: connectionStr,
		}

		return testAndSaveResult{
			name:       name,
			connection: newConnection,
			success:    true,
			err:        nil,
		}
	}
}

// Implement Update to handle results
func (m model) handleConnectResult(msg connectResult) (model, tea.Cmd) {
	if msg.err != nil {
		return m.handleErrorWithRecovery(msg.err, 3)
	}

	// Check if this is a tables-for-schema result (no db connection, just tables)
	if msg.db == nil && len(msg.tables) > 0 {
		// This is a result from loadTablesForSchema
		m.isLoadingTables = false
		m.tables = msg.tables
		m.tableInfos = msg.tableInfos
		m.state = tablesView
		m.err = nil

		// Create items for tables list with metadata
		items := make([]list.Item, len(msg.tableInfos))
		for i, tableInfo := range msg.tableInfos {
			items[i] = item{
				title: tableInfo.name,
				desc:  tableInfo.description,
			}
		}
		m.tablesList.SetItems(items)
		return m, nil
	}

	// This is a full connection result
	m.isConnecting = false
	m.isSavingConnection = false
	m.db = msg.db
	m.schemas = msg.schemas
	m.selectedSchema = msg.selectedSchema
	m.err = nil

	// If we require schema selection, go to schema view
	if msg.requiresSchemaSelection {
		m.state = schemaView
		m = m.updateSchemasList()
		return m, nil
	}

	// Otherwise, go directly to tables view
	m.tables = msg.tables
	m.tableInfos = msg.tableInfos
	m.state = tablesView

	// Create items for tables list with metadata
	items := make([]list.Item, len(msg.tableInfos))
	for i, tableInfo := range msg.tableInfos {
		items[i] = item{
			title: tableInfo.name,
			desc:  tableInfo.description,
		}
	}
	m.tablesList.SetItems(items)

	return m, nil
}

func (m model) handleTestConnectionResult(msg testConnectionResult) (model, tea.Cmd) {
	m.isTestingConnection = false
	if msg.err != nil {
		return m.handleErrorWithRecovery(msg.err, 3)
	}
	// Clear any previous error and show success message temporarily
	m.err = nil
	m.queryResult = "✅ Connection test successful!"
	// Start timeout to clear the success message
	return m, tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return clearResultMsg{}
	})
}

func (m model) handleColumnsResult(msg columnsResult) (model, tea.Cmd) {
	m.isLoadingColumns = false
	if msg.err != nil {
		return m.handleErrorWithRecovery(msg.err, 3)
	}

	m.err = nil
	m.state = columnsView

	// Convert to table rows
	rows := make([]table.Row, len(msg.columns))
	for i, col := range msg.columns {
		rows[i] = table.Row(col)
	}
	m.columnsTable.SetRows(rows)

	return m, nil
}

func (m model) handleQueryResult(msg queryResult) (model, tea.Cmd) {
	m.isExecutingQuery = false
	if msg.err != nil {
		// Add failed query to history
		m = m.addQueryToHistory(m.queryInput.Value(), false, 0)
		return m.handleErrorWithRecovery(msg.err, 3)
	}

	m.err = nil

	// Store raw data for export
	m.lastQueryColumns = msg.columns
	m.lastQueryRows = msg.rows

	// Set up table for query results
	if len(msg.rows) == 0 {
		m.queryResult = "Query executed successfully. No rows returned."
		// Clear the table
		m.queryResultsTable.SetColumns([]table.Column{})
		m.queryResultsTable.SetRows([]table.Row{})

		// Ensure query input has focus after query execution
		m.queryInput.Focus()
		m.queryResultsTable.Blur()

		// Add successful query to history (no rows)
		m = m.addQueryToHistory(m.queryInput.Value(), true, 0)

		// Start timeout to clear the message
		return m, clearResultAfterTimeout()
	} else {
		// Validate that we have columns
		if len(msg.columns) == 0 {
			m.queryResult = "Query executed successfully but returned no columns."
			m.queryResultsTable.SetColumns([]table.Column{})
			m.queryResultsTable.SetRows([]table.Row{})

			// Ensure query input has focus after query execution
			m.queryInput.Focus()
			m.queryResultsTable.Blur()

			// Add successful query to history (no columns)
			m = m.addQueryToHistory(m.queryInput.Value(), true, 0)

			// Start timeout to clear the message
			return m, clearResultAfterTimeout()
		}

		// Create columns for the table
		columns := make([]table.Column, len(msg.columns))
		for i, col := range msg.columns {
			width := 15 // Default width
			if len(col) > 15 {
				width = len(col) + 2
			}
			if width > 25 {
				width = 25 // Max width to prevent overflow
			}
			columns[i] = table.Column{
				Title: col,
				Width: width,
			}
		}

		// Create rows for the table with extra defensive validation
		rows := make([]table.Row, 0)
		expectedColumnCount := len(msg.columns)

		for _, row := range msg.rows {
			// Skip rows that don't match the expected column count
			if len(row) != expectedColumnCount {
				continue
			}

			// Ensure the row has the exact number of columns expected
			tableRow := make(table.Row, expectedColumnCount)
			validRow := true

			for j := 0; j < expectedColumnCount; j++ {
				if j < len(row) {
					val := row[j]
					// Truncate long values for display
					if len(val) > 23 {
						val = val[:20] + "..."
					}
					tableRow[j] = val
				} else {
					// This shouldn't happen due to our validation, but just in case
					tableRow[j] = "NULL"
					validRow = false
				}
			}

			// Only add valid rows
			if validRow {
				rows = append(rows, tableRow)
			}
		}

		// Recreate the table with proper initialization to avoid any state issues
		m.queryResultsTable = table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(10),
		)
		m.queryResultsTable.SetStyles(getMagentaTableStyles())

		m.queryResult = fmt.Sprintf("Query returned %d rows", len(rows))

		// Ensure query input has focus after query execution
		m.queryInput.Focus()
		m.queryResultsTable.Blur()

		// Add successful query to history
		m = m.addQueryToHistory(m.queryInput.Value(), true, len(rows))

		return m, nil
	}
}

func (m model) handleDataPreviewResult(msg dataPreviewResult) (model, tea.Cmd) {
	m.isLoadingPreview = false
	if msg.err != nil {
		return m.handleErrorWithRecovery(msg.err, 3)
	}

	m.err = nil
	m.state = dataPreviewView

	// Store raw data for export
	m.lastPreviewColumns = msg.columns
	m.lastPreviewRows = msg.rows

	// Validate that we have columns
	if len(msg.columns) == 0 {
		m.err = fmt.Errorf("no columns returned from query")
		return m, tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
			return clearErrorMsg{}
		})
	}

	// Create columns for the table
	columns := make([]table.Column, len(msg.columns))
	for i, col := range msg.columns {
		width := 15 // Default width
		if len(col) > 15 {
			width = len(col) + 2
		}
		if width > 25 {
			width = 25 // Max width to prevent overflow
		}
		columns[i] = table.Column{
			Title: col,
			Width: width,
		}
	}

	// Create rows for the table with extra defensive validation
	rows := make([]table.Row, 0)
	expectedColumnCount := len(msg.columns)

	// Ensure we have at least one column to avoid empty table issues
	if expectedColumnCount == 0 {
		m.err = fmt.Errorf("no columns available for preview")
		return m, tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
			return clearErrorMsg{}
		})
	}

	for _, row := range msg.rows {
		// Skip rows that don't match the expected column count
		if len(row) != expectedColumnCount {
			continue
		}

		// Ensure the row has the exact number of columns expected
		tableRow := make(table.Row, expectedColumnCount)
		validRow := true

		for j := 0; j < expectedColumnCount; j++ {
			if j < len(row) {
				val := row[j]
				// Truncate long values for display
				if len(val) > 23 {
					val = val[:20] + "..."
				}
				tableRow[j] = val
			} else {
				// This shouldn't happen due to our validation, but just in case
				tableRow[j] = "NULL"
				validRow = false
			}
		}

		// Only add valid rows
		if validRow {
			rows = append(rows, tableRow)
		}
	}

	// Ensure we have at least some data to display
	if len(rows) == 0 {
		m.err = fmt.Errorf("no valid data rows found for preview")
		return m, tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
			return clearErrorMsg{}
		})
	}

	// Recreate the table with proper initialization to avoid any state issues
	m.dataPreviewTable = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	m.dataPreviewTable.SetStyles(getMagentaTableStyles())

	return m, nil
}

func (m model) handleIndexesResult(msg indexesResult) (model, tea.Cmd) {
	if msg.err != nil {
		return m.handleErrorWithRecovery(msg.err, 3)
	}

	m.err = nil
	m.state = indexesView

	// Create rows for the indexes table
	rows := make([]table.Row, 0)
	for _, indexRow := range msg.indexes {
		if len(indexRow) >= 4 {
			rows = append(rows, table.Row(indexRow))
		}
	}

	// Update the indexes table
	m.indexesTable.SetRows(rows)
	m.indexesTable.Focus()

	return m, nil
}

func (m model) handleExportResult(msg exportResult) (model, tea.Cmd) {
	m.isExporting = false
	if msg.err != nil {
		m.err = msg.err
		// Start timeout to clear the error message
		return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
			return clearErrorMsg{}
		})
	}

	// Show enhanced success message with row count and file info
	m.err = nil
	rowText := "row"
	if msg.rowCount != 1 {
		rowText = "rows"
	}

	m.queryResult = fmt.Sprintf("✅ Exported %d %s to %s\n📄 %s",
		msg.rowCount, rowText, msg.format, msg.filename)

	// Start timeout to clear the success message (longer timeout for more detailed message)
	return m, tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return clearResultMsg{}
	})
}

func (m model) handleTestAndSaveResult(msg testAndSaveResult) (model, tea.Cmd) {
	m.isSavingConnection = false
	if !msg.success {
		return m.handleErrorWithRecovery(msg.err, 3)
	}

	// Connection test successful, save and connect
	m.err = nil
	m.savedConnections = append(m.savedConnections, msg.connection)
	saveConnections(m.savedConnections)

	// Show success message briefly then connect
	m.queryResult = fmt.Sprintf("✅ Connection validated and saved as '%s'", msg.name)
	m.isConnecting = true

	// Connect to the database and go to tables view
	return m, tea.Batch(
		tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
			return clearResultMsg{}
		}),
		m.connectDB(),
	)
}

func (m model) handleFieldValueResult(msg fieldValueResult) (model, tea.Cmd) {
	if msg.err != nil {
		return m.handleErrorWithRecovery(msg.err, 3)
	}

	// Set the complete field value and ensure we're in field detail view
	m.selectedFieldValue = msg.value
	if m.state != fieldDetailView {
		m.state = fieldDetailView
	}

	return m, nil
}

func (m model) copyToClipboard(content string) tea.Cmd {
	return func() tea.Msg {
		// Try to copy to clipboard using various methods
		var err error

		// Method 1: Try xclip (Linux)
		if _, cmdErr := exec.LookPath("xclip"); cmdErr == nil {
			cmd := exec.Command("xclip", "-selection", "clipboard")
			cmd.Stdin = strings.NewReader(content)
			err = cmd.Run()
			if err == nil {
				return clipboardResult{success: true, err: nil}
			}
		}

		// Method 2: Try xsel (Linux alternative)
		if _, cmdErr := exec.LookPath("xsel"); cmdErr == nil {
			cmd := exec.Command("xsel", "--clipboard", "--input")
			cmd.Stdin = strings.NewReader(content)
			err = cmd.Run()
			if err == nil {
				return clipboardResult{success: true, err: nil}
			}
		}

		// Method 3: Try pbcopy (macOS)
		if _, cmdErr := exec.LookPath("pbcopy"); cmdErr == nil {
			cmd := exec.Command("pbcopy")
			cmd.Stdin = strings.NewReader(content)
			err = cmd.Run()
			if err == nil {
				return clipboardResult{success: true, err: nil}
			}
		}

		// Method 4: Try clip (Windows)
		if _, cmdErr := exec.LookPath("clip"); cmdErr == nil {
			cmd := exec.Command("clip")
			cmd.Stdin = strings.NewReader(content)
			err = cmd.Run()
			if err == nil {
				return clipboardResult{success: true, err: nil}
			}
		}

		// If all methods failed
		return clipboardResult{
			success: false,
			err:     fmt.Errorf("clipboard not available (install xclip, xsel, pbcopy, or clip)"),
		}
	}
}

func (m model) handleClipboardResult(msg clipboardResult) (model, tea.Cmd) {
	if msg.success {
		m.queryResult = "✅ Content copied to clipboard"
	} else {
		m.err = msg.err
	}

	// Clear the message after a short delay
	return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return clearResultMsg{}
	})
}

// Helper functions for getting database information

func getTables(db *sql.DB, driver string) ([]string, error) {
	var query string
	switch driver {
	case "postgres":
		query = "SELECT tablename FROM pg_tables WHERE schemaname = 'public'"
	case "mysql":
		query = "SHOW TABLES"
	case "sqlite3":
		query = "SELECT name FROM sqlite_master WHERE type='table'"
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, nil
}

func getSchemas(db *sql.DB, driver string) ([]schemaInfo, error) {
	var schemas []schemaInfo

	switch driver {
	case "postgres":
		query := `
			SELECT 
				schema_name,
				CASE 
					WHEN schema_name = 'public' THEN 'Default public schema'
					WHEN schema_name IN ('information_schema', 'pg_catalog', 'pg_toast') THEN 'System schema'
					ELSE 'User schema'
				END as description
			FROM information_schema.schemata 
			WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
			ORDER BY 
				CASE WHEN schema_name = 'public' THEN 0 ELSE 1 END,
				schema_name`

		rows, err := db.Query(query)
		if err != nil {
			// If schema query fails, return just the public schema
			return []schemaInfo{{"public", "Default public schema"}}, nil
		}
		defer rows.Close()

		for rows.Next() {
			var schema schemaInfo
			err := rows.Scan(&schema.name, &schema.description)
			if err != nil {
				continue
			}
			schemas = append(schemas, schema)
		}

		// If no schemas found, add public as fallback
		if len(schemas) == 0 {
			schemas = append(schemas, schemaInfo{"public", "Default public schema"})
		}

	case "mysql", "sqlite3":
		// MySQL and SQLite don't have schemas in the same way PostgreSQL does
		// For MySQL, we could use databases, but for now we'll just return empty
		// This means schema selection won't be shown for these databases
		return []schemaInfo{}, nil
	}

	return schemas, nil
}

func getTableInfos(db *sql.DB, driver, schema string) ([]tableInfo, error) {
	var tableInfos []tableInfo

	switch driver {
	case "postgres":
		query := `
			SELECT 
				table_name,
				table_schema as schema_name,
				table_type,
				CASE 
					WHEN table_type = 'BASE TABLE' THEN COALESCE(s.n_tup_ins + s.n_tup_upd - s.n_tup_del, 0)
					ELSE 0
				END as estimated_rows
			FROM information_schema.tables t
			LEFT JOIN pg_stat_user_tables s ON t.table_name = s.relname AND t.table_schema = s.schemaname
			WHERE t.table_schema = $1 
				AND t.table_type IN ('BASE TABLE', 'VIEW')
			ORDER BY t.table_type, t.table_name`

		rows, err := db.Query(query, schema)
		if err != nil {
			// Fallback to simple table list if stats are not available
			return getSimpleTableInfos(db, driver, schema)
		}
		defer rows.Close()

		for rows.Next() {
			var info tableInfo
			var estimatedRows sql.NullInt64
			err := rows.Scan(&info.name, &info.schema, &info.tableType, &estimatedRows)
			if err != nil {
				continue
			}

			// Determine the object type display name
			var objectType string
			var emoji string
			if info.tableType == "VIEW" {
				objectType = "view"
				emoji = "👁️"
			} else {
				objectType = "table"
				emoji = "📊"
			}

			if info.tableType == "BASE TABLE" && estimatedRows.Valid && estimatedRows.Int64 > 0 {
				info.rowCount = estimatedRows.Int64
				if info.rowCount < 0 {
					info.rowCount = 0
				}
				if info.schema != "" && info.schema != "public" {
					info.description = fmt.Sprintf("%s %s.%s • ~%d rows", emoji, info.schema, objectType, info.rowCount)
				} else {
					info.description = fmt.Sprintf("%s %s • ~%d rows", emoji, strings.Title(objectType), info.rowCount)
				}
			} else {
				if info.schema != "" && info.schema != "public" {
					info.description = fmt.Sprintf("%s %s.%s", emoji, info.schema, objectType)
				} else {
					info.description = fmt.Sprintf("%s %s", emoji, strings.Title(objectType))
				}
			}

			tableInfos = append(tableInfos, info)
		}

	case "mysql":
		query := `
			SELECT 
				TABLE_NAME,
				TABLE_SCHEMA,
				TABLE_TYPE,
				COALESCE(TABLE_ROWS, 0) as table_rows
			FROM INFORMATION_SCHEMA.TABLES 
			WHERE TABLE_SCHEMA = DATABASE()
				AND TABLE_TYPE IN ('BASE TABLE', 'VIEW')
			ORDER BY TABLE_TYPE, TABLE_NAME`

		rows, err := db.Query(query)
		if err != nil {
			return getSimpleTableInfos(db, driver, schema)
		}
		defer rows.Close()

		for rows.Next() {
			var info tableInfo
			var tableRows sql.NullInt64
			err := rows.Scan(&info.name, &info.schema, &info.tableType, &tableRows)
			if err != nil {
				continue
			}

			// Determine the object type display name
			var objectType string
			var emoji string
			if info.tableType == "VIEW" {
				objectType = "view"
				emoji = "👁️"
			} else {
				objectType = "table"
				emoji = "📊"
			}

			if info.tableType == "BASE TABLE" && tableRows.Valid && tableRows.Int64 > 0 {
				info.rowCount = tableRows.Int64
				info.description = fmt.Sprintf("%s %s • ~%d rows", emoji, strings.Title(objectType), info.rowCount)
			} else {
				info.description = fmt.Sprintf("%s %s", emoji, strings.Title(objectType))
			}

			tableInfos = append(tableInfos, info)
		}

	case "sqlite3":
		// SQLite: Get both tables and views from sqlite_master
		query := `
			SELECT name, type 
			FROM sqlite_master 
			WHERE type IN ('table', 'view') 
				AND name NOT LIKE 'sqlite_%'
			ORDER BY type, name`

		rows, err := db.Query(query)
		if err != nil {
			return getSimpleTableInfos(db, driver, schema)
		}
		defer rows.Close()

		for rows.Next() {
			var name, objType string
			err := rows.Scan(&name, &objType)
			if err != nil {
				continue
			}

			info := tableInfo{
				name:   name,
				schema: "main", // SQLite uses "main" as the default schema
			}

			// Determine the object type display name
			var objectType string
			var emoji string
			if objType == "view" {
				info.tableType = "VIEW"
				objectType = "view"
				emoji = "👁️"
			} else {
				info.tableType = "BASE TABLE"
				objectType = "table"
				emoji = "📊"
			}

			// Try to get row count for tables only (views don't have meaningful row counts)
			if objType == "table" {
				countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, name)
				var count int64
				err := db.QueryRow(countQuery).Scan(&count)
				if err == nil {
					info.rowCount = count
					info.description = fmt.Sprintf("%s %s • %d rows", emoji, strings.Title(objectType), count)
				} else {
					info.description = fmt.Sprintf("%s %s", emoji, strings.Title(objectType))
				}
			} else {
				info.description = fmt.Sprintf("%s %s", emoji, strings.Title(objectType))
			}

			tableInfos = append(tableInfos, info)
		}

	default:
		return getSimpleTableInfos(db, driver, schema)
	}

	return tableInfos, nil
}

func getSimpleTableInfos(db *sql.DB, driver, schema string) ([]tableInfo, error) {
	tables, err := getTables(db, driver)
	if err != nil {
		return nil, err
	}

	var tableInfos []tableInfo
	for _, tableName := range tables {
		var schemaName string
		switch driver {
		case "postgres":
			schemaName = schema
		case "mysql":
			schemaName = "mysql" // Default schema name for MySQL
		case "sqlite3":
			schemaName = "main"
		default:
			schemaName = ""
		}

		tableInfos = append(tableInfos, tableInfo{
			name:        tableName,
			schema:      schemaName,
			tableType:   "BASE TABLE",
			description: "📊 Table",
		})
	}

	return tableInfos, nil
}

func getColumns(db *sql.DB, driver, tableName, schema string) ([][]string, error) {
	var query string
	switch driver {
	case "postgres":
		query = `SELECT column_name, data_type, is_nullable, column_default 
				 FROM information_schema.columns 
				 WHERE table_name = $1 AND table_schema = $2
				 ORDER BY ordinal_position`
	case "mysql":
		query = `SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT 
				 FROM INFORMATION_SCHEMA.COLUMNS 
				 WHERE TABLE_NAME = ? 
				 ORDER BY ORDINAL_POSITION`
	case "sqlite3":
		query = fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	}

	var rows *sql.Rows
	var err error

	switch driver {
	case "postgres":
		rows, err = db.Query(query, tableName, schema)
	case "mysql":
		rows, err = db.Query(query, tableName)
	case "sqlite3":
		rows, err = db.Query(query)
	default:
		rows, err = db.Query(query, tableName)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns [][]string
	for rows.Next() {
		if driver == "sqlite3" {
			var cid int
			var name, dataType string
			var notNull int
			var defaultValue sql.NullString
			var pk int

			err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
			if err != nil {
				return nil, err
			}

			nullable := "YES"
			if notNull == 1 {
				nullable = "NO"
			}

			def := ""
			if defaultValue.Valid {
				def = defaultValue.String
			}

			columns = append(columns, []string{name, dataType, nullable, def})
		} else {
			var name, dataType, nullable string
			var defaultValue sql.NullString

			err := rows.Scan(&name, &dataType, &nullable, &defaultValue)
			if err != nil {
				return nil, err
			}

			def := ""
			if defaultValue.Valid {
				def = defaultValue.String
			}

			columns = append(columns, []string{name, dataType, nullable, def})
		}
	}

	return columns, nil
}

func getIndexes(db *sql.DB, driver, tableName, schema string) ([][]string, error) {
	var query string
	switch driver {
	case "postgres":
		query = `SELECT 
					i.indexname as index_name,
					i.indexdef as index_definition,
					CASE 
						WHEN i.indexdef LIKE '%UNIQUE%' THEN 'UNIQUE'
						WHEN i.indexdef LIKE '%PRIMARY%' THEN 'PRIMARY'
						ELSE 'INDEX'
					END as index_type,
					pg_get_indexdef(pg_class.oid) as columns
				FROM pg_indexes i
				JOIN pg_class ON pg_class.relname = i.indexname
				WHERE i.tablename = $1 AND i.schemaname = $2
				ORDER BY i.indexname`
	case "mysql":
		query = `SELECT 
					INDEX_NAME as index_name,
					CONCAT('INDEX ON ', COLUMN_NAME) as index_definition,
					CASE 
						WHEN NON_UNIQUE = 0 AND INDEX_NAME = 'PRIMARY' THEN 'PRIMARY'
						WHEN NON_UNIQUE = 0 THEN 'UNIQUE'
						ELSE 'INDEX'
					END as index_type,
					GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX) as columns
				FROM INFORMATION_SCHEMA.STATISTICS 
				WHERE TABLE_NAME = ? AND TABLE_SCHEMA = DATABASE()
				GROUP BY INDEX_NAME, NON_UNIQUE
				ORDER BY INDEX_NAME`
	case "sqlite3":
		query = fmt.Sprintf("PRAGMA index_list(%s)", tableName)
	}

	var rows *sql.Rows
	var err error

	switch driver {
	case "postgres":
		rows, err = db.Query(query, tableName, schema)
	case "mysql":
		rows, err = db.Query(query, tableName)
	case "sqlite3":
		rows, err = db.Query(query)
	default:
		rows, err = db.Query(query, tableName)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes [][]string
	for rows.Next() {
		if driver == "sqlite3" {
			var seq int
			var name string
			var unique int
			var origin, partial string

			err := rows.Scan(&seq, &name, &unique, &origin, &partial)
			if err != nil {
				return nil, err
			}

			indexType := "INDEX"
			if unique == 1 {
				indexType = "UNIQUE"
			}

			// Get columns for this index
			indexInfoQuery := fmt.Sprintf("PRAGMA index_info(%s)", name)
			indexInfoRows, err := db.Query(indexInfoQuery)
			if err != nil {
				continue
			}

			var columns []string
			for indexInfoRows.Next() {
				var seqno, cid int
				var colName string
				if err := indexInfoRows.Scan(&seqno, &cid, &colName); err == nil {
					columns = append(columns, colName)
				}
			}
			indexInfoRows.Close()

			columnsStr := strings.Join(columns, ", ")
			definition := fmt.Sprintf("INDEX ON (%s)", columnsStr)

			indexes = append(indexes, []string{name, definition, indexType, columnsStr})
		} else {
			var name, definition, indexType, columns string

			err := rows.Scan(&name, &definition, &indexType, &columns)
			if err != nil {
				return nil, err
			}

			indexes = append(indexes, []string{name, definition, indexType, columns})
		}
	}

	return indexes, nil
}

func getConstraints(db *sql.DB, driver, tableName, schema string) ([][]string, error) {
	var query string
	switch driver {
	case "postgres":
		query = `SELECT 
					tc.constraint_name,
					tc.constraint_type,
					kcu.column_name,
					COALESCE(ccu.table_name || '.' || ccu.column_name, '') as referenced_table_column
				FROM information_schema.table_constraints tc
				LEFT JOIN information_schema.key_column_usage kcu 
					ON tc.constraint_name = kcu.constraint_name 
					AND tc.table_schema = kcu.table_schema
				LEFT JOIN information_schema.constraint_column_usage ccu
					ON tc.constraint_name = ccu.constraint_name
					AND tc.table_schema = ccu.table_schema
				WHERE tc.table_name = $1 AND tc.table_schema = $2
				ORDER BY tc.constraint_name, kcu.ordinal_position`
	case "mysql":
		query = `SELECT 
					CONSTRAINT_NAME as constraint_name,
					CONSTRAINT_TYPE as constraint_type,
					COLUMN_NAME as column_name,
					COALESCE(CONCAT(REFERENCED_TABLE_NAME, '.', REFERENCED_COLUMN_NAME), '') as referenced_table_column
				FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
				JOIN INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc 
					ON kcu.CONSTRAINT_NAME = tc.CONSTRAINT_NAME 
					AND kcu.TABLE_SCHEMA = tc.TABLE_SCHEMA
				WHERE kcu.TABLE_NAME = ? AND kcu.TABLE_SCHEMA = DATABASE()
				ORDER BY kcu.CONSTRAINT_NAME, kcu.ORDINAL_POSITION`
	case "sqlite3":
		query = fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)
	}

	var rows *sql.Rows
	var err error

	switch driver {
	case "postgres":
		rows, err = db.Query(query, tableName, schema)
	case "mysql":
		rows, err = db.Query(query, tableName)
	case "sqlite3":
		rows, err = db.Query(query)
	default:
		rows, err = db.Query(query, tableName)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var constraints [][]string
	if driver == "sqlite3" {
		// Handle SQLite foreign keys
		for rows.Next() {
			var id, seq int
			var table, from, to, onUpdate, onDelete, match string

			err := rows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match)
			if err != nil {
				return nil, err
			}

			constraintName := fmt.Sprintf("fk_%s_%s", tableName, from)
			constraintType := "FOREIGN KEY"
			referencedTableColumn := fmt.Sprintf("%s.%s", table, to)

			constraints = append(constraints, []string{constraintName, constraintType, from, referencedTableColumn})
		}
	} else {
		for rows.Next() {
			var name, constraintType, column, referencedTableColumn string

			err := rows.Scan(&name, &constraintType, &column, &referencedTableColumn)
			if err != nil {
				return nil, err
			}

			constraints = append(constraints, []string{name, constraintType, column, referencedTableColumn})
		}
	}

	return constraints, nil
}

// Method to update saved connections list
func (m model) updateSavedConnectionsList() model {
	items := make([]list.Item, len(m.savedConnections))
	for i, conn := range m.savedConnections {
		// Safely truncate connection string
		connStr := conn.ConnectionStr
		if len(connStr) > 50 {
			connStr = connStr[:50] + "..."
		}

		items[i] = item{
			title: conn.Name,
			desc:  fmt.Sprintf("%s - %s", conn.Driver, connStr),
		}
	}
	m.savedConnectionsList.SetItems(items)
	return m
}

// Search functionality helper functions

// Filter table items based on search term
func (m model) filterTableItems(searchTerm string) []list.Item {
	if searchTerm == "" {
		return m.originalTableItems
	}

	var filtered []list.Item
	searchLower := strings.ToLower(searchTerm)

	for _, listItem := range m.originalTableItems {
		if tableItem, ok := listItem.(item); ok {
			// Search in both title (table name) and description
			if strings.Contains(strings.ToLower(tableItem.title), searchLower) ||
				strings.Contains(strings.ToLower(tableItem.desc), searchLower) {
				filtered = append(filtered, listItem)
			}
		}
	}

	return filtered
}

// Filter column rows based on search term
func (m model) filterColumnRows(searchTerm string) []table.Row {
	if searchTerm == "" {
		return m.originalTableRows
	}

	var filtered []table.Row
	searchLower := strings.ToLower(searchTerm)

	for _, row := range m.originalTableRows {
		// Search across all columns (column name, type, null, default)
		for _, cell := range row {
			if strings.Contains(strings.ToLower(cell), searchLower) {
				filtered = append(filtered, row)
				break // Found match in this row, add it and move to next row
			}
		}
	}

	return filtered
}

// Functions for saved connections management

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".dbx")
	return configDir, nil
}

func getConnectionsFile() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "connections.json"), nil
}

func loadSavedConnections() ([]SavedConnection, error) {
	connectionsFile, err := getConnectionsFile()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(connectionsFile); os.IsNotExist(err) {
		return []SavedConnection{}, nil
	}

	data, err := os.ReadFile(connectionsFile)
	if err != nil {
		return nil, err
	}

	var connections []SavedConnection
	err = json.Unmarshal(data, &connections)
	if err != nil {
		return nil, err
	}

	return connections, nil
}

func saveConnections(connections []SavedConnection) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}

	connectionsFile, err := getConnectionsFile()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(connections, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(connectionsFile, data, 0644)
}

// Query history management functions

func getQueryHistoryFile() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "query_history.json"), nil
}

func loadQueryHistory() ([]QueryHistoryEntry, error) {
	historyFile, err := getQueryHistoryFile()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		return []QueryHistoryEntry{}, nil
	}

	data, err := os.ReadFile(historyFile)
	if err != nil {
		return nil, err
	}

	var history []QueryHistoryEntry
	err = json.Unmarshal(data, &history)
	if err != nil {
		return nil, err
	}

	return history, nil
}

func saveQueryHistory(history []QueryHistoryEntry) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}

	historyFile, err := getQueryHistoryFile()
	if err != nil {
		return err
	}

	// Limit history to last 100 entries
	maxEntries := 100
	if len(history) > maxEntries {
		history = history[len(history)-maxEntries:]
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyFile, data, 0644)
}

func (m model) addQueryToHistory(query string, success bool, rowCount int) model {
	entry := QueryHistoryEntry{
		Query:     query,
		Timestamp: time.Now(),
		Database:  "", // TODO: could add database name if available
		Success:   success,
		RowCount:  rowCount,
	}

	m.queryHistory = append(m.queryHistory, entry)

	// Save to file
	go saveQueryHistory(m.queryHistory) // Save in background to avoid blocking UI

	return m
}

func (m model) updateSchemasList() model {
	items := make([]list.Item, len(m.schemas))
	for i, schema := range m.schemas {
		items[i] = item{
			title: schema.name,
			desc:  schema.description,
		}
	}
	m.schemasList.SetItems(items)
	return m
}

func (m *model) updateQueryHistoryList() {
	// Convert query history to list items (reverse order to show newest first)
	items := make([]list.Item, len(m.queryHistory))
	for i := len(m.queryHistory) - 1; i >= 0; i-- {
		entry := m.queryHistory[i]

		// Clean and truncate query for display
		displayQuery := strings.ReplaceAll(entry.Query, "\n", " ") // Replace newlines with spaces
		displayQuery = strings.ReplaceAll(displayQuery, "\t", " ") // Replace tabs with spaces
		displayQuery = strings.TrimSpace(displayQuery)             // Remove leading/trailing whitespace

		// Collapse multiple spaces into single spaces
		for strings.Contains(displayQuery, "  ") {
			displayQuery = strings.ReplaceAll(displayQuery, "  ", " ")
		}

		// Ensure we have at least some text to display
		if displayQuery == "" {
			displayQuery = "[Empty Query]"
		}

		if len(displayQuery) > 60 {
			displayQuery = displayQuery[:60] + "..."
		}

		// Format timestamp
		timeStr := entry.Timestamp.Format("2006-01-02 15:04")

		// Create description with success status and row count
		var status string
		if entry.Success {
			if entry.RowCount > 0 {
				status = fmt.Sprintf("✅ %s (%d rows)", timeStr, entry.RowCount)
			} else {
				status = fmt.Sprintf("✅ %s (no rows)", timeStr)
			}
		} else {
			status = fmt.Sprintf("❌ %s (failed)", timeStr)
		}

		items[len(m.queryHistory)-1-i] = item{
			title: displayQuery,
			desc:  fmt.Sprintf("Query %d: %s", len(m.queryHistory)-i, status),
		}
	}

	m.queryHistoryList.SetItems(items)
}

// Export functions for query results
func exportToCSV(columns []string, rows [][]string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(columns); err != nil {
		return err
	}

	// Write data rows
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func exportToJSON(columns []string, rows [][]string, filename string) error {
	// Convert to array of maps for JSON
	var jsonData []map[string]interface{}

	for _, row := range rows {
		record := make(map[string]interface{})
		for i, col := range columns {
			if i < len(row) {
				// Convert "NULL" back to nil for JSON
				if row[i] == "NULL" {
					record[col] = nil
				} else {
					record[col] = row[i]
				}
			}
		}
		jsonData = append(jsonData, record)
	}

	// Write to file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonData)
}

// Generate default filename with timestamp
func generateExportFilename(tableName, format string) string {
	timestamp := time.Now().Format("20060102_150405")
	if tableName != "" {
		return fmt.Sprintf("%s_%s.%s", tableName, timestamp, format)
	}
	return fmt.Sprintf("query_result_%s.%s", timestamp, format)
}

// Validate connection string format for different database types
func validateConnectionString(driver, connectionStr string) error {
	switch driver {
	case "postgres":
		if !strings.Contains(connectionStr, "://") {
			return fmt.Errorf("PostgreSQL connection string must use format: postgres://user:password@host:port/database")
		}
		if !strings.HasPrefix(connectionStr, "postgres://") && !strings.HasPrefix(connectionStr, "postgresql://") {
			return fmt.Errorf("PostgreSQL connection string must start with 'postgres://' or 'postgresql://'")
		}
	case "mysql":
		if strings.Contains(connectionStr, "://") {
			return fmt.Errorf("MySQL connection string must use format: user:password@tcp(host:port)/database")
		}
		if !strings.Contains(connectionStr, "@tcp(") {
			return fmt.Errorf("MySQL connection string must include '@tcp(host:port)' format")
		}
	case "sqlite3":
		if strings.Contains(connectionStr, "://") || strings.Contains(connectionStr, "@") {
			return fmt.Errorf("SQLite connection string should be a file path: /path/to/database.db")
		}
		if connectionStr == ":memory:" {
			return nil // Special case for in-memory database
		}
		// Check if it looks like a valid file path
		if len(connectionStr) < 1 {
			return fmt.Errorf("SQLite connection string cannot be empty")
		}

		// Enhanced SQLite validation
		if err := validateSQLiteConnection(connectionStr); err != nil {
			return err
		}
	}
	return nil
}

// Enhanced validation for SQLite connections
func validateSQLiteConnection(path string) error {
	if path == ":memory:" {
		return nil
	}

	// Clean the path
	path = filepath.Clean(path)

	// Check if path is a directory
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return fmt.Errorf("path is a directory, not a database file: %s", path)
		}
		// File exists, check if it's readable
		if file, err := os.Open(path); err != nil {
			return fmt.Errorf("cannot read database file: %s", err.Error())
		} else {
			file.Close()
		}
		return nil
	}

	// File doesn't exist, check if parent directory exists and is writable
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	// Check if we can create the file (for write permissions)
	testFile := filepath.Join(dir, ".dbx_write_test")
	if file, err := os.Create(testFile); err != nil {
		return fmt.Errorf("cannot write to directory: %s", dir)
	} else {
		file.Close()
		os.Remove(testFile) // Clean up test file
	}

	return nil
}

// Enhance connection errors with user-friendly messages
func enhanceConnectionError(driver string, err error) error {
	errStr := err.Error()

	switch driver {
	case "postgres":
		if strings.Contains(errStr, "connection refused") {
			return fmt.Errorf("PostgreSQL server is not running or not accepting connections on the specified host/port")
		}
		if strings.Contains(errStr, "authentication failed") || strings.Contains(errStr, "password authentication failed") {
			return fmt.Errorf("PostgreSQL authentication failed - check username and password")
		}
		if strings.Contains(errStr, "database") && strings.Contains(errStr, "does not exist") {
			return fmt.Errorf("PostgreSQL database does not exist - check database name")
		}
		if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline exceeded") {
			return fmt.Errorf("PostgreSQL connection timeout - check host and port, ensure server is accessible")
		}
	case "mysql":
		if strings.Contains(errStr, "connection refused") {
			return fmt.Errorf("MySQL server is not running or not accepting connections on the specified host/port")
		}
		if strings.Contains(errStr, "Access denied") {
			return fmt.Errorf("MySQL access denied - check username and password")
		}
		if strings.Contains(errStr, "Unknown database") {
			return fmt.Errorf("MySQL database does not exist - check database name")
		}
		if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline exceeded") {
			return fmt.Errorf("MySQL connection timeout - check host and port, ensure server is accessible")
		}
	case "sqlite3":
		if strings.Contains(errStr, "no such file") {
			return fmt.Errorf("SQLite database file does not exist: %s", errStr)
		}
		if strings.Contains(errStr, "permission denied") {
			return fmt.Errorf("SQLite permission denied - check file permissions")
		}
		if strings.Contains(errStr, "database is locked") {
			return fmt.Errorf("SQLite database is locked - close other connections to this file")
		}
	}

	// Return enhanced error with original message for unknown cases
	return fmt.Errorf("%s connection error: %s", strings.Title(driver), errStr)
}

// Enhance query execution errors with user-friendly messages
func enhanceQueryError(err error) error {
	errStr := err.Error()

	// Common SQL error patterns
	if strings.Contains(errStr, "syntax error") {
		return fmt.Errorf("SQL syntax error: %s", errStr)
	}
	if strings.Contains(errStr, "no such table") || strings.Contains(errStr, "doesn't exist") {
		return fmt.Errorf("table or column does not exist: %s", errStr)
	}
	if strings.Contains(errStr, "permission denied") || strings.Contains(errStr, "access denied") {
		return fmt.Errorf("insufficient permissions: %s", errStr)
	}
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline exceeded") {
		return fmt.Errorf("query timeout - operation took too long to complete")
	}
	if strings.Contains(errStr, "connection") && strings.Contains(errStr, "lost") {
		return fmt.Errorf("database connection lost - please reconnect")
	}
	if strings.Contains(errStr, "deadlock") {
		return fmt.Errorf("database deadlock detected - try again")
	}

	// Return original error if no enhancement available
	return err
}

// Reset all loading states to prevent stuck UI states
func (m model) resetLoadingStates() model {
	m.isConnecting = false
	m.isTestingConnection = false
	m.isExecutingQuery = false
	m.isLoadingColumns = false
	m.isLoadingPreview = false
	m.isExporting = false
	m.isSavingConnection = false
	return m
}

// Enhanced error handling with automatic loading state recovery
func (m model) handleErrorWithRecovery(err error, timeoutSeconds int) (model, tea.Cmd) {
	// Reset any stuck loading states
	m = m.resetLoadingStates()

	// Set error and start timeout to clear it
	m.err = err
	return m, tea.Tick(time.Duration(timeoutSeconds)*time.Second, func(t time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
