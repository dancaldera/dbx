package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/dancaldera/dbx/internal/config"
	"github.com/dancaldera/dbx/internal/models"
	"github.com/dancaldera/dbx/internal/state"
	"github.com/dancaldera/dbx/internal/styles"
	"github.com/dancaldera/dbx/internal/utils"
	"github.com/dancaldera/dbx/internal/views"
)

const version = "v0.1.0"

func initialModel() models.Model {
	// Database types list
	items := make([]list.Item, len(models.SupportedDatabaseTypes))
	for i, db := range models.SupportedDatabaseTypes {
		items[i] = models.Item{
			ItemTitle: db.Name,
			ItemDesc:  fmt.Sprintf("Connect to %s database", db.Name),
		}
	}

	dbList := list.New(items, styles.GetBlueListDelegate(), 0, 0)
	dbList.Title = fmt.Sprintf("DBX — Database Explorer %s", version)
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

	// Query history list
	queryHistoryList := list.New([]list.Item{}, styles.GetBlueListDelegate(), 0, 0)
	queryHistoryList.Title = "Query History"
	qhLS := list.DefaultStyles()
	qhLS.Title = styles.ListTitleStyle
	qhLS.TitleBar = lipgloss.NewStyle()
	queryHistoryList.Styles = qhLS
	queryHistoryList.SetShowStatusBar(false)
	queryHistoryList.SetFilteringEnabled(false)
	queryHistoryList.SetShowHelp(false)

	// Populate query history list items
	if len(queryHistory) > 0 {
		historyItems := make([]list.Item, len(queryHistory))
		for i, entry := range queryHistory {
			// Format timestamp
			timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
			// Create description with success status and row count
			desc := fmt.Sprintf("%s • %s", timestamp, entry.Database)
			if entry.Success && entry.RowCount > 0 {
				desc += fmt.Sprintf(" • %d rows", entry.RowCount)
			} else if !entry.Success {
				desc += " • Failed"
			}

			historyItems[i] = models.Item{
				ItemTitle: entry.Query,
				ItemDesc:  desc,
			}
		}
		queryHistoryList.SetItems(historyItems)
	}

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
		QueryHistoryList:        queryHistoryList,
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
		updatedModel, cmd := utils.HandleConnectResult(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.TestConnectionResult:
		updatedModel, cmd := utils.HandleTestConnectionResult(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.ColumnsResult:
		updatedModel, cmd := utils.HandleColumnsResult(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.DataPreviewResult:
		updatedModel, cmd := utils.HandleDataPreviewResult(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.RelationshipsResult:
		updatedModel, cmd := utils.HandleRelationshipsResult(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.FieldUpdateResult:
		updatedModel, cmd := utils.HandleFieldUpdateResult(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.QueryResultMsg:
		m.IsExecutingQuery = false

		if msg.Err != nil {
			m.Err = msg.Err
			m.QueryResult = ""
		} else {
			m.Err = nil
			m.QueryResult = msg.Result

			// Update query results table if we have columns and rows
			if len(msg.Columns) > 0 && len(msg.Rows) > 0 {
				// Create table columns
				columns := make([]table.Column, len(msg.Columns))
				for i, col := range msg.Columns {
					columns[i] = table.Column{Title: col, Width: 20}
				}

				// Create table rows
				rows := make([]table.Row, len(msg.Rows))
				for i, row := range msg.Rows {
					tableRow := make(table.Row, len(row))
					copy(tableRow, row)
					rows[i] = tableRow
				}

				// Update the table
				m.QueryResultsTable = table.New(
					table.WithColumns(columns),
					table.WithRows(rows),
					table.WithFocused(true),
					table.WithHeight(10),
				)
				m.QueryResultsTable.SetStyles(styles.GetBlueTableStyles())
			}
		}

		return m, nil
	case models.ClearResultMsg:
		m.QueryResult = ""
		return m, nil
	case models.ClearErrorMsg:
		m.Err = nil
		m.ErrorTimeout = nil
		return m, nil
	case models.ErrorTimeoutMsg:
		updatedModel := utils.ClearErrorTimeout(m.Model)
		m.Model = updatedModel
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
		textareaWidth := utils.Max(msg.Width-h-4, 40)
		textareaHeight := utils.Max(msg.Height-v-8, 5) // Reserve space for title and help text only
		m.FieldTextarea.SetWidth(textareaWidth)
		m.FieldTextarea.SetHeight(textareaHeight)

		// Recompute data preview table to fill available space
		if len(m.DataPreviewAllColumns) > 0 && len(m.DataPreviewAllRows) > 0 {
			m.Model = utils.CreateDataPreviewTable(m.Model)
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
				updatedModel, cmd := state.HandleDBTypeViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			}

		case "esc":
			switch m.State {
			case models.SavedConnectionsView:
				updatedModel, cmd := state.HandleSavedConnectionsViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			case models.ConnectionView:
				updatedModel, cmd := state.HandleConnectionViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			case models.SaveConnectionView:
				updatedModel, cmd := state.HandleSaveConnectionViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			case models.EditConnectionView:
				m.State = models.SavedConnectionsView
				m.Err = nil
				m.EditingConnectionIdx = -1
				return m, nil
			case models.TablesView:
				updatedModel, cmd := state.HandleTablesViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			case models.ColumnsView:
				updatedModel, cmd := state.HandleColumnsViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			case models.DataPreviewView:
				updatedModel, cmd := state.HandleDataPreviewViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			case models.RelationshipsView:
				m.State = models.TablesView
				return m, nil
				// Note: RowDetailView ESC handling is done in the specific handler below
			}

		case "s":
			if m.State == models.DBTypeView {
				updatedModel, cmd := state.HandleDBTypeViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			}
			if m.State == models.ColumnsView {
				updatedModel, cmd := state.HandleColumnsViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			}

		case "f1":
			if m.State == models.ConnectionView {
				updatedModel, cmd := state.HandleConnectionViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			}

		case "enter":
			switch m.State {
			case models.DBTypeView:
				updatedModel, cmd := state.HandleDBTypeViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd

			case models.ConnectionView:
				updatedModel, cmd := state.HandleConnectionViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd

			case models.SavedConnectionsView:
				updatedModel, cmd := state.HandleSavedConnectionsViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd

			case models.SaveConnectionView:
				updatedModel, cmd := state.HandleSaveConnectionViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd

			case models.TablesView:
				updatedModel, cmd := state.HandleTablesViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			}
		case "p":
			if m.State == models.TablesView {
				updatedModel, cmd := state.HandleTablesViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			}
		case "v":
			if m.State == models.TablesView {
				updatedModel, cmd := state.HandleTablesViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			}
		case "f":
			if m.State == models.TablesView {
				updatedModel, cmd := state.HandleTablesViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			}
		case "d":
			if m.State == models.SavedConnectionsView && !m.IsConnecting {
				updatedModel, cmd := state.HandleSavedConnectionsViewUpdate(m.Model, msg)
				m.Model = updatedModel
				return m, cmd
			}
		case "r":
			// Navigate to QueryView from TablesView only
			if m.State == models.TablesView {
				m.State = models.QueryView
				return m, nil
			}
		case "ctrl+h":
			// Navigate to QueryHistoryView from TablesView and QueryView only
			if m.State == models.TablesView || m.State == models.QueryView {
				m.State = models.QueryHistoryView
				return m, nil
			}
		}
	}

	// Update components according to state
	switch m.State {
	case models.DBTypeView:
		updatedModel, cmd := state.HandleDBTypeViewUpdate(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.SavedConnectionsView:
		updatedModel, cmd := state.HandleSavedConnectionsViewUpdate(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.ConnectionView:
		updatedModel, cmd := state.HandleConnectionViewUpdate(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.SaveConnectionView:
		updatedModel, cmd := state.HandleSaveConnectionViewUpdate(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.TablesView:
		updatedModel, cmd := state.HandleTablesViewUpdate(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.ColumnsView:
		updatedModel, cmd := state.HandleColumnsViewUpdate(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.DataPreviewView:
		// Handle 'enter' key separately to avoid dependency cycle with private fieldItemDelegate
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
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
						items := utils.UpdateRowDetailList(m.DataPreviewAllColumns, m.SelectedRowData)

						// Initialize the row detail list (full-width/height)
						// Use custom delegate to show type badges aligned right
						m.RowDetailList = list.New(items, state.FieldItemDelegate{}, 0, 0)
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
		}

		// Delegate all other messages to the state handler
		updatedModel, cmd := state.HandleDataPreviewViewUpdate(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.RowDetailView:
		updatedModel, cmd := state.HandleRowDetailViewUpdate(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.QueryView:
		updatedModel, cmd := state.HandleQueryViewUpdate(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
	case models.QueryHistoryView:
		updatedModel, cmd := state.HandleQueryHistoryViewUpdate(m.Model, msg)
		m.Model = updatedModel
		return m, cmd
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
	case models.QueryView:
		return views.QueryView(m.Model)
	case models.QueryHistoryView:
		return views.QueryHistoryView(m.Model)
	default:
		return "View not implemented yet"
	}
}

func main() {
	m := appModel{Model: initialModel()}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
