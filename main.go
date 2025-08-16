package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/danielcaldera/dbx/internal/config"
	"github.com/danielcaldera/dbx/internal/database"
	"github.com/danielcaldera/dbx/internal/models"
	"github.com/danielcaldera/dbx/internal/styles"
	"github.com/danielcaldera/dbx/internal/views"
)

// Database types
var dbTypes = []models.DBType{
	{"PostgreSQL", "postgres"},
	{"MySQL", "mysql"},
	{"SQLite", "sqlite3"},
}

func initialModel() models.Model {
	// Database types list
	items := make([]list.Item, len(dbTypes))
	for i, db := range dbTypes {
		items[i] = models.Item{
			ItemTitle: db.Name,
			ItemDesc:  fmt.Sprintf("Connect to %s database", db.Name),
		}
	}

    dbList := list.New(items, list.NewDefaultDelegate(), 0, 0)
    dbList.Title = "🗄️ DBX — Database Explorer"
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
    savedConnectionsList.Title = "💾 Saved Connections"
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

	// Tables list
    tablesList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
    tablesList.Title = "📊 Available Tables"
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

	// Initialize textarea for field editing
	ta := textarea.New()
	ta.Placeholder = "Enter field content..."
	ta.SetWidth(80)
	ta.SetHeight(20)

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
		SelectedSchema:          "public", // Default to public schema for PostgreSQL
		SavedConnections:        savedConnections,
		QueryHistory:            queryHistory,
		EditingConnectionIdx:    -1,
		RowDetailItemsPerPage:   8,   // Show 8 fields per page
		FullTextLinesPerPage:    20,  // Show 20 lines per page in full text view
		FullTextItemsPerPage:    5,   // Show 5 fields per page in full text view
		FieldDetailLinesPerPage: 25,  // Show 25 lines per page in field detail view
		FieldDetailCharsPerLine: 120, // Show 120 characters per line in field detail view
		FieldTextarea:           ta,  // Initialize textarea for field editing
	}

	return m
}

// Wrapper type to add methods to the imported Model
type appModel struct {
	models.Model
}

