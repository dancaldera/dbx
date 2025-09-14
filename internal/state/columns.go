package state

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancaldera/dbx/internal/models"
)

// HandleColumnsViewUpdate handles all updates for the ColumnsView state.
func HandleColumnsViewUpdate(m models.Model, msg tea.Msg) (models.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// Go back to the tables view
			m.State = models.TablesView
			m.Err = nil
			return m, nil

		case "s":
			// Allow saving the current connection from this view
			if m.ConnectionStr != "" {
				m.State = models.SaveConnectionView
				m.NameInput.SetValue("")
				m.NameInput.Focus()
				return m, nil
			}
		}
	}

	// If the message was not a key press handled above, it's likely a navigation
	// key (up/down/etc.) that should be handled by the table component.
	m.ColumnsTable, cmd = m.ColumnsTable.Update(msg)
	return m, cmd
}
