package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Estilos globales
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

// Estados de la aplicación
type viewState int

const (
	dbTypeView viewState = iota
	savedConnectionsView
	connectionView
	saveConnectionView
	tablesView
	columnsView
)

// Conexión guardada
type SavedConnection struct {
	Name          string `json:"name"`
	Driver        string `json:"driver"`
	ConnectionStr string `json:"connection_str"`
}

// Tipos de base de datos
type dbType struct {
	name   string
	driver string
}

var dbTypes = []dbType{
	{"PostgreSQL", "postgres"},
	{"MySQL", "mysql"},
	{"SQLite", "sqlite3"},
}

// Item para la lista
type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// Modelo principal
type model struct {
	state               viewState
	dbTypeList          list.Model
	savedConnectionsList list.Model
	textInput           textinput.Model
	nameInput           textinput.Model
	tablesList          list.Model
	columnsTable        table.Model
	selectedDB          dbType
	connectionStr       string
	db                  *sql.DB
	err                 error
	tables              []string
	selectedTable       string
	savedConnections    []SavedConnection
	width               int
	height              int
}

func initialModel() model {
	// Lista de tipos de base de datos
	items := make([]list.Item, len(dbTypes))
	for i, db := range dbTypes {
		items[i] = item{
			title: db.name,
			desc:  fmt.Sprintf("Conectar a base de datos %s", db.name),
		}
	}

	dbList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	dbList.Title = "🗄️  DBX - Database Explorer"
	dbList.SetShowStatusBar(false)
	dbList.SetFilteringEnabled(false)

	// Cargar conexiones guardadas
	savedConnections, _ := loadSavedConnections()
	
	// Lista de conexiones guardadas
	savedConnectionsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 20)
	savedConnectionsList.Title = "💾 Conexiones Guardadas"
	savedConnectionsList.SetShowStatusBar(false)
	savedConnectionsList.SetFilteringEnabled(false)
	
	// Poblar la lista con las conexiones guardadas
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

	// Input de conexión
	ti := textinput.New()
	ti.Placeholder = "Ingresa la cadena de conexión..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 80

	// Input para nombre de conexión
	ni := textinput.New()
	ni.Placeholder = "Nombre para esta conexión..."
	ni.CharLimit = 100
	ni.Width = 80

	// Lista de tablas
	tablesList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	tablesList.Title = "📊 Tablas disponibles"

	// Tabla de columnas
	columns := []table.Column{
		{Title: "Columna", Width: 20},
		{Title: "Tipo", Width: 15},
		{Title: "Nulo", Width: 8},
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

	return model{
		state:               dbTypeView,
		dbTypeList:          dbList,
		savedConnectionsList: savedConnectionsList,
		textInput:           ti,
		nameInput:           ni,
		tablesList:          tablesList,
		columnsTable:        t,
		savedConnections:    savedConnections,
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.dbTypeList.SetSize(msg.Width-h, msg.Height-v-5)
		m.savedConnectionsList.SetSize(msg.Width-h, msg.Height-v-5)
		m.tablesList.SetSize(msg.Width-h, msg.Height-v-5)
		m.textInput.Width = msg.Width - h - 4
		m.nameInput.Width = msg.Width - h - 4

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.db != nil {
				m.db.Close()
			}
			return m, tea.Quit

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
			case tablesView:
				m.state = connectionView
			case columnsView:
				m.state = tablesView
			}
			return m, nil

		case "s":
			if m.state == dbTypeView {
				m.state = savedConnectionsView
				// Recargar conexiones desde archivo
				if connections, err := loadSavedConnections(); err == nil {
					m.savedConnections = connections
				}
				m = m.updateSavedConnectionsList()
				return m, nil
			}
			if (m.state == connectionView || m.state == tablesView || m.state == columnsView) && m.connectionStr != "" {
				// Ir a la vista de nombrar conexión
				m.state = saveConnectionView
				m.nameInput.SetValue("")
				m.nameInput.Focus()
				return m, nil
			}

		case "n":
			if m.state == tablesView || m.state == columnsView {
				if m.db != nil {
					m.db.Close()
				}
				m.state = dbTypeView
				m.err = nil
				m.connectionStr = ""
				m.tables = nil
				m.selectedTable = ""
			}
			return m, nil

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

					// Establecer placeholder según el tipo de DB
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
					// Buscar la conexión guardada por nombre
					for _, conn := range m.savedConnections {
						if conn.Name == i.title {
							// Establecer el tipo de DB basado en el driver
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
					// Guardar conexión con el nombre proporcionado
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

			case tablesView:
				if i, ok := m.tablesList.SelectedItem().(item); ok {
					m.selectedTable = i.title
					return m, m.loadColumns()
				}
			}
		}
	}

	// Actualizar componentes según el estado
	switch m.state {
	case dbTypeView:
		m.dbTypeList, cmd = m.dbTypeList.Update(msg)
	case savedConnectionsView:
		m.savedConnectionsList, cmd = m.savedConnectionsList.Update(msg)
	case connectionView:
		m.textInput, cmd = m.textInput.Update(msg)
	case saveConnectionView:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case tablesView:
		m.tablesList, cmd = m.tablesList.Update(msg)
	case columnsView:
		m.columnsTable, cmd = m.columnsTable.Update(msg)
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
	case tablesView:
		return m.tablesView()
	case columnsView:
		return m.columnsView()
	}
	return ""
}

