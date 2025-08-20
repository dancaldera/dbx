package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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

	dbList := list.New(items, styles.GetBlueListDelegate(), 0, 0)
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
	savedConnectionsList := list.New([]list.Item{}, styles.GetBlueListDelegate(), 50, 20)
	savedConnectionsList.Title = "Saved Connections"
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
	tblDelegate := styles.GetBlueListDelegate()
	tblDelegate.ShowDescription = false
	tblDelegate.SetSpacing(0)
	tablesList := list.New([]list.Item{}, tblDelegate, 0, 0)
	tablesList.Title = "Available Tables"
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

	t.SetStyles(styles.GetBlueTableStyles())

	// Query results table
	queryResultsTable := table.New(
		table.WithColumns([]table.Column{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	queryResultsTable.SetStyles(styles.GetBlueTableStyles())

	// Initialize textarea for field editing
	ta := textarea.New()
	ta.Placeholder = "Enter field content..."
	ta.SetWidth(100) // Will be dynamically resized
	ta.SetHeight(20) // Will be dynamically resized
	ta.ShowLineNumbers = true

	// Initialize filter input
	filterInput := textinput.New()
	filterInput.Placeholder = "Type to filter all columns..."
	filterInput.Width = 60

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
		FullTextItemsPerPage:    5,           // Show 5 fields per page in full text view
		FieldDetailLinesPerPage: 25,          // Show 25 lines per page in field detail view
		FieldDetailCharsPerLine: 120,         // Show 120 characters per line in field detail view
		FieldTextarea:           ta,          // Initialize textarea for field editing
		DataPreviewCurrentPage:  0,           // Start at first page
		DataPreviewItemsPerPage: 40,          // Show 40 items per page
		DataPreviewTotalRows:    0,           // Will be set when loading data
		DataPreviewScrollOffset: 0,           // Start at first column
		DataPreviewVisibleCols:  6,           // Show 6 columns at once
		DataPreviewFilterActive: false,       // Start without filter
		DataPreviewFilterValue:  "",          // No initial filter
		DataPreviewFilterInput:  filterInput, // Filter input component
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
	case models.FieldUpdateResult:
		return handleFieldUpdateResult(m, msg)
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
		// Only resize RowDetailList if it has been initialized
		if len(m.RowDetailList.Items()) > 0 {
			m.RowDetailList.SetSize(msg.Width-h, msg.Height-v-8)
		}
		m.TextInput.Width = msg.Width - h - 4
		m.NameInput.Width = msg.Width - h - 4
		m.QueryInput.Width = msg.Width - h - 4
		m.SearchInput.Width = msg.Width - h - 4

		// Update textarea size for field editing
		textareaWidth := msg.Width - h - 4
		textareaHeight := msg.Height - v - 8 // Reserve space for title and help text only
		if textareaWidth < 40 {
			textareaWidth = 40
		}
		if textareaHeight < 5 {
			textareaHeight = 5
		}
		m.FieldTextarea.SetWidth(textareaWidth)
		m.FieldTextarea.SetHeight(textareaHeight)

		// Recompute data preview table to fill available space
		if len(m.DataPreviewAllColumns) > 0 && len(m.DataPreviewAllRows) > 0 {
			m = createDataPreviewTable(m)
		}

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
				return m, nil
			case models.ConnectionView:
				m.State = models.DBTypeView
				m.Err = nil
				return m, nil
			case models.SaveConnectionView:
				m.State = models.ConnectionView
				m.Err = nil
				return m, nil
			case models.EditConnectionView:
				m.State = models.SavedConnectionsView
				m.Err = nil
				m.EditingConnectionIdx = -1
				return m, nil
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
				return m, nil
			case models.ColumnsView:
				m.State = models.TablesView
				return m, nil
			case models.DataPreviewView:
				m.State = models.TablesView
				return m, nil
			case models.RelationshipsView:
				m.State = models.TablesView
				return m, nil
				// Note: RowDetailView ESC handling is done in the specific handler below
			}

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

			case models.ConnectionView:
				if !m.IsConnecting && !m.IsTestingConnection {
					m.ConnectionStr = m.TextInput.Value()
					if m.ConnectionStr != "" {
						// Save connection if a name is provided
						connectionName := strings.TrimSpace(m.NameInput.Value())
						if connectionName != "" {
							// Check if connection name already exists
							nameExists := false
							for i, conn := range m.SavedConnections {
								if conn.Name == connectionName {
									// Update existing connection
									m.SavedConnections[i] = models.SavedConnection{
										Name:          connectionName,
										Driver:        m.SelectedDB.Driver,
										ConnectionStr: m.ConnectionStr,
									}
									nameExists = true
									break
								}
							}
							// Add new connection if name doesn't exist
							if !nameExists {
								newConnection := models.SavedConnection{
									Name:          connectionName,
									Driver:        m.SelectedDB.Driver,
									ConnectionStr: m.ConnectionStr,
								}
								m.SavedConnections = append(m.SavedConnections, newConnection)
							}
							config.SaveConnections(m.SavedConnections)
						}

						// Connect to database
						m.IsConnecting = true
						m.Err = nil
						m.QueryResult = ""
						return m, connectDB(m)
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
					m.DataPreviewCurrentPage = 0 // Reset to first page
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
					m.DataPreviewCurrentPage = 0 // Reset to first page
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
			// Handle filter mode
			if m.DataPreviewFilterActive {
				switch km.String() {
				case "enter":
					// Apply filter
					m.DataPreviewFilterValue = m.DataPreviewFilterInput.Value()
					m.DataPreviewFilterActive = false
					m.DataPreviewFilterInput.Blur()
					m.DataPreviewCurrentPage = 0 // Reset to first page
					return m, loadDataPreviewWithFilter(m)
				case "esc":
					// Cancel filter
					m.DataPreviewFilterActive = false
					m.DataPreviewFilterInput.Blur()
					m.DataPreviewFilterInput.SetValue("")
					return m, nil
				default:
					// Update filter input
					m.DataPreviewFilterInput, cmd = m.DataPreviewFilterInput.Update(msg)
					return m, cmd
				}
			}

			// Handle sort mode
			if m.DataPreviewSortMode {
				switch km.String() {
				case "up", "k":
					// Move to previous column
					currentIdx := -1
					for i, col := range m.DataPreviewAllColumns {
						if col == m.DataPreviewSortColumn {
							currentIdx = i
							break
						}
					}
					if currentIdx > 0 {
						m.DataPreviewSortColumn = m.DataPreviewAllColumns[currentIdx-1]
					}
					return m, nil
				case "down", "j":
					// Move to next column
					currentIdx := -1
					for i, col := range m.DataPreviewAllColumns {
						if col == m.DataPreviewSortColumn {
							currentIdx = i
							break
						}
					}
					if currentIdx >= 0 && currentIdx < len(m.DataPreviewAllColumns)-1 {
						m.DataPreviewSortColumn = m.DataPreviewAllColumns[currentIdx+1]
					} else if currentIdx == -1 && len(m.DataPreviewAllColumns) > 0 {
						m.DataPreviewSortColumn = m.DataPreviewAllColumns[0]
					}
					return m, nil
				case "enter":
					// Toggle sort direction: off -> asc -> desc -> off
					switch m.DataPreviewSortDirection {
					case models.SortOff:
						m.DataPreviewSortDirection = models.SortAsc
					case models.SortAsc:
						m.DataPreviewSortDirection = models.SortDesc
					case models.SortDesc:
						m.DataPreviewSortDirection = models.SortOff
						m.DataPreviewSortColumn = ""
					}
					m.DataPreviewSortMode = false
					m.DataPreviewCurrentPage = 0 // Reset to first page when sorting changes
					return m, loadDataPreviewWithSort(m)
				case "esc":
					// Exit sort mode
					m.DataPreviewSortMode = false
					return m, nil
				}
				return m, nil
			}

			// Normal navigation mode
			switch km.String() {
			case "enter":
				// Enter row detail view
				if len(m.DataPreviewAllRows) > 0 {
					selectedRow := m.DataPreviewTable.Cursor()
					if selectedRow >= 0 && selectedRow < len(m.DataPreviewAllRows) {
						// Calculate the actual row index based on current page and table position
						actualRowIndex := (m.DataPreviewCurrentPage * m.DataPreviewItemsPerPage) + selectedRow
						if actualRowIndex < len(m.DataPreviewAllRows) {
							m.SelectedRowData = m.DataPreviewAllRows[selectedRow] // Use the displayed row
							m.SelectedRowIndex = actualRowIndex                   // Track the actual position in the dataset

							// Create list items for each field
							var items []list.Item
							for i, col := range m.DataPreviewAllColumns {
								var value string
								if i < len(m.SelectedRowData) {
									value = m.SelectedRowData[i]
								} else {
									value = "NULL"
								}
								items = append(items, models.FieldItem{Name: col, Value: value})
							}

							// Initialize the row detail list (full-width/height)
							// Use custom delegate to show type badges aligned right
							m.RowDetailList = list.New(items, fieldItemDelegate{}, 0, 0)
							// Keep the outer view title; hide internal list title for cleaner look
							m.RowDetailList.Title = ""
							m.RowDetailList.SetShowTitle(false)
							m.RowDetailList.SetShowStatusBar(false)
							m.RowDetailList.SetFilteringEnabled(false)
							// Hide built-in help to avoid duplicate help sections
							m.RowDetailList.SetShowHelp(false)
							// Size the list to available viewport immediately
							h, v := styles.DocStyle.GetFrameSize()
							// Reserve fewer lines so the title is visible and top items are not clipped
							m.RowDetailList.SetSize(m.Width-h, m.Height-v-5)
							m.IsViewingFieldDetail = false

							m.State = models.RowDetailView
							return m, nil
						}
					}
				}
				return m, nil
			case "/":
				// Start filter mode
				m.DataPreviewFilterActive = true
				m.DataPreviewFilterInput.Focus()
				return m, nil
			case "s":
				// Start sort mode
				m.DataPreviewSortMode = true
				if m.DataPreviewSortColumn == "" && len(m.DataPreviewAllColumns) > 0 {
					m.DataPreviewSortColumn = m.DataPreviewAllColumns[0] // Start with first column
				}
				return m, nil
			case "r":
				return m, loadDataPreview(m)
			case "left":
				// Previous page
				if m.DataPreviewCurrentPage > 0 {
					m.DataPreviewCurrentPage--
					return m, loadDataPreviewWithPagination(m)
				}
				return m, nil
			case "right":
				// Next page
				totalPages := (m.DataPreviewTotalRows + m.DataPreviewItemsPerPage - 1) / m.DataPreviewItemsPerPage
				if m.DataPreviewCurrentPage < totalPages-1 {
					m.DataPreviewCurrentPage++
					return m, loadDataPreviewWithPagination(m)
				}
				return m, nil
			case "h":
				// Scroll left (show previous columns)
				if m.DataPreviewScrollOffset > 0 {
					m.DataPreviewScrollOffset--
					m = createDataPreviewTable(m)
				}
				return m, nil
			case "l":
				// Scroll right (show next columns)
				totalCols := len(m.DataPreviewAllColumns)
				if m.DataPreviewScrollOffset+m.DataPreviewVisibleCols < totalCols {
					m.DataPreviewScrollOffset++
					m = createDataPreviewTable(m)
				}
				return m, nil
			case "j", "k", "up", "down":
				// Allow j/k and arrow keys to navigate within the table rows
				m.DataPreviewTable, cmd = m.DataPreviewTable.Update(msg)
				return m, cmd
			}
		}
		m.DataPreviewTable, cmd = m.DataPreviewTable.Update(msg)
	case models.RowDetailView:
		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.String() {
			case "enter":
				if !m.IsViewingFieldDetail {
					// Enter field detail view
					if selectedItem, ok := m.RowDetailList.SelectedItem().(models.FieldItem); ok {
						m.SelectedFieldForDetail = selectedItem.Name
						m.IsViewingFieldDetail = true
						// Reset scroll positions when entering field detail
						m.FieldDetailScrollOffset = 0
						m.FieldDetailHorizontalOffset = 0
					}
				}
				return m, nil
			case "e":
				if !m.IsViewingFieldDetail && !m.IsEditingField {
					// Enter field edit mode
					if selectedItem, ok := m.RowDetailList.SelectedItem().(models.FieldItem); ok {
						m.EditingFieldName = selectedItem.Name
						m.OriginalFieldValue = selectedItem.Value

						// Find the field index
						for i, col := range m.DataPreviewAllColumns {
							if col == selectedItem.Name {
								m.EditingFieldIndex = i
								break
							}
						}

						// Initialize textarea with current value and place cursor at start
						m.FieldTextarea.SetValue(selectedItem.Value)
						m.FieldTextarea.CursorStart()

						// Set responsive textarea size
						h, v := styles.DocStyle.GetFrameSize()
						textareaWidth := m.Width - h - 4
						// Use more space now that we only render the textarea in edit view
						textareaHeight := m.Height - v - 8 // Reserve space for title and help text only
						if textareaWidth < 40 {
							textareaWidth = 40
						}
						if textareaHeight < 5 {
							textareaHeight = 5
						}
						m.FieldTextarea.SetWidth(textareaWidth)
						m.FieldTextarea.SetHeight(textareaHeight)

						m.FieldTextarea.Focus()
						m.IsEditingField = true
					}
				}
				return m, nil
			case "esc":
				if m.IsEditingField {
					// Exit edit mode without saving
					m.IsEditingField = false
					m.FieldTextarea.Blur()
					m.EditingFieldName = ""
					m.OriginalFieldValue = ""
					return m, nil
				} else if m.IsViewingFieldDetail {
					// Exit field detail view, back to field list
					m.IsViewingFieldDetail = false
				} else {
					// Return to data preview
					m.State = models.DataPreviewView
				}
				return m, nil
			case "ctrl+s":
				if m.IsEditingField {
					// Save the edited field
					newValue := m.FieldTextarea.Value()
					return m, saveFieldEdit(m, newValue)
				}
				return m, nil
			case "ctrl+k":
				if m.IsEditingField {
					// Clear all text in the edit textarea
					m.FieldTextarea.SetValue("")
					m.FieldTextarea.CursorStart()
				}
				return m, nil
			case "up", "k":
				if m.IsViewingFieldDetail {
					// Scroll up in field detail view
					if m.FieldDetailScrollOffset > 0 {
						m.FieldDetailScrollOffset--
					}
					return m, nil
				}
			case "down", "j":
				if m.IsViewingFieldDetail {
					// Scroll down in field detail view
					fieldValue := ""
					for i, col := range m.DataPreviewAllColumns {
						if col == m.SelectedFieldForDetail && i < len(m.SelectedRowData) {
							fieldValue = m.SelectedRowData[i]
							break
						}
					}
					// Calculate max scroll based on field content and dynamic height
					lines := len(strings.Split(fieldValue, "\n"))
					availableHeight := m.Height - 10 // Same calculation as in view
					if availableHeight < 5 {
						availableHeight = 5
					}
					maxScroll := lines - availableHeight
					if maxScroll < 0 {
						maxScroll = 0
					}
					if m.FieldDetailScrollOffset < maxScroll {
						m.FieldDetailScrollOffset++
					}
					return m, nil
				}
			case "left", "h":
				if m.IsViewingFieldDetail {
					// Calculate scroll increment based on available width
					availableWidth := m.Width - 10
					if availableWidth < 40 {
						availableWidth = 40
					}
					if availableWidth > 200 {
						availableWidth = 200
					}
					scrollIncrement := availableWidth / 4 // Scroll by 1/4 of screen width
					if scrollIncrement < 5 {
						scrollIncrement = 5 // Minimum scroll
					}

					m.FieldDetailHorizontalOffset -= scrollIncrement
					if m.FieldDetailHorizontalOffset < 0 {
						m.FieldDetailHorizontalOffset = 0
					}
					return m, nil
				}
			case "right", "l":
				if m.IsViewingFieldDetail {
					// Calculate scroll increment based on available width
					availableWidth := m.Width - 10
					if availableWidth < 40 {
						availableWidth = 40
					}
					if availableWidth > 200 {
						availableWidth = 200
					}
					scrollIncrement := availableWidth / 4 // Scroll by 1/4 of screen width
					if scrollIncrement < 5 {
						scrollIncrement = 5 // Minimum scroll
					}

					m.FieldDetailHorizontalOffset += scrollIncrement
					return m, nil
				}
			}
		}

		// Update components based on mode
		if m.IsEditingField {
			// Update textarea when in edit mode
			m.FieldTextarea, cmd = m.FieldTextarea.Update(msg)
			return m, cmd
		} else if !m.IsViewingFieldDetail {
			// Update list when in field list mode
			m.RowDetailList, cmd = m.RowDetailList.Update(msg)
		} else {
			// When viewing field detail, don't let other key handlers interfere
			return m, nil
		}
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
	case models.RowDetailView:
		return views.RowDetailView(m.Model)
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

	// Sort tables alphabetically
	sort.Strings(m.Tables)

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

	// Update tables list (show only table names)
	items := make([]list.Item, len(m.TableInfos))
	for i, info := range m.TableInfos {
		items[i] = models.Item{
			ItemTitle: info.Name,
			// omit description to avoid redundant "📊 Table" line per item
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
		// Reset pagination and load first page
		totalRows, err := database.GetTableRowCount(m.DB, m.SelectedDB.Driver, m.SelectedTable, m.SelectedSchema)
		if err != nil {
			return models.DataPreviewResult{Columns: nil, Rows: nil, Err: err}
		}

		// Determine sort parameters
		var sortColumn, sortDirection string
		if m.DataPreviewSortDirection != models.SortOff && m.DataPreviewSortColumn != "" {
			sortColumn = m.DataPreviewSortColumn
			switch m.DataPreviewSortDirection {
			case models.SortAsc:
				sortDirection = "ASC"
			case models.SortDesc:
				sortDirection = "DESC"
			}
		}

		cols, rows, err := database.GetTablePreviewPaginatedWithSort(m.DB, m.SelectedDB.Driver, m.SelectedTable, m.SelectedSchema, m.DataPreviewItemsPerPage, 0, sortColumn, sortDirection)
		return models.DataPreviewResult{Columns: cols, Rows: rows, Err: err, TotalRows: totalRows}
	})
}

func loadDataPreviewWithPagination(m appModel) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Determine sort parameters
		var sortColumn, sortDirection string
		if m.DataPreviewSortDirection != models.SortOff && m.DataPreviewSortColumn != "" {
			sortColumn = m.DataPreviewSortColumn
			switch m.DataPreviewSortDirection {
			case models.SortAsc:
				sortDirection = "ASC"
			case models.SortDesc:
				sortDirection = "DESC"
			}
		}

		offset := m.DataPreviewCurrentPage * m.DataPreviewItemsPerPage
		if m.DataPreviewFilterValue != "" {
			cols, rows, err := database.GetTablePreviewPaginatedWithFilterAndSort(m.DB, m.SelectedDB.Driver, m.SelectedTable, m.SelectedSchema, m.DataPreviewItemsPerPage, offset, m.DataPreviewFilterValue, m.DataPreviewAllColumns, sortColumn, sortDirection)
			return models.DataPreviewResult{Columns: cols, Rows: rows, Err: err, TotalRows: m.DataPreviewTotalRows}
		}
		cols, rows, err := database.GetTablePreviewPaginatedWithSort(m.DB, m.SelectedDB.Driver, m.SelectedTable, m.SelectedSchema, m.DataPreviewItemsPerPage, offset, sortColumn, sortDirection)
		return models.DataPreviewResult{Columns: cols, Rows: rows, Err: err, TotalRows: m.DataPreviewTotalRows}
	})
}

