package handlers

import (
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dancaldera/dbx/internal/models"
	"github.com/dancaldera/dbx/internal/styles"
)

// Init initializes the Bubble Tea program
func (m models.Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, textarea.Blink, m.Spinner.Tick)
}

// Update handles all Bubble Tea messages and updates the model
func (m models.Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle connection and column result messages first
	switch msg := msg.(type) {
	case models.ConnectResult:
		return handleConnectResult(m, msg)
	case models.TestConnectionResult:
		return handleTestConnectionResult(m, msg)
	case models.ColumnsResult:
		return handleColumnsResult(m, msg)
	case models.QueryResult:
		return handleQueryResult(m, msg)
	case models.DataPreviewResult:
		return handleDataPreviewResult(m, msg)
	case models.IndexesResult:
		return handleIndexesResult(m, msg)
	case models.RelationshipsResult:
		return handleRelationshipsResult(m, msg)
	case models.ExportResult:
		return handleExportResult(m, msg)
	case models.TestAndSaveResult:
		return handleTestAndSaveResult(m, msg)
	case models.FieldValueResult:
		return handleFieldValueResult(m, msg)
	case models.FieldUpdateResult:
		return handleFieldUpdateResult(m, msg)
	case models.ClipboardResult:
		return handleClipboardResult(m, msg)
	case models.ClearResultMsg:
		m.QueryResult = ""
		return m, nil
	case models.ClearErrorMsg:
		m.Err = nil
		return m, nil
	case tea.WindowSizeMsg:
		return handleWindowResize(m, msg)
	case tea.KeyMsg:
		return handleKeyPress(m, msg)
	}

	// Update components according to state
	return updateComponents(m, msg)
}

// handleWindowResize handles window resize events
func handleWindowResize(m models.Model, msg tea.WindowSizeMsg) (models.Model, tea.Cmd) {
	m.Width = msg.Width
	m.Height = msg.Height
	h, v := styles.DocStyle.GetFrameSize()

	// Update list sizes
	m.DBTypeList.SetSize(msg.Width-h, msg.Height-v-5)
	m.SavedConnectionsList.SetSize(msg.Width-h, msg.Height-v-5)
	m.TablesList.SetSize(msg.Width-h, msg.Height-v-5)
	m.QueryHistoryList.SetSize(msg.Width-h, msg.Height-v-5)
	m.SchemasList.SetSize(msg.Width-h, msg.Height-v-5)

	// Update input widths
	m.TextInput.Width = msg.Width - h - 4
	m.NameInput.Width = msg.Width - h - 4
	m.QueryInput.Width = msg.Width - h - 4
	m.SearchInput.Width = msg.Width - h - 4

	// Update query results table height
	tableHeight := msg.Height - v - 15 // Reserve space for query input and help text
	if tableHeight < 5 {
		tableHeight = 5
	}

	// Simply adjust the height for the existing table
	if len(m.QueryResultsTable.Columns()) > 0 {
		cols := m.QueryResultsTable.Columns()
		rows := m.QueryResultsTable.Rows()
		m.QueryResultsTable = table.New(
			table.WithColumns(cols),
			table.WithRows(rows),
			table.WithFocused(false),
			table.WithHeight(tableHeight),
		)
		// Apply blue theme styles
		m.QueryResultsTable.SetStyles(styles.GetBlueTableStyles())
	}

	return m, nil
}

// Helper function to clear results after a timeout
func clearResultAfterTimeout() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return models.ClearResultMsg{}
	})
}
