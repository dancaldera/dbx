package state

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancaldera/mirador/internal/models"
	"github.com/dancaldera/mirador/internal/utils"
)

// HandleTablesViewUpdate handles all updates for the TablesView state.
func HandleTablesViewUpdate(m models.Model, msg tea.Msg) (models.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// Disconnect from DB, reset state, and go back to the DB type view
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

		case "enter", "p":
			// Load data preview for the selected table
			if i, ok := m.TablesList.SelectedItem().(models.Item); ok && !m.IsLoadingPreview {
				m.SelectedTable = i.ItemTitle
				m.IsLoadingPreview = true
				m.DataPreviewCurrentPage = 0 // Reset to first page
				m.Err = nil
				return m, utils.LoadDataPreview(m.DB, m.SelectedDB, m.SelectedTable, m.SelectedSchema, m.DataPreviewItemsPerPage, m.DataPreviewSortDirection, m.DataPreviewSortColumn)
			}

		case "v":
			// View columns for the selected table
			if i, ok := m.TablesList.SelectedItem().(models.Item); ok && !m.IsLoadingColumns {
				m.SelectedTable = i.ItemTitle
				m.IsLoadingColumns = true
				m.Err = nil
				return m, utils.LoadColumns(m.DB, m.SelectedDB, m.SelectedTable, m.SelectedSchema)
			}

		case "f":
			// View foreign key relationships for the current schema
			if m.DB != nil {
				return m, utils.LoadRelationships(m.DB, m.SelectedDB, m.SelectedSchema)
			}
		}
	}

	// If the message was not a key press handled above, it's likely a navigation
	// key (up/down) that should be handled by the list component.
	m.TablesList, cmd = m.TablesList.Update(msg)
	return m, cmd
}
