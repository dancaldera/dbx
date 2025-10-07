package state

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancaldera/mirador/internal/config"
	"github.com/dancaldera/mirador/internal/models"
)

// HandleSaveConnectionViewUpdate handles all updates for the SaveConnectionView state.
func HandleSaveConnectionViewUpdate(m models.Model, msg tea.Msg) (models.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// Go back to the connection view
			m.State = models.ConnectionView
			m.Err = nil
			return m, nil

		case "enter":
			// Save the new connection
			name := m.NameInput.Value()
			if name != "" {
				newConnection := models.SavedConnection{
					Name:          name,
					Driver:        m.SelectedDB.Driver,
					ConnectionStr: m.ConnectionStr,
				}
				m.SavedConnections = append(m.SavedConnections, newConnection)
				config.SaveConnections(m.SavedConnections)
				m.State = models.ConnectionView // Go back to connection view after saving
				return m, nil
			}
		}
	}

	// Update the name input field, which is the only active component in this view
	m.NameInput, cmd = m.NameInput.Update(msg)
	return m, cmd
}