func (m appModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, textarea.Blink)
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle basic message types
	switch msg := msg.(type) {
	case models.ConnectResult:
		return handleConnectResult(m, msg)
	case models.TestConnectionResult:
		return handleTestConnectionResult(m, msg)
    case models.ColumnsResult:
        return handleColumnsResult(m, msg)
    case models.DataPreviewResult:
        return handleDataPreviewResult(m, msg)
    case models.RelationshipsResult:
        return handleRelationshipsResult(m, msg)
	case models.ClearResultMsg:
		m.QueryResult = ""
		return m, nil
	case models.ClearErrorMsg:
		m.Err = nil
		return m, nil
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		h, v := styles.DocStyle.GetFrameSize()
		m.DBTypeList.SetSize(msg.Width-h, msg.Height-v-5)
		m.SavedConnectionsList.SetSize(msg.Width-h, msg.Height-v-5)
		m.TablesList.SetSize(msg.Width-h, msg.Height-v-5)
		m.TextInput.Width = msg.Width - h - 4
		m.NameInput.Width = msg.Width - h - 4
		m.QueryInput.Width = msg.Width - h - 4
		m.SearchInput.Width = msg.Width - h - 4

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.DB != nil {
				m.DB.Close()
			}
			return m, tea.Quit

		case "q":
			if m.State == models.DBTypeView {
				if m.DB != nil {
					m.DB.Close()
				}
				return m, tea.Quit
			}

        case "esc":
            switch m.State {
			case models.SavedConnectionsView:
				m.State = models.DBTypeView
				m.Err = nil
			case models.ConnectionView:
				m.State = models.DBTypeView
				m.Err = nil
			case models.SaveConnectionView:
				m.State = models.ConnectionView
				m.Err = nil
			case models.EditConnectionView:
				m.State = models.SavedConnectionsView
				m.Err = nil
				m.EditingConnectionIdx = -1
			case models.TablesView:
				if m.DB != nil {
					m.DB.Close()
					m.DB = nil
				}
				m.State = models.DBTypeView
				m.ConnectionStr = ""
				m.Tables = nil
				m.TableInfos = nil
				m.SelectedTable = ""
				m.Err = nil
            case models.ColumnsView:
                m.State = models.TablesView
            case models.DataPreviewView:
                m.State = models.TablesView
            case models.RelationshipsView:
                m.State = models.TablesView
            }
            return m, nil

		case "s":
			if m.State == models.DBTypeView {
				m.State = models.SavedConnectionsView
				if connections, err := config.LoadSavedConnections(); err == nil {
					m.SavedConnections = connections
				}
				m = updateSavedConnectionsList(m)
				return m, nil
			}
			if m.State == models.ColumnsView && m.ConnectionStr != "" {
				m.State = models.SaveConnectionView
				m.NameInput.SetValue("")
				m.NameInput.Focus()
				return m, nil
			}

		case "f1":
			if m.State == models.ConnectionView && !m.IsTestingConnection {
				m.ConnectionStr = m.TextInput.Value()
				if m.ConnectionStr != "" {
					m.IsTestingConnection = true
					m.Err = nil
					m.QueryResult = ""
					return m, testConnection(m)
				}
			}

        case "enter":
            switch m.State {
			case models.DBTypeView:
				if i, ok := m.DBTypeList.SelectedItem().(models.Item); ok {
					for _, db := range dbTypes {
						if db.Name == i.ItemTitle {
							m.SelectedDB = db
							break
						}
					}
					m.State = models.ConnectionView
					m.NameInput.SetValue("")
					m.TextInput.SetValue("")
					m.TextInput.Blur()
					m.NameInput.Focus()

					// Set placeholder according to DB type
					switch m.SelectedDB.Driver {
					case "postgres":
						m.TextInput.Placeholder = "postgres://user:password@localhost/dbname?sslmode=disable"
					case "mysql":
						m.TextInput.Placeholder = "user:password@tcp(localhost:3306)/dbname"
					case "sqlite3":
						m.TextInput.Placeholder = "/path/to/database.db"
					}
				}

			case models.SavedConnectionsView:
				if i, ok := m.SavedConnectionsList.SelectedItem().(models.Item); ok && !m.IsConnecting {
					for _, conn := range m.SavedConnections {
						if conn.Name == i.ItemTitle {
							for _, db := range dbTypes {
								if db.Driver == conn.Driver {
									m.SelectedDB = db
									break
								}
							}
							m.ConnectionStr = conn.ConnectionStr
							m.IsConnecting = true
							m.Err = nil
							return m, connectDB(m)
						}
					}
				}

			case models.SaveConnectionView:
				name := m.NameInput.Value()
				if name != "" {
					newConnection := models.SavedConnection{
						Name:          name,
						Driver:        m.SelectedDB.Driver,
						ConnectionStr: m.ConnectionStr,
					}
					m.SavedConnections = append(m.SavedConnections, newConnection)
					config.SaveConnections(m.SavedConnections)
					m.State = models.ConnectionView
					return m, nil
				}

            case models.TablesView:
                if i, ok := m.TablesList.SelectedItem().(models.Item); ok && !m.IsLoadingPreview {
                    m.SelectedTable = i.ItemTitle
                    m.IsLoadingPreview = true
                    m.Err = nil
                    return m, loadDataPreview(m)
                }
            }
        case "p":
            // Keep "p" as an alias for preview
            if m.State == models.TablesView {
                if i, ok := m.TablesList.SelectedItem().(models.Item); ok && !m.IsLoadingPreview {
                    m.SelectedTable = i.ItemTitle
                    m.IsLoadingPreview = true
                    m.Err = nil
                    return m, loadDataPreview(m)
                }
            }
        case "v":
            if m.State == models.TablesView {
                if i, ok := m.TablesList.SelectedItem().(models.Item); ok && !m.IsLoadingColumns {
                    m.SelectedTable = i.ItemTitle
                    m.IsLoadingColumns = true
                    m.Err = nil
                    return m, loadColumns(m)
                }
            }
		case "f":
			if m.State == models.TablesView && m.DB != nil {
				return m, loadRelationships(m)
			}
		}
	}

	// Update components according to state
	switch m.State {
	case models.DBTypeView:
		m.DBTypeList, cmd = m.DBTypeList.Update(msg)
	case models.SavedConnectionsView:
		m.SavedConnectionsList, cmd = m.SavedConnectionsList.Update(msg)
	case models.ConnectionView:
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "tab":
				if m.NameInput.Focused() {
					m.NameInput.Blur()
					m.TextInput.Focus()
				} else {
					m.TextInput.Blur()
					m.NameInput.Focus()
				}
				return m, nil
			}
		}

		if m.NameInput.Focused() {
			m.NameInput, cmd = m.NameInput.Update(msg)
		} else {
			m.TextInput, cmd = m.TextInput.Update(msg)
		}
	case models.SaveConnectionView:
		m.NameInput, cmd = m.NameInput.Update(msg)
    case models.TablesView:
        m.TablesList, cmd = m.TablesList.Update(msg)
    case models.ColumnsView:
        m.ColumnsTable, cmd = m.ColumnsTable.Update(msg)
    case models.DataPreviewView:
        if km, ok := msg.(tea.KeyMsg); ok {
            switch km.String() {
            case "r":
                return m, loadDataPreview(m)
            }
        }
        m.DataPreviewTable, cmd = m.DataPreviewTable.Update(msg)
    }

	return m, cmd
}

