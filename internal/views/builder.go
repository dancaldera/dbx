package views

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dancaldera/dbx/internal/styles"
)

// ViewBuilder provides a consistent way to build views with standard spacing and layout
type ViewBuilder struct {
	title      string
	status     string
	statusType StatusType
	content    []string
	helpText   string
}

// StatusType represents different types of status messages
type StatusType int

const (
	StatusNone StatusType = iota
	StatusLoading
	StatusError
	StatusSuccess
	StatusWarning
	StatusInfo
)

// NewViewBuilder creates a new ViewBuilder instance
func NewViewBuilder() *ViewBuilder {
	return &ViewBuilder{
		content:    make([]string, 0),
		statusType: StatusNone,
	}
}

// WithTitle sets the main title for the view
func (vb *ViewBuilder) WithTitle(title string) *ViewBuilder {
	vb.title = title
	return vb
}

// WithStatus sets a status message with the specified type
func (vb *ViewBuilder) WithStatus(message string, statusType StatusType) *ViewBuilder {
	vb.status = message
	vb.statusType = statusType
	return vb
}

// WithContent adds content sections to the view
func (vb *ViewBuilder) WithContent(elements ...string) *ViewBuilder {
	vb.content = append(vb.content, elements...)
	return vb
}

// WithHelp sets the help text for the view
func (vb *ViewBuilder) WithHelp(helpText string) *ViewBuilder {
	vb.helpText = helpText
	return vb
}

// Render builds and returns the final view string
func (vb *ViewBuilder) Render() string {
	var elements []string

	// Add title with inline status if present
	if vb.title != "" {
		if vb.status != "" {
			titleLine := vb.renderTitleWithStatus()
			elements = append(elements, titleLine)
		} else {
			elements = append(elements, styles.TitleStyle.Render(vb.title))
		}
		// Add spacing after title
		elements = append(elements, "")
	}

	// Add standalone status message if no title
	if vb.title == "" && vb.status != "" {
		elements = append(elements, vb.renderStatus())
		elements = append(elements, "")
	}

	// Add content sections
	elements = append(elements, vb.content...)

	// Add help text
	if vb.helpText != "" {
		elements = append(elements, vb.helpText)
	}

	// Join all elements and render with DocStyle
	content := lipgloss.JoinVertical(lipgloss.Left, elements...)
	return styles.DocStyle.Render(content)
}

// renderTitleWithStatus creates a title with inline status message
func (vb *ViewBuilder) renderTitleWithStatus() string {
	titlePart := styles.TitleStyle.Render(vb.title)
	statusPart := vb.renderStatus()
	return lipgloss.JoinHorizontal(lipgloss.Left, titlePart, "  ", statusPart)
}

// renderStatus renders the status message with appropriate styling
func (vb *ViewBuilder) renderStatus() string {
	switch vb.statusType {
	case StatusLoading:
		return styles.LoadingStyle.Render(vb.status)
	case StatusError:
		return styles.ErrorStyle.Render(vb.status)
	case StatusSuccess:
		return styles.SuccessStyle.Render(vb.status)
	case StatusWarning:
		return styles.WarningStyle.Render(vb.status)
	case StatusInfo:
		return styles.InfoStyle.Render(vb.status)
	default:
		return vb.status
	}
}

// RenderInputField renders a labeled input field with consistent styling
func RenderInputField(label, fieldView string, focused bool) string {
	labelRendered := styles.SubtitleStyle.Render(label)

	var fieldRendered string
	if focused {
		fieldRendered = styles.InputFocusedStyle.Render(fieldView)
	} else {
		fieldRendered = styles.InputStyle.Render(fieldView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, labelRendered, fieldRendered)
}

// RenderSectionTitle renders a section subtitle
func RenderSectionTitle(title string) string {
	return styles.SubtitleStyle.Render(title)
}

// RenderInfoBox renders an informational message box
func RenderInfoBox(content string) string {
	return styles.InfoStyle.Render(content)
}

// RenderEmptyState renders a standard empty state message
func RenderEmptyState(icon, message string) string {
	return styles.InfoStyle.Render(icon + " " + message)
}