func loadDataPreviewWithFilter(m appModel) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Get total rows with filter
		totalRows, err := database.GetTableRowCountWithFilter(m.DB, m.SelectedDB.Driver, m.SelectedTable, m.SelectedSchema, m.DataPreviewFilterValue, m.DataPreviewAllColumns)
		if err != nil {
			return models.DataPreviewResult{Columns: nil, Rows: nil, Err: err}
		}

		// Determine sort parameters
		var sortColumn, sortDirection string
		if m.DataPreviewSortDirection != models.SortOff && m.DataPreviewSortColumn != "" {
			sortColumn = m.DataPreviewSortColumn
			switch m.DataPreviewSortDirection {
			case models.SortAsc:
				sortDirection = "ASC"
			case models.SortDesc:
				sortDirection = "DESC"
			}
		}

		// Get filtered and sorted data
		cols, rows, err := database.GetTablePreviewPaginatedWithFilterAndSort(m.DB, m.SelectedDB.Driver, m.SelectedTable, m.SelectedSchema, m.DataPreviewItemsPerPage, 0, m.DataPreviewFilterValue, m.DataPreviewAllColumns, sortColumn, sortDirection)
		return models.DataPreviewResult{Columns: cols, Rows: rows, Err: err, TotalRows: totalRows}
	})
}

