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
