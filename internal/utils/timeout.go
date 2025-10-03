package utils

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancaldera/dbx/internal/database"
	"github.com/dancaldera/dbx/internal/models"
)

// ClearResultAfterTimeout returns a command to clear result messages after 3 seconds
func ClearResultAfterTimeout() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return models.ClearResultMsg{}
	})
}

// TestConnection performs a database connection test with timeout
func TestConnection(driver, connectionStr string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		return database.TestConnectionWithTimeout(driver, connectionStr)
	})
}
