package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/list"
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
	primaryMagenta   = lipgloss.Color("#D946EF")  // Main magenta
	lightMagenta     = lipgloss.Color("#F3E8FF")  // Light magenta background
	darkMagenta      = lipgloss.Color("#7C2D91")  // Dark magenta
	accentMagenta    = lipgloss.Color("#A855F7")  // Purple accent
	
	// Supporting colors
	darkGray         = lipgloss.Color("#374151")
	lightGray        = lipgloss.Color("#9CA3AF")
	white            = lipgloss.Color("#FFFFFF")
	successGreen     = lipgloss.Color("#10B981")
	errorRed         = lipgloss.Color("#EF4444")
	warningOrange    = lipgloss.Color("#F59E0B")

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
func getLoadingText(message string) string {
	// Simple rotating spinner
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	// Use a simple counter based approach for demonstration
	// In a real app, you'd use time-based animation
	spinnerIndex := 0 // This could be animated with a ticker
	return loadingStyle.Render(spinners[spinnerIndex] + " " + message)
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
	tablesView
	columnsView
	queryView
	dataPreviewView
)

// Saved connection
type SavedConnection struct {
	Name          string `json:"name"`
	Driver        string `json:"driver"`
	ConnectionStr string `json:"connection_str"`
}

// Database types
type dbType struct {
	name   string
	driver string
}