func (m appModel) View() string {
	switch m.State {
	case models.DBTypeView:
		return views.DBTypeView(m.Model)
	case models.SavedConnectionsView:
		return views.SavedConnectionsView(m.Model)
	case models.ConnectionView:
		return views.ConnectionView(m.Model)
	case models.SaveConnectionView:
		return views.SaveConnectionView(m.Model)
	case models.TablesView:
		return views.TablesView(m.Model)
	case models.DataPreviewView:
		return views.DataPreviewView(m.Model)
	case models.RelationshipsView:
		return views.RelationshipsView(m.Model)
	case models.ColumnsView:
		return views.ColumnsView(m.Model)
	default:
		return "View not implemented yet"
	}
}

// Helper functions

func testConnection(m appModel) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		return database.TestConnectionWithTimeout(m.SelectedDB.Driver, m.ConnectionStr)
	})
}

func handleTestConnectionResult(m appModel, msg models.TestConnectionResult) (appModel, tea.Cmd) {
	m.IsTestingConnection = false
	if msg.Success {
		m.QueryResult = "✅ Connection successful!"
	} else {
		m.Err = fmt.Errorf("connection failed: %v", msg.Err)
	}

	return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return models.ClearResultMsg{}
	})
}

func connectDB(m appModel) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		db, err := sql.Open(m.SelectedDB.Driver, m.ConnectionStr)
		if err != nil {
			return models.ConnectResult{Err: err}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = db.PingContext(ctx)
		if err != nil {
			db.Close()
			return models.ConnectResult{Err: err}
		}

		tables, err := database.GetTables(db, m.SelectedDB.Driver)
		if err != nil {
			db.Close()
			return models.ConnectResult{Err: err}
		}

		schema := "public"
		if m.SelectedDB.Driver == "mysql" {
			schema = "mysql"
		} else if m.SelectedDB.Driver == "sqlite3" {
			schema = "main"
		}

		return models.ConnectResult{
			DB:     db,
			Driver: m.SelectedDB.Driver,
			Tables: tables,
			Schema: schema,
		}
	})
}

func handleConnectResult(m appModel, msg models.ConnectResult) (appModel, tea.Cmd) {
	m.IsConnecting = false

	if msg.Err != nil {
		m.Err = msg.Err
		return m, nil
	}

	m.DB = msg.DB
	m.Tables = msg.Tables
	m.SelectedSchema = msg.Schema

	// Create simple table infos
	m.TableInfos = make([]models.TableInfo, len(m.Tables))
	for i, tableName := range m.Tables {
		m.TableInfos[i] = models.TableInfo{
			Name:        tableName,
			Schema:      m.SelectedSchema,
			TableType:   "BASE TABLE",
			Description: "📊 Table",
		}
	}

	// Update tables list
	items := make([]list.Item, len(m.TableInfos))
	for i, info := range m.TableInfos {
		items[i] = models.Item{
			ItemTitle: info.Name,
			ItemDesc:  info.Description,
		}
	}
	m.TablesList.SetItems(items)

	m.State = models.TablesView
	return m, nil
}

