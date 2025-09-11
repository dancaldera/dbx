package utils

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancaldera/dbx/internal/database"
	"github.com/dancaldera/dbx/internal/models"
)

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// CalculateTotalPages computes total pages for pagination
func CalculateTotalPages(totalRows, itemsPerPage int) int {
	if itemsPerPage <= 0 {
		return 0
	}
	return (totalRows + itemsPerPage - 1) / itemsPerPage
}

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