func loadDataPreviewWithSort(m appModel) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Determine sort parameters
		var sortColumn, sortDirection string
		if m.DataPreviewSortDirection != models.SortOff && m.DataPreviewSortColumn != "" {
			sortColumn = m.DataPreviewSortColumn
			switch m.DataPreviewSortDirection {
			case models.SortAsc:
				sortDirection = "ASC"
			case models.SortDesc:
				sortDirection = "DESC"
			}
		}

		offset := m.DataPreviewCurrentPage * m.DataPreviewItemsPerPage

		// Use appropriate function based on whether filter is active
		if m.DataPreviewFilterValue != "" {
			cols, rows, err := database.GetTablePreviewPaginatedWithFilterAndSort(m.DB, m.SelectedDB.Driver, m.SelectedTable, m.SelectedSchema, m.DataPreviewItemsPerPage, offset, m.DataPreviewFilterValue, m.DataPreviewAllColumns, sortColumn, sortDirection)
			return models.DataPreviewResult{Columns: cols, Rows: rows, Err: err, TotalRows: m.DataPreviewTotalRows}
		} else {
			cols, rows, err := database.GetTablePreviewPaginatedWithSort(m.DB, m.SelectedDB.Driver, m.SelectedTable, m.SelectedSchema, m.DataPreviewItemsPerPage, offset, sortColumn, sortDirection)
			return models.DataPreviewResult{Columns: cols, Rows: rows, Err: err, TotalRows: m.DataPreviewTotalRows}
		}
	})
}

