package views

import (
	"fmt"

	"github.com/dancaldera/dbx/internal/models"
	"github.com/dancaldera/dbx/internal/styles"
)

// SchemaView renders the schema selection screen
func SchemaView(m models.Model) string {
	builder := NewViewBuilder()

	// Add loading or empty state
	if m.IsLoadingSchemas {
		builder.WithStatus("⏳ Loading schemas...", StatusLoading)
	} else if len(m.Schemas) == 0 {
		emptyState := RenderEmptyState("🗂️", "No additional schemas found.\n\nUsing default schema.")
		builder.WithContent(m.SchemasList.View(), emptyState)
	} else {
		builder.WithContent(m.SchemasList.View())
	}

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": select schema • " +
			styles.KeyStyle.Render("esc") + ": back",
	)

	return builder.WithHelp(helpText).Render()
}

// TablesView renders the tables listing screen
func TablesView(m models.Model) string {
	builder := NewViewBuilder()

	if m.IsLoadingColumns {
		builder.WithStatus("⏳ Loading table columns...", StatusLoading).
			WithContent(m.TablesList.View())
	} else if len(m.Tables) == 0 {
		emptyState := RenderEmptyState("📋", "No tables found in this database.")
		builder.WithContent(m.TablesList.View(), emptyState)
	} else {
		// Show tables list without success banner
		builder.WithContent(m.TablesList.View())
	}

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("enter") + ": preview data • " +
			styles.KeyStyle.Render("v") + ": view columns • " +
			styles.KeyStyle.Render("f") + ": relationships\n" +
			styles.KeyStyle.Render("r") + ": run SQL queries • " +
			styles.KeyStyle.Render("ctrl+h") + ": view query history • " +
			styles.KeyStyle.Render("esc") + ": disconnect",
	)

	return builder.WithHelp(helpText).Render()
}

// ColumnsView renders the table columns display screen
func ColumnsView(m models.Model) string {
	title := fmt.Sprintf("Columns of table: %s", m.SelectedTable)

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("↑/↓") + ": navigate • " +
			styles.KeyStyle.Render("esc") + ": back to tables",
	)

	return NewViewBuilder().
		WithTitle(title).
		WithContent(m.ColumnsTable.View()).
		WithHelp(helpText).
		Render()
}

// IndexesView renders the table indexes and constraints screen
func IndexesView(m models.Model) string {
	title := fmt.Sprintf("🔑 Indexes & Constraints: %s", m.SelectedTable)
	builder := NewViewBuilder().WithTitle(title)

	// Add error status if present
	if m.Err != nil {
		builder.WithStatus("❌ "+m.Err.Error(), StatusError)
	}

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("↑/↓") + ": navigate • " +
			styles.KeyStyle.Render("enter") + ": view details • " +
			styles.KeyStyle.Render("esc") + ": back to columns",
	)

	return builder.
		WithContent(m.IndexesTable.View()).
		WithHelp(helpText).
		Render()
}

// RelationshipsView renders the foreign key relationships screen
func RelationshipsView(m models.Model) string {
	builder := NewViewBuilder().WithTitle("🔗 Foreign Key Relationships")

	// Add error status if present
	if m.Err != nil {
		builder.WithStatus("❌ "+m.Err.Error(), StatusError)
	}

	helpText := styles.HelpStyle.Render(
		styles.KeyStyle.Render("↑/↓") + ": navigate • " +
			styles.KeyStyle.Render("esc") + ": back to tables",
	)

	return builder.
		WithContent(m.RelationshipsTable.View()).
		WithHelp(helpText).
		Render()
}
