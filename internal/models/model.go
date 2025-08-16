package models

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the Bubble Tea program
func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, textarea.Blink, m.Spinner.Tick)
}

// Update will be implemented in the handlers package
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// This will be handled by the handlers package
	return m, nil
}

// View will be implemented in the views package
func (m Model) View() string {
	// This will be handled by the views package
	return ""
}
