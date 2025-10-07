package state

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancaldera/mirador/internal/config"
	"github.com/dancaldera/mirador/internal/models"
	"github.com/dancaldera/mirador/internal/utils"
)

// HandleConnectionViewUpdate handles all updates for the ConnectionView state.
func HandleConnectionViewUpdate(m models.Model, msg tea.Msg) (models.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle key messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			// Go back to the DB type selection view
			m.State = models.DBTypeView
			m.Err = nil
			return m, nil

		case "f1":
			// Test the connection
			if !m.IsTestingConnection {
				m.ConnectionStr = m.TextInput.Value()
				if m.ConnectionStr != "" {
					m.IsTestingConnection = true
					m.Err = nil
					m.QueryResult = ""
					return m, utils.TestConnection(m.SelectedDB.Driver, m.ConnectionStr)
				}
			}
			return m, nil // Do nothing if already testing

		case "enter":
			// Connect to the database
			if !m.IsConnecting && !m.IsTestingConnection {
				m.ConnectionStr = m.TextInput.Value()
				if m.ConnectionStr != "" {
					// Save connection if a name is provided
					connectionName := strings.TrimSpace(m.NameInput.Value())
					if connectionName != "" {
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
					return m, utils.ConnectToDB(m.SelectedDB, m.ConnectionStr)
				}
			}
			return m, nil // Do nothing if already connecting/testing

		case "tab":
			// Switch focus between name and connection string inputs
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

	// Update the focused text input
	if m.NameInput.Focused() {
		m.NameInput, cmd = m.NameInput.Update(msg)
	} else {
		m.TextInput, cmd = m.TextInput.Update(msg)
	}

	return m, cmd
}
