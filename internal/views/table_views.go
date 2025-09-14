package views

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/dancaldera/dbx/internal/models"
	"github.com/dancaldera/dbx/internal/styles"
)

// SchemaView renders the schema selection screen
func SchemaView(m models.Model) string {
	var content string

	if m.IsLoadingSchemas {
		loadingMsg := "⏳ Loading schemas..."
		content = m.SchemasList.View() + "\n" + loadingMsg
	} else if len(m.Schemas) == 0 {
		emptyMsg := styles.InfoStyle.Render("🗂️ No additional schemas found.\n\nUsing default schema.")
		content = m.SchemasList.View() + "\n" + emptyMsg
	} else {
		content = m.SchemasList.View()
	}

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": select schema • " +
			styles.KeyStyle.Render("esc") + ": back",
	)

	return styles.DocStyle.Render(content + "\n" + helpText)
}

// TablesView renders the tables listing screen
func TablesView(m models.Model) string {
	var elements []string

	if m.IsLoadingColumns {
		loadingMsg := "⏳ Loading table columns..."
		elements = append(elements, m.TablesList.View())
		elements = append(elements, loadingMsg)
	} else if len(m.Tables) == 0 {
		emptyMsg := styles.InfoStyle.Render("📋 No tables found in this database.")
		elements = append(elements, m.TablesList.View())
		elements = append(elements, emptyMsg)
	} else {
		// Show tables list without success banner
		elements = append(elements, m.TablesList.View())
	}

	content := lipgloss.JoinVertical(lipgloss.Left, elements...)

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": preview data • " +
			styles.KeyStyle.Render("v") + ": view columns • " +
			styles.KeyStyle.Render("f") + ": relationships • " +
			styles.KeyStyle.Render("esc") + ": disconnect")

	return styles.DocStyle.Render(content + "\n" + helpText)
}

// ColumnsView renders the table columns display screen
func ColumnsView(m models.Model) string {
	content := styles.TitleStyle.Render(fmt.Sprintf("Columns of table: %s", m.SelectedTable)) + "\n\n"
	content += m.ColumnsTable.View()

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("↑/↓") + ": navigate • " +
			styles.KeyStyle.Render("esc") + ": back to tables")

	content += "\n" + helpText
	return styles.DocStyle.Render(content)
}

// IndexesView renders the table indexes and constraints screen
func IndexesView(m models.Model) string {
	content := styles.TitleStyle.Render(fmt.Sprintf("🔑 Indexes & Constraints: %s", m.SelectedTable)) + "\n\n"

	if m.Err != nil {
		content += styles.ErrorStyle.Render("❌ "+m.Err.Error()) + "\n\n"
	}

	// Show the indexes table
	content += m.IndexesTable.View() + "\n\n"

	// Help text
	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("↑/↓") + ": navigate • " +
			styles.KeyStyle.Render("enter") + ": view details • " +
			styles.KeyStyle.Render("esc") + ": back to columns")

	content += "\n" + helpText
	return styles.DocStyle.Render(content)
}

// RelationshipsView renders the foreign key relationships screen
func RelationshipsView(m models.Model) string {
	content := styles.TitleStyle.Render("🔗 Foreign Key Relationships") + "\n\n"

	if m.Err != nil {
		content += styles.ErrorStyle.Render("❌ "+m.Err.Error()) + "\n\n"
	}

	// Show the relationships table
	content += m.RelationshipsTable.View() + "\n\n"

	// Help text
	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("↑/↓") + ": navigate • " +
			styles.KeyStyle.Render("esc") + ": back to tables")

	content += "\n" + helpText
	return styles.DocStyle.Render(content)
}