// Table information
type tableInfo struct {
	name        string
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
	selectedDB           dbType
	connectionStr        string
	db                   *sql.DB
	err                  error
	tables               []string
	tableInfos           []tableInfo
	selectedTable        string
	savedConnections     []SavedConnection
	editingConnectionIdx int
	queryResult          string
	width                int
	height               int
	// Loading states
	isTestingConnection  bool
	isConnecting         bool
	isSavingConnection   bool
	isLoadingTables      bool
	isLoadingColumns     bool
	isExecutingQuery     bool
	isLoadingPreview     bool
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

	return model{
		state:                dbTypeView,
		dbTypeList:           dbList,
		savedConnectionsList: savedConnectionsList,
		textInput:            ti,
		nameInput:            ni,
		queryInput:           qi,
		tablesList:           tablesList,
		columnsTable:         t,
		queryResultsTable:    queryResultsTable,
		dataPreviewTable:     dataPreviewTable,
		savedConnections:     savedConnections,
		editingConnectionIdx: -1,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
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
		m.textInput.Width = msg.Width - h - 4
		m.nameInput.Width = msg.Width - h - 4
		m.queryInput.Width = msg.Width - h - 4

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
			case tablesView:
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
				m.state = tablesView
			case queryView:
				m.state = tablesView
			case dataPreviewView:
				m.state = tablesView
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
			if m.state == connectionView && !m.isSavingConnection && !m.isConnecting {
				// Save connection directly with both name and connection string
				name := m.nameInput.Value()
				connectionStr := m.textInput.Value()
				if name != "" && connectionStr != "" {
					m.isSavingConnection = true
					m.err = nil
					m.queryResult = ""
					newConnection := SavedConnection{
						Name:          name,
						Driver:        m.selectedDB.driver,
						ConnectionStr: connectionStr,
					}
					m.savedConnections = append(m.savedConnections, newConnection)
					saveConnections(m.savedConnections)
					// Connect to the database and go to tables view
					m.connectionStr = connectionStr
					m.isSavingConnection = false
					m.isConnecting = true
					return m, m.connectDB()
				}
			}

		case "r":
			if (m.state == tablesView || m.state == columnsView) && m.db != nil {
				// Go to query view
				m.state = queryView
				m.queryInput.SetValue("")
				m.queryInput.Focus()
				m.queryResultsTable.Blur()
				return m, nil
			}

		case "p":
			if m.state == tablesView && m.db != nil && !m.isLoadingPreview {
				// Preview table data
				if i, ok := m.tablesList.SelectedItem().(item); ok {
					m.selectedTable = i.title
					m.isLoadingPreview = true
					m.err = nil
					return m, m.loadDataPreview()
				}
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

			case tablesView:
				if i, ok := m.tablesList.SelectedItem().(item); ok && !m.isLoadingColumns {
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
		m.tablesList, cmd = m.tablesList.Update(msg)
	case columnsView:
		m.columnsTable, cmd = m.columnsTable.Update(msg)
	case dataPreviewView:
		m.dataPreviewTable, cmd = m.dataPreviewTable.Update(msg)
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
	case tablesView:
		return m.tablesView()
	case columnsView:
		return m.columnsView()
	case queryView:
		return m.queryView()
	case dataPreviewView:
		return m.dataPreviewView()
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
		loadingMsg := getLoadingText("Connecting to saved connection...")
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
		messageContent = getLoadingText("Testing connection...")
	} else if m.isSavingConnection {
		messageContent = getLoadingText("Saving connection...")
	} else if m.isConnecting {
		messageContent = getLoadingText("Connecting to database...")
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
		keyStyle.Render("F2") + ": save & connect • " +
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
	
	if m.isLoadingColumns {
		loadingMsg := getLoadingText("Loading table columns...")
		content = m.tablesList.View() + "\n" + loadingMsg
	} else if m.isLoadingPreview {
		loadingMsg := getLoadingText("Loading table data preview...")
		content = m.tablesList.View() + "\n" + loadingMsg
	} else if len(m.tables) == 0 {
		emptyMsg := infoStyle.Render("📋 No tables found in this database.\n\nThe database might be empty or you might not have sufficient permissions.")
		content = m.tablesList.View() + "\n" + emptyMsg
	} else {
		statusText := successStyle.Render(fmt.Sprintf("✅ Connected successfully (%d tables found)", len(m.tables)))
		content = "\n" + statusText + "\n\n" + m.tablesList.View()
	}
	
	helpText := helpStyle.Render(
		keyStyle.Render("enter") + ": view columns • " +
		keyStyle.Render("p") + ": preview data • " +
		keyStyle.Render("r") + ": run query • " +
		keyStyle.Render("esc") + ": disconnect",
	)
	
	return docStyle.Render(content + "\n" + helpText)
}

func (m model) columnsView() string {
	content := titleStyle.Render(fmt.Sprintf("Columns of table: %s", m.selectedTable)) + "\n\n"
	content += m.columnsTable.View()
	content += "\n" + helpStyle.Render("↑/↓: navigate • r: run query • s: save connection • esc: back to tables")
	return docStyle.Render(content)
}

func (m model) queryView() string {
	title := titleStyle.Render("⚡  SQL Query Runner")
	
	var messageContent string
	if m.isExecutingQuery {
		messageContent = getLoadingText("Executing query...")
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

	if m.err != nil {
		content += errorStyle.Render("❌ Error: "+m.err.Error()) + "\n\n"
	}

	// Only show the table if it has both columns and rows, and they match
	if len(m.dataPreviewTable.Columns()) > 0 && len(m.dataPreviewTable.Rows()) > 0 {
		content += successStyle.Render(fmt.Sprintf("Showing first 10 rows from %s", m.selectedTable)) + "\n\n"
		content += m.dataPreviewTable.View()
	} else if m.err == nil {
		content += helpStyle.Render("Loading data preview...")
	}

	content += "\n\n" + helpStyle.Render("↑/↓: navigate rows • esc: back to tables")
	return docStyle.Render(content)
}

// Command to connect to database
func (m model) testConnection() tea.Cmd {
	return func() tea.Msg {
		db, err := sql.Open(m.selectedDB.driver, m.connectionStr)
		if err != nil {
			return testConnectionResult{success: false, err: err}
		}
		err = db.Ping()
		db.Close() // Always close the test connection
		if err != nil {
			return testConnectionResult{success: false, err: err}
		}
		return testConnectionResult{success: true, err: nil}
	}
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

		// Get tables list with metadata
		tableInfos, err := getTableInfos(db, m.selectedDB.driver)
		if err != nil {
			db.Close()
			return connectResult{err: err}
		}

		// Extract table names for backward compatibility
		tables := make([]string, len(tableInfos))
		for i, info := range tableInfos {
			tables[i] = info.name
		}

		return connectResult{db: db, tables: tables, tableInfos: tableInfos}
	}
}

// Command to load columns
func (m model) loadColumns() tea.Cmd {
	return func() tea.Msg {
		columns, err := getColumns(m.db, m.selectedDB.driver, m.selectedTable)
		if err != nil {
			return columnsResult{err: err}
		}
		return columnsResult{columns: columns}
	}
}

// Command to load data preview
func (m model) loadDataPreview() tea.Cmd {
	return func() tea.Msg {
		// Build query with proper table name quoting based on database type
		var query string
		switch m.selectedDB.driver {
		case "postgres":
			query = fmt.Sprintf(`SELECT * FROM "%s" LIMIT 10`, m.selectedTable)
		case "mysql":
			query = fmt.Sprintf("SELECT * FROM `%s` LIMIT 10", m.selectedTable)
		case "sqlite3":
			query = fmt.Sprintf(`SELECT * FROM "%s" LIMIT 10`, m.selectedTable)
		default:
			query = fmt.Sprintf("SELECT * FROM %s LIMIT 10", m.selectedTable)
		}
		rows, err := m.db.Query(query)
		if err != nil {
			return dataPreviewResult{err: err}
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

// Command to execute query
func (m model) executeQuery(query string) tea.Cmd {
	return func() tea.Msg {
		rows, err := m.db.Query(query)
		if err != nil {
			return queryResult{err: err}
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
	db         *sql.DB
	tables     []string
	tableInfos []tableInfo
	err        error
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

type clearResultMsg struct{}
type clearErrorMsg struct{}

// Command to clear query result after timeout
func clearResultAfterTimeout() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return clearResultMsg{}
	})
}

// Implement Update to handle results
func (m model) handleConnectResult(msg connectResult) (model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		// Start timeout to clear the error message
		return m, tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
			return clearErrorMsg{}
		})
	}

	m.isConnecting = false
	m.isSavingConnection = false
	m.db = msg.db
	m.tables = msg.tables
	m.tableInfos = msg.tableInfos
	m.err = nil
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
		m.err = msg.err
		// Start timeout to clear the error message
		return m, tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
			return clearErrorMsg{}
		})
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
		m.err = msg.err
		// Start timeout to clear the error message
		return m, tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
			return clearErrorMsg{}
		})
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
		m.err = msg.err
		// Start timeout to clear the error message
		return m, tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
			return clearErrorMsg{}
		})
	}

	m.err = nil

	// Set up table for query results
	if len(msg.rows) == 0 {
		m.queryResult = "Query executed successfully. No rows returned."
		// Clear the table
		m.queryResultsTable.SetColumns([]table.Column{})
		m.queryResultsTable.SetRows([]table.Row{})
		
		// Ensure query input has focus after query execution
		m.queryInput.Focus()
		m.queryResultsTable.Blur()
		
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

		return m, nil
	}
}