func handleDataPreviewResult(m appModel, msg models.DataPreviewResult) (appModel, tea.Cmd) {
	m.IsLoadingPreview = false
	if msg.Err != nil {
		m.Err = msg.Err
		return m, nil
	}

	// Update total rows count
	if msg.TotalRows > 0 {
		m.DataPreviewTotalRows = msg.TotalRows
	}

	// Store all columns and rows for horizontal scrolling
	m.DataPreviewAllColumns = msg.Columns
	m.DataPreviewAllRows = msg.Rows
	m.DataPreviewScrollOffset = 0 // Reset scroll position

	// Create the initial table view
	m = createDataPreviewTable(m)
	m.State = models.DataPreviewView
	return m, nil
}

func createDataPreviewTable(m appModel) appModel {
	if len(m.DataPreviewAllColumns) == 0 {
		return m
	}

	// Determine available width for table content within the document frame
	h, v := styles.DocStyle.GetFrameSize()
	availableWidth := m.Width - h - 4
	if availableWidth < 20 {
		availableWidth = 20
	}

	// Precompute width for every column (based on header and data) capped to [8, 40]
	colWidths := make([]int, len(m.DataPreviewAllColumns))
	for colIdx, c := range m.DataPreviewAllColumns {
		maxWidth := len(c)
		for _, row := range m.DataPreviewAllRows {
			if colIdx < len(row) {
				cellLength := len(row[colIdx])
				if cellLength > 40 {
					cellLength = 40
				}
				if cellLength > maxWidth {
					maxWidth = cellLength
				}
			}
		}
		if maxWidth < 8 {
			maxWidth = 8
		} else if maxWidth > 40 {
			maxWidth = 40
		}
		colWidths[colIdx] = maxWidth
	}

	// Compute how many columns fit starting from the current scroll offset
	startCol := m.DataPreviewScrollOffset
	sum := 0
	endCol := startCol
	for endCol < len(colWidths) {
		// Rough allowance for padding/separators per column
		next := colWidths[endCol] + 3
		if sum+next > availableWidth {
			break
		}
		sum += next
		endCol++
	}
	if endCol == startCol {
		// Ensure at least one column is visible
		endCol = min(startCol+1, len(colWidths))
	}
	visibleCount := endCol - startCol
	if visibleCount < 0 {
		visibleCount = 0
	}
	m.DataPreviewVisibleCols = visibleCount

	// Build visible columns
	visibleCols := m.DataPreviewAllColumns[startCol:endCol]
	cols := make([]table.Column, len(visibleCols))
	for i, c := range visibleCols {
		cols[i] = table.Column{Title: c, Width: colWidths[startCol+i]}
	}

	// Build visible rows with content truncation per computed column width
	rows := make([]table.Row, len(m.DataPreviewAllRows))
	for i, r := range m.DataPreviewAllRows {
		visibleCells := make(table.Row, len(visibleCols))
		for j := 0; j < len(visibleCols); j++ {
			colIndex := startCol + j
			if colIndex < len(r) {
				cell := r[colIndex]
				maxW := colWidths[colIndex]
				if len(cell) > maxW {
					ellipsis := 0
					if maxW >= 3 {
						ellipsis = 3
					}
					trim := max(0, maxW-ellipsis)
					if trim <= 0 {
						visibleCells[j] = ""
					} else if ellipsis > 0 {
						visibleCells[j] = cell[:trim] + "..."
					} else {
						visibleCells[j] = cell[:trim]
					}
				} else {
					visibleCells[j] = cell
				}
			} else {
				visibleCells[j] = ""
			}
		}
		rows[i] = visibleCells
	}

	// Compute dynamic height to use remaining vertical space
	reserved := 10 // Title + info + help, approximate
	availableHeight := m.Height - v - reserved
	if availableHeight < 5 {
		availableHeight = 5
	}

	// Recreate table with visible columns/rows and dynamic height
	m.DataPreviewTable = table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(availableHeight),
	)
	m.DataPreviewTable.SetStyles(styles.GetBlueTableStyles())

	return m
}