func loadColumns(m appModel) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		columns, err := database.GetColumns(m.DB, m.SelectedDB.Driver, m.SelectedTable, m.SelectedSchema)
		return models.ColumnsResult{
			Columns: columns,
			Err:     err,
		}
	})
}

func handleColumnsResult(m appModel, msg models.ColumnsResult) (appModel, tea.Cmd) {
	m.IsLoadingColumns = false

	if msg.Err != nil {
		m.Err = msg.Err
		return m, nil
	}

	// Convert to table rows
	rows := make([]table.Row, len(msg.Columns))
	for i, col := range msg.Columns {
		rows[i] = table.Row{col[0], col[1], col[2], col[3]}
	}

	m.ColumnsTable.SetRows(rows)
	m.State = models.ColumnsView

	return m, nil
}

func loadDataPreview(m appModel) tea.Cmd {
    return tea.Cmd(func() tea.Msg {
        cols, rows, err := database.GetTablePreview(m.DB, m.SelectedDB.Driver, m.SelectedTable, m.SelectedSchema, 10)
        return models.DataPreviewResult{Columns: cols, Rows: rows, Err: err}
    })
}

func handleDataPreviewResult(m appModel, msg models.DataPreviewResult) (appModel, tea.Cmd) {
    m.IsLoadingPreview = false
    if msg.Err != nil {
        m.Err = msg.Err
        return m, nil
    }
    // Build columns meta
    cols := make([]table.Column, len(msg.Columns))
    for i, c := range msg.Columns {
        cols[i] = table.Column{Title: c, Width: 16}
    }
    // Build rows
    rows := make([]table.Row, len(msg.Rows))
    for i, r := range msg.Rows {
        tr := make(table.Row, len(r))
        copy(tr, r)
        rows[i] = tr
    }
    // Recreate table with new columns/rows
    m.DataPreviewTable = table.New(
        table.WithColumns(cols),
        table.WithRows(rows),
        table.WithFocused(true),
        table.WithHeight(10),
    )
    m.DataPreviewTable.SetStyles(styles.GetMagentaTableStyles())
    m.State = models.DataPreviewView
    return m, nil
}

func loadRelationships(m appModel) tea.Cmd {
    return tea.Cmd(func() tea.Msg {
        rels, err := database.GetForeignKeyRelationships(m.DB, m.SelectedDB.Driver, m.SelectedSchema)
        return models.RelationshipsResult{Relationships: rels, Err: err}
    })
}

func handleRelationshipsResult(m appModel, msg models.RelationshipsResult) (appModel, tea.Cmd) {
    if msg.Err != nil {
        m.Err = msg.Err
        return m, nil
    }
    // columns: From Table, From Column, To Table, To Column, Constraint Name
    cols := []table.Column{
        {Title: "From Table", Width: 20},
        {Title: "From Column", Width: 20},
        {Title: "To Table", Width: 20},
        {Title: "To Column", Width: 20},
        {Title: "Constraint Name", Width: 25},
    }
    rows := make([]table.Row, len(msg.Relationships))
    for i, rel := range msg.Relationships {
        rows[i] = table.Row(rel)
    }
    m.RelationshipsTable = table.New(
        table.WithColumns(cols),
        table.WithRows(rows),
        table.WithFocused(true),
        table.WithHeight(10),
    )
    m.RelationshipsTable.SetStyles(styles.GetMagentaTableStyles())
    m.State = models.RelationshipsView
    return m, nil
}

// Helper functions

func updateSavedConnectionsList(m appModel) appModel {
	savedItems := make([]list.Item, len(m.SavedConnections))
	for i, conn := range m.SavedConnections {
		connStr := conn.ConnectionStr
		if len(connStr) > 50 {
			connStr = connStr[:50] + "..."
		}
		savedItems[i] = models.Item{
			ItemTitle: conn.Name,
			ItemDesc:  fmt.Sprintf("%s - %s", conn.Driver, connStr),
		}
	}
	m.SavedConnectionsList.SetItems(savedItems)
	return m
}

func main() {
	m := appModel{Model: initialModel()}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
