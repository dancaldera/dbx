package styles

import (
    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/lipgloss"
)

// Global styles with magenta theme
var (
	// Primary magenta colors
	PrimaryMagenta = lipgloss.Color("#D946EF") // Main magenta
	LightMagenta   = lipgloss.Color("#F3E8FF") // Light magenta background
	DarkMagenta    = lipgloss.Color("#7C2D91") // Dark magenta
	AccentMagenta  = lipgloss.Color("#A855F7") // Purple accent

	// Supporting colors
	DarkGray      = lipgloss.Color("#374151")
	LightGray     = lipgloss.Color("#9CA3AF")
	White         = lipgloss.Color("#FFFFFF")
	SuccessGreen  = lipgloss.Color("#10B981")
	ErrorRed      = lipgloss.Color("#EF4444")
	WarningOrange = lipgloss.Color("#F59E0B")

    // Main title style used in content views
    TitleStyle = lipgloss.NewStyle().
            Foreground(PrimaryMagenta).
            Padding(0, 1).
            Margin(0, 0, 1, 0).
            Bold(true)

    // List header title style (no extra spacing)
    ListTitleStyle = lipgloss.NewStyle().
            Foreground(PrimaryMagenta).
            Padding(0).
            Margin(0).
            Bold(true)

	// Subtitle for sections
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(DarkMagenta).
			Bold(true).
			Margin(0, 0, 1, 0)

    // Invisible/transparent-like border to keep layout spacing without drawing lines
    TransparentBorder = lipgloss.Border{
            Top:        " ",
            Bottom:     " ",
            Left:       " ",
            Right:      " ",
            TopLeft:    " ",
            TopRight:   " ",
            BottomLeft: " ",
            BottomRight:" ",
    }

    // Focused/selected item style
    FocusedStyle = lipgloss.NewStyle().
            Foreground(AccentMagenta).
            Padding(0, 1).
            Bold(true).
            Border(TransparentBorder)

	// Input field styling
    InputStyle = lipgloss.NewStyle().
            Border(TransparentBorder).
            Padding(0, 1).
            Margin(0, 0, 1, 0)

	// Input field when focused
    InputFocusedStyle = lipgloss.NewStyle().
                Border(TransparentBorder).
                Padding(0, 1).
                Margin(0, 0, 1, 0)

	// Help text style
    HelpStyle = lipgloss.NewStyle().
            Foreground(LightGray).
            Italic(true).
            Margin(1, 0).
            Border(TransparentBorder).
            Padding(0, 1)

	// Key binding help style
	KeyStyle = lipgloss.NewStyle().
			Foreground(AccentMagenta).
			Bold(true)

	// Error messages
    ErrorStyle = lipgloss.NewStyle().
            Foreground(ErrorRed).
            Padding(0, 1).
            Bold(true).
            Border(TransparentBorder)

	// Success messages
    SuccessStyle = lipgloss.NewStyle().
            Foreground(SuccessGreen).
            Padding(0, 1).
            Bold(true).
            Border(TransparentBorder)

	// Warning messages
    WarningStyle = lipgloss.NewStyle().
            Foreground(WarningOrange).
            Padding(0, 1).
            Bold(true).
            Border(TransparentBorder)

	// Information boxes
    InfoStyle = lipgloss.NewStyle().
            Foreground(DarkMagenta).
            Padding(1, 2).
            Border(TransparentBorder).
            Margin(0, 0, 1, 0)

	// Table header style
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(DarkMagenta).
				Bold(true).
				Padding(0, 1).
				Align(lipgloss.Center)

	// Main document container
	DocStyle = lipgloss.NewStyle().
			Margin(2, 2).
			Padding(1)

	// Card-like container for sections
    CardStyle = lipgloss.NewStyle().
            Border(TransparentBorder).
            Padding(1, 2).
            Margin(0, 0, 1, 0)

	// Loading indicator style
	LoadingStyle = lipgloss.NewStyle().
			Foreground(AccentMagenta).
			Bold(true).
			Italic(true)
)

// GetMagentaTableStyles returns table styles with magenta theme
func GetMagentaTableStyles() table.Styles {
    s := table.DefaultStyles()
    s.Header = s.Header.
        Foreground(DarkMagenta).
        Bold(true).
        Align(lipgloss.Center)
    s.Selected = s.Selected.
        Foreground(AccentMagenta).
        Bold(true)
    s.Cell = s.Cell.
        Padding(0, 1)
    return s
}