func (m model) handleDataPreviewResult(msg dataPreviewResult) (model, tea.Cmd) {
	m.isLoadingPreview = false
	if msg.err != nil {
		m.err = msg.err
		// Start timeout to clear the error message
		return m, tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
			return clearErrorMsg{}
		})
	}

	m.err = nil
	m.state = dataPreviewView

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

func getTableInfos(db *sql.DB, driver string) ([]tableInfo, error) {
	var tableInfos []tableInfo
	
	switch driver {
	case "postgres":
		query := `
			SELECT 
				t.tablename as table_name,
				'BASE TABLE' as table_type,
				COALESCE(s.n_tup_ins + s.n_tup_upd - s.n_tup_del, 0) as estimated_rows
			FROM pg_tables t
			LEFT JOIN pg_stat_user_tables s ON t.tablename = s.relname
			WHERE t.schemaname = 'public'
			ORDER BY t.tablename`
		
		rows, err := db.Query(query)
		if err != nil {
			// Fallback to simple table list if stats are not available
			return getSimpleTableInfos(db, driver)
		}
		defer rows.Close()
		
		for rows.Next() {
			var info tableInfo
			var estimatedRows sql.NullInt64
			err := rows.Scan(&info.name, &info.tableType, &estimatedRows)
			if err != nil {
				continue
			}
			
			if estimatedRows.Valid {
				info.rowCount = estimatedRows.Int64
				if info.rowCount < 0 {
					info.rowCount = 0
				}
				info.description = fmt.Sprintf("Table • ~%d rows", info.rowCount)
			} else {
				info.description = "Table"
			}
			
			tableInfos = append(tableInfos, info)
		}
		
	case "mysql":
		query := `
			SELECT 
				TABLE_NAME,
				TABLE_TYPE,
				COALESCE(TABLE_ROWS, 0) as table_rows
			FROM INFORMATION_SCHEMA.TABLES 
			WHERE TABLE_SCHEMA = DATABASE()
			ORDER BY TABLE_NAME`
		
		rows, err := db.Query(query)
		if err != nil {
			return getSimpleTableInfos(db, driver)
		}
		defer rows.Close()
		
		for rows.Next() {
			var info tableInfo
			var tableRows sql.NullInt64
			err := rows.Scan(&info.name, &info.tableType, &tableRows)
			if err != nil {
				continue
			}
			
			if tableRows.Valid && tableRows.Int64 > 0 {
				info.rowCount = tableRows.Int64
				info.description = fmt.Sprintf("Table • ~%d rows", info.rowCount)
			} else {
				info.description = "Table"
			}
			
			tableInfos = append(tableInfos, info)
		}
		
	case "sqlite3":
		// SQLite doesn't have built-in row count stats, so we'll get table names and count separately
		tables, err := getTables(db, driver)
		if err != nil {
			return nil, err
		}
		
		for _, tableName := range tables {
			info := tableInfo{
				name:      tableName,
				tableType: "table",
			}
			
			// Try to get row count (this might be slow for large tables)
			countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, tableName)
			var count int64
			err := db.QueryRow(countQuery).Scan(&count)
			if err == nil {
				info.rowCount = count
				info.description = fmt.Sprintf("Table • %d rows", count)
			} else {
				info.description = "Table"
			}
			
			tableInfos = append(tableInfos, info)
		}
		
	default:
		return getSimpleTableInfos(db, driver)
	}
	
	return tableInfos, nil
}

func getSimpleTableInfos(db *sql.DB, driver string) ([]tableInfo, error) {
	tables, err := getTables(db, driver)
	if err != nil {
		return nil, err
	}
	
	var tableInfos []tableInfo
	for _, tableName := range tables {
		tableInfos = append(tableInfos, tableInfo{
			name:        tableName,
			tableType:   "table",
			description: "Table",
		})
	}
	
	return tableInfos, nil
}

func getColumns(db *sql.DB, driver, tableName string) ([][]string, error) {
	var query string
	switch driver {
	case "postgres":
		query = `SELECT column_name, data_type, is_nullable, column_default 
				 FROM information_schema.columns 
				 WHERE table_name = $1 
				 ORDER BY ordinal_position`
	case "mysql":
		query = `SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT 
				 FROM INFORMATION_SCHEMA.COLUMNS 
				 WHERE TABLE_NAME = ? 
				 ORDER BY ORDINAL_POSITION`
	case "sqlite3":
		query = fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	}

	rows, err := db.Query(query, tableName)
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

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
