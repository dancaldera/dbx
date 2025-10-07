package state

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancaldera/mirador/internal/models"
	"github.com/dancaldera/mirador/internal/styles"
	"github.com/dancaldera/mirador/internal/utils"
)

// HandleRowDetailViewUpdate handles all updates for the RowDetailView state.
func HandleRowDetailViewUpdate(m models.Model, msg tea.Msg) (models.Model, tea.Cmd) {
	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// If editing, the textarea consumes all key presses except a few special ones.
		if m.IsEditingField {
			switch keyMsg.String() {
			case "esc":
				// Exit edit mode without saving
				m.IsEditingField = false
				m.FieldTextarea.Blur()
				m.EditingFieldName = ""
				m.OriginalFieldValue = ""
				m.Err = nil
				return m, nil
			case "ctrl+s":
				// Save the edited field
				newValue := m.FieldTextarea.Value()
				return m, utils.SaveFieldEdit(m.DB, m.SelectedDB, m.SelectedSchema, m.SelectedTable, m.EditingFieldName, m.DataPreviewAllColumns, m.SelectedRowData, m.EditingFieldIndex, newValue)
			case "ctrl+k":
				// Clear all text in the edit textarea
				m.FieldTextarea.SetValue("")
				m.FieldTextarea.CursorStart()
				return m, nil
			default:
				// Pass all other key presses to the textarea
				m.FieldTextarea, cmd = m.FieldTextarea.Update(msg)
				return m, cmd
			}
		}

		// If viewing field detail, handle scrolling.
		if m.IsViewingFieldDetail {
			switch keyMsg.String() {
			case "esc":
				// Exit field detail view, back to field list
				m.IsViewingFieldDetail = false
				m.Err = nil
				return m, nil
			case "up", "k":
				// Scroll up in field detail view
				if m.FieldDetailScrollOffset > 0 {
					m.FieldDetailScrollOffset--
				}
				return m, nil
			case "down", "j":
				// Scroll down in field detail view
				fieldValue := ""
				for i, col := range m.DataPreviewAllColumns {
					if col == m.SelectedFieldForDetail && i < len(m.SelectedRowData) {
						fieldValue = m.SelectedRowData[i]
						break
					}
				}
				// Format field value same as in view (handles JSON formatting)
				fieldValue = utils.FormatFieldValue(fieldValue)

				// Calculate max scroll based on formatted field content and dynamic height
				// Must match calculation in query_views.go
				lines := len(strings.Split(fieldValue, "\n"))
				_, v := styles.DocStyle.GetFrameSize()
				availableHeight := m.Height - v - 12 // Same as view calculation
				if availableHeight < 5 {
					availableHeight = 5
				}
				maxScroll := lines - availableHeight
				maxScroll = max(maxScroll, 0)
				if m.FieldDetailScrollOffset < maxScroll {
					m.FieldDetailScrollOffset++
				}
				return m, nil
			case "left", "h":
				// Horizontal scroll left
				availableWidth := min(max(m.Width-10, 40), 200)
				scrollIncrement := max(availableWidth/4, 5) // Scroll by 1/4 of screen width, minimum 5
				m.FieldDetailHorizontalOffset = max(m.FieldDetailHorizontalOffset-scrollIncrement, 0)
				return m, nil
			case "right", "l":
				// Horizontal scroll right
				availableWidth := min(max(m.Width-10, 40), 200)
				scrollIncrement := max(availableWidth/4, 5) // Scroll by 1/4 of screen width, minimum 5
				m.FieldDetailHorizontalOffset += scrollIncrement
				return m, nil
			default:
				// Absorb other keys when in detail view
				return m, nil
			}
		}

		// Default mode: navigating the list of fields.
		switch keyMsg.String() {
		case "esc":
			// Return to data preview
			m.State = models.DataPreviewView
			m.Err = nil
			return m, nil
		case "enter":
			// Enter field detail view
			if selectedItem, ok := m.RowDetailList.SelectedItem().(models.FieldItem); ok {
				m.SelectedFieldForDetail = selectedItem.Name
				m.IsViewingFieldDetail = true
				// Reset scroll positions
				m.FieldDetailScrollOffset = 0
				m.FieldDetailHorizontalOffset = 0
			}
			return m, nil
		case "e":
			// Enter field edit mode
			if selectedItem, ok := m.RowDetailList.SelectedItem().(models.FieldItem); ok {
				m.EditingFieldName = selectedItem.Name
				m.OriginalFieldValue = selectedItem.Value

				// Find the field index
				for i, col := range m.DataPreviewAllColumns {
					if col == selectedItem.Name {
						m.EditingFieldIndex = i
						break
					}
				}

				// Initialize textarea with current value
				m.FieldTextarea.SetValue(selectedItem.Value)
				m.FieldTextarea.CursorStart()

				// Set responsive textarea size
				h, v := styles.DocStyle.GetFrameSize()
				textareaWidth := max(m.Width-h-4, 40)
				textareaHeight := max(m.Height-v-8, 5)
				m.FieldTextarea.SetWidth(textareaWidth)
				m.FieldTextarea.SetHeight(textareaHeight)

				m.FieldTextarea.Focus()
				m.IsEditingField = true
			}
			return m, nil
		}
	}

	// If not editing or viewing detail, update the list component
	if !m.IsEditingField && !m.IsViewingFieldDetail {
		m.RowDetailList, cmd = m.RowDetailList.Update(msg)
	}

	return m, cmd
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