// Small helpers for integer min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func loadRelationships(m appModel) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		rels, err := database.GetForeignKeyRelationships(m.DB, m.SelectedDB.Driver, m.SelectedSchema)
		return models.RelationshipsResult{Relationships: rels, Err: err}
	})
}

func handleFieldUpdateResult(m appModel, msg models.FieldUpdateResult) (appModel, tea.Cmd) {
	if msg.ExitEdit {
		m.IsEditingField = false
		m.FieldTextarea.Blur()
	}

	if msg.Success {
		// Update the row data with the new value
		if m.EditingFieldIndex >= 0 && m.EditingFieldIndex < len(m.SelectedRowData) {
			m.SelectedRowData[m.EditingFieldIndex] = msg.NewValue
		}

		// Update the row detail list with the new value
		items := make([]list.Item, len(m.DataPreviewAllColumns))
		for i, col := range m.DataPreviewAllColumns {
			var value string
			if i < len(m.SelectedRowData) {
				value = m.SelectedRowData[i]
			} else {
				value = "NULL"
			}
			items[i] = models.FieldItem{Name: col, Value: value}
		}
		m.RowDetailList.SetItems(items)

		m.QueryResult = "✅ Field updated successfully!"
		m.Err = nil

		// Clear the editing state
		m.EditingFieldName = ""
		m.OriginalFieldValue = ""

		return m, clearResultAfterTimeout()
	} else {
		m.Err = msg.Err
		return m, nil
	}
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
	m.RelationshipsTable.SetStyles(styles.GetBlueTableStyles())
	m.State = models.RelationshipsView
	return m, nil
}

