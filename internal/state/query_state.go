package state

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancaldera/mirador/internal/models"
	"github.com/dancaldera/mirador/internal/utils"
)

// HandleQueryViewUpdate handles all updates for the QueryView state.
func HandleQueryViewUpdate(m models.Model, msg tea.Msg) (models.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// Go back to the data preview view
			m.State = models.DataPreviewView
			m.Err = nil
			m.QueryResult = ""
			return m, nil

		case "enter":
			// Execute the SQL query
			if !m.IsExecutingQuery {
				query := strings.TrimSpace(m.QueryInput.Value())
				if query != "" {
					m.IsExecutingQuery = true
					m.Err = nil
					m.QueryResult = ""
					return m, utils.ExecuteQuery(m.DB, m.SelectedDB, query)
				}
			}
			return m, nil // Do nothing if already executing

		case "tab":
			// Switch focus between query input and results
			if m.QueryInput.Focused() {
				m.QueryInput.Blur()
			} else {
				m.QueryInput.Focus()
			}
			return m, nil
		}
	}

	// Update the query input if it's focused
	if m.QueryInput.Focused() {
		m.QueryInput, cmd = m.QueryInput.Update(msg)
	}

	return m, cmd
}

// HandleQueryHistoryViewUpdate handles all updates for the QueryHistoryView state.
func HandleQueryHistoryViewUpdate(m models.Model, msg tea.Msg) (models.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// Go back to the data preview view
			m.State = models.DataPreviewView
			m.Err = nil
			return m, nil

		case "enter":
			// Select and use the query from history
			if i, ok := m.QueryHistoryList.SelectedItem().(models.Item); ok {
				// Find the corresponding history entry
				for _, entry := range m.QueryHistory {
					if entry.Query == i.ItemTitle {
						// Set the query in the input and switch to query view
						m.QueryInput.SetValue(entry.Query)
						m.State = models.QueryView
						return m, nil
					}
				}
			}
		}
	}

	// Update the query history list
	m.QueryHistoryList, cmd = m.QueryHistoryList.Update(msg)
	return m, cmd
}
