package styles

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// Spacing Standards:
// -------------------
// All views use the ViewBuilder pattern which automatically applies DocStyle (Margin: 1 top/bottom, 2 left/right)
// This ensures consistent spacing across all views without manual margin management
//
// Standard margins/padding:
// - DocStyle: Margin(1, 2) - Base container for all views
// - HelpStyle: Margin(1, 0), Padding(0, 1) - Help text spacing
// - TitleStyle: Padding(0, 1), Margin(0, 0, 1, 0) - Title with bottom margin
// - SubtitleStyle: Margin(0, 0, 1, 0) - Section titles with bottom margin
// - CardStyle: Padding(1, 2), Margin(0, 0, 1, 0) - Content containers
// - InputStyle/InputFocusedStyle: Padding(0, 1), Margin(0, 0, 1, 0) - Form inputs
// - Status styles (Error/Success/Warning/Info/Loading): Padding(0, 1) - Status messages
//
// All view functions should use ViewBuilder to ensure consistent application of these standards.

// Global styles with blue theme
var (
	// Primary blue colors
	PrimaryBlue = lipgloss.Color("#00b8db") // Main blue
	LightBlue   = lipgloss.Color("#53eafd") // Light blue
	DarkBlue    = lipgloss.Color("#008ba3") // Dark cyan-blue accent
	AccentBlue  = lipgloss.Color("#29d3ea") // Cyan accent

	// Supporting colors
	DarkGray      = lipgloss.Color("#374151")
	LightGray     = lipgloss.Color("#9CA3AF")
	White         = lipgloss.Color("#FFFFFF")
	SuccessGreen  = lipgloss.Color("#10B981")
	ErrorRed      = lipgloss.Color("#EF4444")
	WarningOrange = lipgloss.Color("#F59E0B")

	// Main title style used in content views
	TitleStyle = lipgloss.NewStyle().
			Foreground(PrimaryBlue).
			Padding(0, 1).
			Margin(0, 0, 1, 0).
			Bold(true)

	// List header title style (looser spacing)
	ListTitleStyle = lipgloss.NewStyle().
			Foreground(PrimaryBlue).
			Padding(0, 1).
			Margin(0, 0, 1, 0).
			Bold(true)

	// Subtitle for sections
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(DarkBlue).
			Bold(true).
			Margin(0, 0, 1, 0)

	// Invisible/transparent-like border to keep layout spacing without drawing lines
	TransparentBorder = lipgloss.Border{
		Top:         " ",
		Bottom:      " ",
		Left:        " ",
		Right:       " ",
		TopLeft:     " ",
		TopRight:    " ",
		BottomLeft:  " ",
		BottomRight: " ",
	}

	// Focused/selected item style
	FocusedStyle = lipgloss.NewStyle().
			Foreground(AccentBlue).
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
			Padding(0, 1)

	// Key binding help style
	KeyStyle = lipgloss.NewStyle().
			Foreground(AccentBlue).
			Bold(true)

		// Error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorRed).
			Padding(0, 1).
			Bold(true)

		// Success messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessGreen).
			Padding(0, 1).
			Bold(true)

		// Warning messages
	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningOrange).
			Padding(0, 1).
			Bold(true)

		// Information boxes
	InfoStyle = lipgloss.NewStyle().
			Foreground(DarkBlue).
			Padding(0, 1).
			Margin(0)

	// Table header style
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(DarkBlue).
				Bold(true).
				Padding(0, 1).
				Align(lipgloss.Center)

	// Main document container
	DocStyle = lipgloss.NewStyle().
			Margin(1, 2).
			Padding(0)

		// Card-like container for sections
	CardStyle = lipgloss.NewStyle().
			Border(TransparentBorder).
			Padding(1, 2).
			Margin(0, 0, 1, 0)

	// Loading indicator style
	LoadingStyle = lipgloss.NewStyle().
			Foreground(AccentBlue).
			Bold(true).
			Italic(true)

	// Type badge style for row details
	TypeBadgeStyle = lipgloss.NewStyle().
			Foreground(AccentBlue).
			Bold(true)
)

// GetBlueTableStyles returns table styles with blue theme
func GetBlueTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		Foreground(DarkBlue).
		Bold(true).
		Align(lipgloss.Center)
	s.Selected = s.Selected.
		Foreground(AccentBlue).
		Bold(true)
	s.Cell = s.Cell.
		Padding(0, 1)
	return s
}

// GetBlueListDelegate returns a list delegate with blue theme
func GetBlueListDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(PrimaryBlue).
		Bold(true).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(AccentBlue).
		Padding(0, 0, 0, 1)
	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(DarkBlue).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(AccentBlue).
		Padding(0, 0, 0, 1)
	d.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(LightGray)
	d.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(LightGray)
	return d
}