// Helper functions

// Helper function to clear results after a timeout
func clearResultAfterTimeout() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return models.ClearResultMsg{}
	})
}

// saveFieldEdit creates and executes an UPDATE statement for the edited field
func saveFieldEdit(m appModel, newValue string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Find the primary key column and value for the WHERE clause
		var primaryKeyColumn, primaryKeyValue string

		// Look for common primary key column names
		for i, col := range m.DataPreviewAllColumns {
			if col == "id" || col == "Id" || col == "ID" {
				if i < len(m.SelectedRowData) {
					primaryKeyColumn = col
					primaryKeyValue = m.SelectedRowData[i]
					break
				}
			}
		}

		// If no 'id' column found, try other common patterns
		if primaryKeyColumn == "" {
			for i, col := range m.DataPreviewAllColumns {
				colLower := strings.ToLower(col)
				if strings.HasSuffix(colLower, "_id") || strings.HasSuffix(colLower, "id") {
					if i < len(m.SelectedRowData) {
						primaryKeyColumn = col
						primaryKeyValue = m.SelectedRowData[i]
						break
					}
				}
			}
		}

		if primaryKeyColumn == "" {
			return models.FieldUpdateResult{
				Success:  false,
				Err:      fmt.Errorf("no primary key column found for safe update"),
				ExitEdit: false,
			}
		}

		// Build UPDATE SQL statement
		var updateSQL string
		switch m.SelectedDB.Driver {
		case "postgres":
			updateSQL = fmt.Sprintf(`UPDATE "%s"."%s" SET "%s" = $1 WHERE "%s" = $2`,
				m.SelectedSchema, m.SelectedTable, m.EditingFieldName, primaryKeyColumn)
		case "mysql":
			updateSQL = fmt.Sprintf("UPDATE `%s`.`%s` SET `%s` = ? WHERE `%s` = ?",
				m.SelectedSchema, m.SelectedTable, m.EditingFieldName, primaryKeyColumn)
		case "sqlite3":
			updateSQL = fmt.Sprintf(`UPDATE "%s" SET "%s" = ? WHERE "%s" = ?`,
				m.SelectedTable, m.EditingFieldName, primaryKeyColumn)
		default:
			return models.FieldUpdateResult{
				Success:  false,
				Err:      fmt.Errorf("unsupported database driver: %s", m.SelectedDB.Driver),
				ExitEdit: false,
			}
		}

		// Execute the UPDATE statement
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := m.DB.ExecContext(ctx, updateSQL, newValue, primaryKeyValue)
		if err != nil {
			return models.FieldUpdateResult{
				Success:  false,
				Err:      fmt.Errorf("failed to update field: %v", err),
				ExitEdit: false,
			}
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return models.FieldUpdateResult{
				Success:  false,
				Err:      fmt.Errorf("failed to get affected rows: %v", err),
				ExitEdit: false,
			}
		}

		if rowsAffected == 0 {
			return models.FieldUpdateResult{
				Success:  false,
				Err:      fmt.Errorf("no rows were updated - record may not exist"),
				ExitEdit: false,
			}
		}

		return models.FieldUpdateResult{
			Success:  true,
			ExitEdit: true,
			NewValue: newValue,
		}
	})
}

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

