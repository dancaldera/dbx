package state

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancaldera/dbx/internal/config"
	"github.com/dancaldera/dbx/internal/models"
	"github.com/dancaldera/dbx/internal/utils"
)

// HandleSavedConnectionsViewUpdate handles all updates for the SavedConnectionsView state.
func HandleSavedConnectionsViewUpdate(m models.Model, msg tea.Msg) (models.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Only handle key messages, other messages are handled in the main update function
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// Go back to the DB type selection view
			m.State = models.DBTypeView
			m.Err = nil
			return m, nil

		case "enter":
			// Connect to the selected saved connection
			if i, ok := m.SavedConnectionsList.SelectedItem().(models.Item); ok && !m.IsConnecting {
				for _, conn := range m.SavedConnections {
					if conn.Name == i.ItemTitle {
						// Find the corresponding DB driver info
						for _, db := range models.SupportedDatabaseTypes {
							if db.Driver == conn.Driver {
								m.SelectedDB = db
								break
							}
						}
						m.ConnectionStr = conn.ConnectionStr
						m.IsConnecting = true
						m.Err = nil
						m.QueryResult = "" // Clear any previous messages
						return m, utils.ConnectToDB(m.SelectedDB, m.ConnectionStr)
					}
				}
			}
			// If we are here, something went wrong or we are already connecting,
			// so we don't return a command.

		case "d":
			// Delete the currently selected saved connection
			if selectedItem, ok := m.SavedConnectionsList.SelectedItem().(models.Item); ok {
				connectionName := selectedItem.ItemTitle
				// Find and remove the connection
				for i, conn := range m.SavedConnections {
					if conn.Name == connectionName {
						// Remove connection from slice
						m.SavedConnections = append(m.SavedConnections[:i], m.SavedConnections[i+1:]...)
						// Save updated connections
						config.SaveConnections(m.SavedConnections)
						// Update the list component
						m = utils.UpdateSavedConnectionsList(m)
						// Show success message
						m.QueryResult = fmt.Sprintf("✅ Deleted connection '%s'", connectionName)
						// Return a command to clear the message after a timeout
						return m, utils.ClearResultAfterTimeout()
					}
				}
			}
		}
	}

	// If the message was not a key press handled above, it's likely a navigation
	// key (up/down) that should be handled by the list component.
	m.SavedConnectionsList, cmd = m.SavedConnectionsList.Update(msg)
	return m, cmd
}