func (m model) dbTypeView() string {
	content := titleStyle.Render("Selecciona el tipo de base de datos") + "\n\n"
	content += m.dbTypeList.View()
	content += "\n" + helpStyle.Render("↑/↓: navegar • enter: seleccionar • s: conexiones guardadas • q: salir")
	return docStyle.Render(content)
}

func (m model) savedConnectionsView() string {
	content := titleStyle.Render("Conexiones Guardadas") + "\n\n"
	
	if len(m.savedConnections) == 0 {
		content += helpStyle.Render("No hay conexiones guardadas aún.")
	} else {
		content += successStyle.Render(fmt.Sprintf("✅ %d conexiones disponibles", len(m.savedConnections))) + "\n\n"
		content += m.savedConnectionsList.View()
	}
	
	content += "\n" + helpStyle.Render("↑/↓: navegar • enter: conectar • esc: volver • q: salir")
	return docStyle.Render(content)
}

func (m model) saveConnectionView() string {
	content := titleStyle.Render("Guardar Conexión") + "\n\n"
	content += "Nombre para esta conexión:\n"
	content += m.nameInput.View() + "\n\n"
	content += "Conexión a guardar:\n"
	content += helpStyle.Render(fmt.Sprintf("%s: %s", m.selectedDB.name, m.connectionStr))
	content += "\n\n" + helpStyle.Render("enter: guardar • esc: cancelar • q: salir")
	return docStyle.Render(content)
}

func (m model) connectionView() string {
	content := titleStyle.Render(fmt.Sprintf("Conectar a %s", m.selectedDB.name)) + "\n\n"

	if m.err != nil {
		content += errorStyle.Render("❌ Error: "+m.err.Error()) + "\n\n"
	}

	content += "Cadena de conexión:\n"
	content += m.textInput.View() + "\n\n"

	content += "Ejemplos:\n"
	switch m.selectedDB.driver {
	case "postgres":
		content += helpStyle.Render("postgres://user:password@localhost/dbname?sslmode=disable")
	case "mysql":
		content += helpStyle.Render("user:password@tcp(localhost:3306)/dbname")
	case "sqlite3":
		content += helpStyle.Render("./database.db o /path/to/database.db")
	}

	content += "\n\n" + helpStyle.Render("enter: conectar • s: nombrar y guardar conexión • esc: volver • q: salir")
	return docStyle.Render(content)
}

func (m model) tablesView() string {
	content := titleStyle.Render(fmt.Sprintf("Tablas en %s", m.selectedDB.name)) + "\n\n"

	if len(m.tables) == 0 {
		content += helpStyle.Render("No se encontraron tablas en esta base de datos.")
	} else {
		content += successStyle.Render(fmt.Sprintf("✅ Conectado exitosamente (%d tablas encontradas)", len(m.tables))) + "\n\n"
		content += m.tablesList.View()
	}

	content += "\n" + helpStyle.Render("↑/↓: navegar • enter: ver columnas • s: guardar conexión • n: nueva conexión • esc: volver • q: salir")
	return docStyle.Render(content)
}

func (m model) columnsView() string {
	content := titleStyle.Render(fmt.Sprintf("Columnas de la tabla: %s", m.selectedTable)) + "\n\n"
	content += m.columnsTable.View()
	content += "\n" + helpStyle.Render("↑/↓: navegar • s: guardar conexión • n: nueva conexión • esc: volver a tablas • q: salir")
	return docStyle.Render(content)
}

// Comando para conectar a la base de datos
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

		// Obtener lista de tablas
		tables, err := getTables(db, m.selectedDB.driver)
		if err != nil {
			db.Close()
			return connectResult{err: err}
		}

		return connectResult{db: db, tables: tables}
	}
}

// Comando para cargar columnas
func (m model) loadColumns() tea.Cmd {
	return func() tea.Msg {
		columns, err := getColumns(m.db, m.selectedDB.driver, m.selectedTable)
		if err != nil {
			return columnsResult{err: err}
		}
		return columnsResult{columns: columns}
	}
}

// Mensajes de resultado
type connectResult struct {
	db     *sql.DB
	tables []string
	err    error
}

type columnsResult struct {
	columns [][]string
	err     error
}

// Implementar Update para manejar resultados
func (m model) handleConnectResult(msg connectResult) (model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		return m, nil
	}

	m.db = msg.db
	m.tables = msg.tables
	m.err = nil
	m.state = tablesView

	// Crear items para la lista de tablas
	items := make([]list.Item, len(msg.tables))
	for i, table := range msg.tables {
		items[i] = item{
			title: table,
			desc:  "Ver estructura de la tabla",
		}
	}
	m.tablesList.SetItems(items)

	return m, nil
}

func (m model) handleColumnsResult(msg columnsResult) (model, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		return m, nil
	}

	m.err = nil
	m.state = columnsView

	// Convertir a filas de tabla
	rows := make([]table.Row, len(msg.columns))
	for i, col := range msg.columns {
		rows[i] = table.Row(col)
	}
	m.columnsTable.SetRows(rows)

	return m, nil
}


// Funciones auxiliares para obtener información de la DB

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

// Método para actualizar la lista de conexiones guardadas
func (m model) updateSavedConnectionsList() model {
	items := make([]list.Item, len(m.savedConnections))
	for i, conn := range m.savedConnections {
		// Truncar la cadena de conexión de forma segura
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


// Funciones para manejo de conexiones guardadas

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