// fieldItemDelegate renders field name/value with a right-aligned type badge.
type fieldItemDelegate struct{}

func (d fieldItemDelegate) Height() int                               { return 1 }
func (d fieldItemDelegate) Spacing() int                              { return 0 }
func (d fieldItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d fieldItemDelegate) Render(w io.Writer, m list.Model, index int, it list.Item) {
	fi, ok := it.(models.FieldItem)
	if !ok {
		return
	}
	width := m.Width()
	if width <= 0 {
		return
	}

	// Determine type and simple prefix when selected
	t := inferFieldType(fi.Value)
	prefix := "  "
	if index == m.Index() {
		prefix = "> "
	}

	// Compose: prefix + Name: value [Type]
	namePart := fi.Name + ": "
	// Style type badge
	badge := styles.TypeBadgeStyle.Render("[" + t + "]")
	// Keep a space before the badge
	spaceBeforeBadge := 1

	// Sanitize value to single line to prevent layout issues
	single := strings.Join(strings.Fields(fi.Value), " ")

	// Budget for value so type is always shown (use display widths)
	budget := width - lipgloss.Width(prefix) - lipgloss.Width(namePart) - spaceBeforeBadge - lipgloss.Width(badge)
	if budget < 0 {
		budget = 0
	}
	val := ansi.Truncate(single, budget, "...")

	line := prefix + namePart + val + strings.Repeat(" ", spaceBeforeBadge) + badge

	// Style when selected (apply to entire line for clarity)
	if index == m.Index() {
		line = styles.FocusedStyle.Render(line)
	}
	fmt.Fprint(w, line)
}

// inferFieldType returns a simple type label based on content heuristics.
func inferFieldType(v string) string {
	if v == "NULL" {
		return "NULL"
	}
	if v == "" {
		return "Text"
	}
	// Boolean
	if v == "true" || v == "false" || v == "TRUE" || v == "FALSE" {
		return "Bool"
	}
	// Numeric
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return "Int"
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return "Float"
	}
	// JSON
	s := strings.TrimSpace(v)
	if (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) || (strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")) {
		return "JSON"
	}
	// DateTime: try to parse with common layouts instead of loose punctuation checks
	if looksLikeDateTime(strings.TrimSpace(v)) {
		return "DateTime"
	}
	return "Text"
}

// looksLikeDateTime attempts parsing with common datetime layouts
func looksLikeDateTime(s string) bool {
	if s == "" {
		return false
	}
	// Avoid obviously long textual content
	if len(s) > 64 {
		return false
	}
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RubyDate,
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04:05.000 -0700 MST",
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if _, err := time.Parse(layout, s); err == nil {
			return true
		}
	}
	return false
}
