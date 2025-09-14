package state

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancaldera/dbx/internal/config"
	"github.com/dancaldera/dbx/internal/models"
	"github.com/dancaldera/dbx/internal/utils"
)

// HandleDBTypeViewUpdate handles all updates for the DBTypeView state.
func HandleDBTypeViewUpdate(m models.Model, msg tea.Msg) (models.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Only handle key messages, other messages (like results) are handled in the main update function
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q", "ctrl+c":
			// In DBTypeView, 'q' is a valid way to quit.
			if m.DB != nil {
				m.DB.Close()
			}
			return m, tea.Quit

		case "s":
			// Switch to saved connections view
			m.State = models.SavedConnectionsView
			connections, err := config.LoadSavedConnections()
			if err == nil {
				m.SavedConnections = connections
			}
			m = utils.UpdateSavedConnectionsList(m)
			return m, nil

		case "enter":
			// Select a database type and move to the connection view
			if i, ok := m.DBTypeList.SelectedItem().(models.Item); ok {
				for _, db := range models.SupportedDatabaseTypes {
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

				// Set placeholder text for the connection string input
				switch m.SelectedDB.Driver {
				case "postgres":
					m.TextInput.Placeholder = "postgres://user:password@localhost/dbname?sslmode=disable"
				case "mysql":
					m.TextInput.Placeholder = "user:password@tcp(localhost:3306)/dbname"
				case "sqlite3":
					m.TextInput.Placeholder = "/path/to/database.db"
				}
			}
			return m, nil
		}
	}

	// If the message was not a key press handled above, it's likely a navigation
	// key (up/down) that should be handled by the list component.
	m.DBTypeList, cmd = m.DBTypeList.Update(msg)
	return m, cmd
}
