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

// Global styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1).
			Bold(true)

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#01BE85")).
			Background(lipgloss.Color("#00432F")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Margin(1, 0)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF87")).
			Bold(true)

	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

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
	selectedTable        string
	savedConnections     []SavedConnection
	editingConnectionIdx int
	queryResult          string
	width                int
	height               int
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

	// Load saved connections
	savedConnections, _ := loadSavedConnections()

	// Saved connections list
	savedConnectionsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 20)
	savedConnectionsList.Title = "💾 Saved Connections"
	savedConnectionsList.SetShowStatusBar(false)
	savedConnectionsList.SetFilteringEnabled(false)

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

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	// Query results table
	queryResultsTable := table.New(
		table.WithColumns([]table.Column{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	queryResultsTable.SetStyles(s)

	// Data preview table
	dataPreviewTable := table.New(
		table.WithColumns([]table.Column{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	dataPreviewTable.SetStyles(s)

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
			// Apply default styles
			s := table.DefaultStyles()
			s.Header = s.Header.
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				BorderBottom(true).
				Bold(true)
			s.Selected = s.Selected.
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Bold(false)
			m.queryResultsTable.SetStyles(s)
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
				m.state = connectionView
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
			if (m.state == connectionView || m.state == tablesView || m.state == columnsView) && m.connectionStr != "" {
				// Go to connection naming view
				m.state = saveConnectionView
				m.nameInput.SetValue("")
				m.nameInput.Focus()
				return m, nil
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
			if m.state == tablesView && m.db != nil {
				// Preview table data
				if i, ok := m.tablesList.SelectedItem().(item); ok {
					m.selectedTable = i.title
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
					m.textInput.SetValue("")
					m.textInput.Focus()

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
				if i, ok := m.savedConnectionsList.SelectedItem().(item); ok {
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
							return m, m.connectDB()
						}
					}
				}

			case connectionView:
				m.connectionStr = m.textInput.Value()
				if m.connectionStr != "" {
					return m, m.connectDB()
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
				if i, ok := m.tablesList.SelectedItem().(item); ok {
					m.selectedTable = i.title
					return m, m.loadColumns()
				}

			case queryView:
				query := m.queryInput.Value()
				if query != "" {
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
		m.textInput, cmd = m.textInput.Update(msg)
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
	content := titleStyle.Render("Select database type") + "\n\n"
	content += m.dbTypeList.View()
	content += "\n" + helpStyle.Render("↑/↓: navigate • enter: select • s: saved connections • q: quit")
	return docStyle.Render(content)
}

func (m model) savedConnectionsView() string {
	content := titleStyle.Render("Saved Connections") + "\n\n"

	if len(m.savedConnections) == 0 {
		content += helpStyle.Render("No saved connections yet.")
	} else {
		content += successStyle.Render(fmt.Sprintf("✅ %d connections available", len(m.savedConnections))) + "\n\n"
		content += m.savedConnectionsList.View()
	}

	content += "\n" + helpStyle.Render("↑/↓: navigate • enter: connect • e: edit • d: delete • esc: back")
	return docStyle.Render(content)
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
	content := titleStyle.Render(fmt.Sprintf("Connect to %s", m.selectedDB.name)) + "\n\n"

	if m.err != nil {
		content += errorStyle.Render("❌ Error: "+m.err.Error()) + "\n\n"
	}

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

	content += "\n\n" + helpStyle.Render("enter: connect • s: name and save connection • esc: back")
	return docStyle.Render(content)
}

func (m model) tablesView() string {
	content := titleStyle.Render(fmt.Sprintf("Tables in %s", m.selectedDB.name)) + "\n\n"

	if len(m.tables) == 0 {
		content += helpStyle.Render("No tables found in this database.")
	} else {
		content += successStyle.Render(fmt.Sprintf("✅ Connected successfully (%d tables found)", len(m.tables))) + "\n\n"
		content += m.tablesList.View()
	}

	content += "\n" + helpStyle.Render("↑/↓: navigate • enter: view columns • p: preview data • r: run query • s: save connection • esc: back")
	return docStyle.Render(content)
}

func (m model) columnsView() string {
	content := titleStyle.Render(fmt.Sprintf("Columns of table: %s", m.selectedTable)) + "\n\n"
	content += m.columnsTable.View()
	content += "\n" + helpStyle.Render("↑/↓: navigate • r: run query • s: save connection • esc: back to tables")
	return docStyle.Render(content)
}

func (m model) queryView() string {
	content := titleStyle.Render("SQL Query") + "\n\n"

	if m.err != nil {
		content += errorStyle.Render("❌ Error: "+m.err.Error()) + "\n\n"
	}

	content += "Enter SQL query:\n"
	content += m.queryInput.View() + "\n\n"

	if m.queryResult != "" {
		content += "Query Result:\n"
		content += successStyle.Render(m.queryResult) + "\n\n"

		// Show the table if there are results
		if len(m.queryResultsTable.Rows()) > 0 {
			content += m.queryResultsTable.View() + "\n\n"
		}
	}

	content += helpStyle.Render("Examples:") + "\n"
	content += helpStyle.Render("  SELECT * FROM users LIMIT 10;") + "\n"
	content += helpStyle.Render("  INSERT INTO users (name, email) VALUES ('John', 'john@example.com');") + "\n"
	content += helpStyle.Render("  UPDATE users SET email = 'new@example.com' WHERE id = 1;") + "\n"
	content += helpStyle.Render("  DELETE FROM users WHERE id = 1;")
	content += "\n\n" + helpStyle.Render("enter: execute query • tab: switch focus • ↑/↓: navigate results • esc: back to tables")
	return docStyle.Render(content)
}

func (m model) dataPreviewView() string {
	content := titleStyle.Render(fmt.Sprintf("Data Preview: %s", m.selectedTable)) + "\n\n"

	if m.err != nil {
		content += errorStyle.Render("❌ Error: "+m.err.Error()) + "\n\n"
	}

	if len(m.dataPreviewTable.Rows()) > 0 {
		content += successStyle.Render(fmt.Sprintf("Showing first 10 rows from %s", m.selectedTable)) + "\n\n"
		content += m.dataPreviewTable.View()
	} else if m.err == nil {
		content += helpStyle.Render("Loading data preview...")
	}

	content += "\n\n" + helpStyle.Render("↑/↓: navigate rows • esc: back to tables")
	return docStyle.Render(content)
}

// Command to connect to database
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

		// Get tables list
		tables, err := getTables(db, m.selectedDB.driver)
		if err != nil {
			db.Close()
			return connectResult{err: err}
		}

		return connectResult{db: db, tables: tables}
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
	db     *sql.DB
	tables []string
	err    error
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

	m.db = msg.db
	m.tables = msg.tables
	m.err = nil
	m.state = tablesView

	// Create items for tables list
	items := make([]list.Item, len(msg.tables))
	for i, table := range msg.tables {
		items[i] = item{
			title: table,
			desc:  "View table structure",
		}
	}
	m.tablesList.SetItems(items)

	return m, nil
}

func (m model) handleColumnsResult(msg columnsResult) (model, tea.Cmd) {
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

		// Create rows for the table with proper column count validation
		rows := make([]table.Row, 0, len(msg.rows))
		expectedColumnCount := len(msg.columns)
		
		for _, row := range msg.rows {
			// Skip rows that don't match the expected column count
			if len(row) != expectedColumnCount {
				continue
			}
			
			tableRow := make(table.Row, expectedColumnCount)
			for j, val := range row {
				// Double-check bounds to prevent panic
				if j >= expectedColumnCount {
					break
				}
				// Truncate long values for display
				if len(val) > 23 {
					val = val[:20] + "..."
				}
				tableRow[j] = val
			}
			rows = append(rows, tableRow)
		}

		m.queryResultsTable.SetColumns(columns)
		m.queryResultsTable.SetRows(rows)
		m.queryResult = fmt.Sprintf("Query returned %d rows", len(rows))
		
		// Ensure query input has focus after query execution
		m.queryInput.Focus()
		m.queryResultsTable.Blur()

		return m, nil
	}
}

func (m model) handleDataPreviewResult(msg dataPreviewResult) (model, tea.Cmd) {
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

	// Create rows for the table with proper column count validation
	rows := make([]table.Row, 0, len(msg.rows))
	expectedColumnCount := len(msg.columns)
	
	for _, row := range msg.rows {
		// Skip rows that don't match the expected column count
		if len(row) != expectedColumnCount {
			// Log the issue but continue with other rows
			continue
		}
		
		tableRow := make(table.Row, expectedColumnCount)
		for j, val := range row {
			// Double-check bounds to prevent panic
			if j >= expectedColumnCount {
				break
			}
			// Truncate long values for display
			if len(val) > 23 {
				val = val[:20] + "..."
			}
			tableRow[j] = val
		}
		rows = append(rows, tableRow)
	}

	// Set columns first, then rows
	m.dataPreviewTable.SetColumns(columns)
	m.dataPreviewTable.SetRows(rows)

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
