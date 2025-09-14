package state

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancaldera/dbx/internal/models"
	"github.com/dancaldera/dbx/internal/styles"
	"github.com/dancaldera/dbx/internal/utils"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(styles.AccentBlue)
)

// HandleDataPreviewViewUpdate handles all updates for the DataPreviewView state.
// Note: The 'enter' key to switch to RowDetailView is handled in main.go due to a dependency on the FieldItemDelegate.
func HandleDataPreviewViewUpdate(m models.Model, msg tea.Msg) (models.Model, tea.Cmd) {
	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Handle filter mode first, as it captures input
		if m.DataPreviewFilterActive {
			switch keyMsg.String() {
			case "enter":
				// Apply filter
				m.DataPreviewFilterValue = m.DataPreviewFilterInput.Value()
				m.DataPreviewFilterActive = false
				m.DataPreviewFilterInput.Blur()
				m.DataPreviewCurrentPage = 0 // Reset to first page
				return m, utils.LoadDataPreviewWithFilter(m.DB, m.SelectedDB, m.SelectedTable, m.SelectedSchema, m.DataPreviewItemsPerPage, m.DataPreviewFilterValue, m.DataPreviewAllColumns, m.DataPreviewSortDirection, m.DataPreviewSortColumn)
			case "esc":
				// Cancel filter
				m.DataPreviewFilterActive = false
				m.DataPreviewFilterInput.Blur()
				m.DataPreviewFilterInput.SetValue("")
				return m, nil
			default:
				// Update filter input
				m.DataPreviewFilterInput, cmd = m.DataPreviewFilterInput.Update(msg)
				return m, cmd
			}
		}

		// Handle sort mode if not in filter mode
		if m.DataPreviewSortMode {
			// Safeguard: Exit sort mode if no columns available
			if len(m.DataPreviewAllColumns) == 0 {
				m.DataPreviewSortMode = false
				return m, nil
			}
			switch keyMsg.String() {
			case "up", "k":
				// Move to previous column for sorting
				currentIdx := -1
				for i, col := range m.DataPreviewAllColumns {
					if col == m.DataPreviewSortColumn {
						currentIdx = i
						break
					}
				}
				if currentIdx > 0 {
					m.DataPreviewSortColumn = m.DataPreviewAllColumns[currentIdx-1]
				}
				return m, nil
			case "down", "j":
				// Move to next column for sorting
				currentIdx := -1
				for i, col := range m.DataPreviewAllColumns {
					if col == m.DataPreviewSortColumn {
						currentIdx = i
						break
					}
				}
				if currentIdx >= 0 && currentIdx < len(m.DataPreviewAllColumns)-1 {
					m.DataPreviewSortColumn = m.DataPreviewAllColumns[currentIdx+1]
				} else if currentIdx == -1 && len(m.DataPreviewAllColumns) > 0 {
					m.DataPreviewSortColumn = m.DataPreviewAllColumns[0]
				}
				return m, nil
			case "enter":
				// Toggle sort direction and apply
				switch m.DataPreviewSortDirection {
				case models.SortOff:
					m.DataPreviewSortDirection = models.SortAsc
				case models.SortAsc:
					m.DataPreviewSortDirection = models.SortDesc
				case models.SortDesc:
					m.DataPreviewSortDirection = models.SortOff
					m.DataPreviewSortColumn = ""
				}
				m.DataPreviewSortMode = false
				m.DataPreviewCurrentPage = 0 // Reset page when sorting changes
				return m, utils.LoadDataPreviewWithSort(m.DB, m.SelectedDB, m.SelectedTable, m.SelectedSchema, m.DataPreviewItemsPerPage, m.DataPreviewCurrentPage, m.DataPreviewSortDirection, m.DataPreviewSortColumn, m.DataPreviewFilterValue, m.DataPreviewAllColumns, m.DataPreviewTotalRows)
			case "esc":
				// Exit sort mode
				m.DataPreviewSortMode = false
				return m, nil
			}
			return m, nil // Absorb all other keys in sort mode
		}

		// Normal navigation mode (not filtering or sorting)
		switch keyMsg.String() {
		case "esc":
			// Go back to the tables view
			m.State = models.TablesView
			return m, nil
		case "/":
			// Start filter mode
			m.DataPreviewFilterActive = true
			m.DataPreviewFilterInput.Focus()
			return m, nil
		case "s":
			// Start sort mode
			if len(m.DataPreviewAllColumns) == 0 {
				return m, nil // No columns to sort
			}
			m.DataPreviewSortMode = true
			if m.DataPreviewSortColumn == "" && len(m.DataPreviewAllColumns) > 0 {
				m.DataPreviewSortColumn = m.DataPreviewAllColumns[0] // Default to first column
			}
			return m, nil
		case "r":
			// Reload/refresh data preview
			return m, utils.LoadDataPreview(m.DB, m.SelectedDB, m.SelectedTable, m.SelectedSchema, m.DataPreviewItemsPerPage, m.DataPreviewSortDirection, m.DataPreviewSortColumn)
		case "left":
			// Previous page
			if m.DataPreviewCurrentPage > 0 {
				m.DataPreviewCurrentPage--
				return m, utils.LoadDataPreviewWithPagination(m.DB, m.SelectedDB, m.SelectedTable, m.SelectedSchema, m.DataPreviewItemsPerPage, m.DataPreviewCurrentPage, m.DataPreviewSortDirection, m.DataPreviewSortColumn, m.DataPreviewFilterValue, m.DataPreviewAllColumns, m.DataPreviewTotalRows)
			}
			return m, nil
		case "right":
			// Next page
			totalPages := utils.CalculateTotalPages(m.DataPreviewTotalRows, m.DataPreviewItemsPerPage)
			if m.DataPreviewCurrentPage < totalPages-1 {
				m.DataPreviewCurrentPage++
				return m, utils.LoadDataPreviewWithPagination(m.DB, m.SelectedDB, m.SelectedTable, m.SelectedSchema, m.DataPreviewItemsPerPage, m.DataPreviewCurrentPage, m.DataPreviewSortDirection, m.DataPreviewSortColumn, m.DataPreviewFilterValue, m.DataPreviewAllColumns, m.DataPreviewTotalRows)
			}
			return m, nil
		case "h":
			// Scroll left (show previous columns)
			if m.DataPreviewScrollOffset > 0 {
				m.DataPreviewScrollOffset--
				m = utils.CreateDataPreviewTable(m)
			}
			return m, nil
		case "l":
			// Scroll right (show next columns)
			totalCols := len(m.DataPreviewAllColumns)
			if m.DataPreviewScrollOffset+m.DataPreviewVisibleCols < totalCols {
				m.DataPreviewScrollOffset++
				m = utils.CreateDataPreviewTable(m)
			}
			return m, nil
		}
	}

	// If no specific key was handled, pass the message to the table component
	// for internal navigation (e.g., moving the cursor up and down).
	m.DataPreviewTable, cmd = m.DataPreviewTable.Update(msg)
	return m, cmd
}

// FieldItemDelegate renders field name/value with a right-aligned type badge.
type FieldItemDelegate struct{}

func (d FieldItemDelegate) Height() int                               { return 1 }
func (d FieldItemDelegate) Spacing() int                              { return 0 }
func (d FieldItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d FieldItemDelegate) Render(w io.Writer, m list.Model, index int, it list.Item) {
	fi, ok := it.(models.FieldItem)
	if !ok {
		return
	}
	width := m.Width()
	if width <= 0 {
		return
	}

	// Determine type
	t := utils.InferFieldType(fi.Value)

	// Compose the display string: Name: value [Type]
	namePart := fi.Name + ": "
	badge := styles.TypeBadgeStyle.Render("[" + t + "]")
	single := utils.SanitizeValueForDisplay(fi.Value)

	// Calculate budget for value to fit within width
	budget := width - lipgloss.Width(namePart) - 1 - lipgloss.Width(badge)
	budget = utils.Max(budget, 0)
	val := utils.TruncateWithEllipsis(single, budget, "...")

	str := namePart + val + " " + badge

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
