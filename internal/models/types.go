package models

import (
	"database/sql"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
)

// Application states
type ViewState int

const (
	DBTypeView ViewState = iota
	SavedConnectionsView
	ConnectionView
	SaveConnectionView
	EditConnectionView
	SchemaView
	TablesView
	ColumnsView
	QueryView
	QueryHistoryView
	DataPreviewView
	RowDetailView
	FieldDetailView
	IndexesView
	IndexDetailView
	RelationshipsView
)

// Sort directions
type SortDirection int

const (
	SortOff SortDirection = iota
	SortAsc
	SortDesc
)

// Database types
type DBType struct {
	Name   string
	Driver string
}

// Saved connection
type SavedConnection struct {
	Name          string `json:"name"`
	Driver        string `json:"driver"`
	ConnectionStr string `json:"connection_str"`
}

// Query history entry
type QueryHistoryEntry struct {
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database,omitempty"`
	Success   bool      `json:"success"`
	RowCount  int       `json:"row_count,omitempty"`
}

// Schema information
type SchemaInfo struct {
	Name        string
	Description string
}

// Table information
type TableInfo struct {
	Name        string
	Schema      string
	TableType   string
	RowCount    int64
	Description string
}

// List item
type Item struct {
	ItemTitle, ItemDesc string
}

func (i Item) Title() string       { return i.ItemTitle }
func (i Item) Description() string { return i.ItemDesc }
func (i Item) FilterValue() string { return i.ItemTitle }

// Main model
type Model struct {
	State                ViewState
	DBTypeList           list.Model
	SavedConnectionsList list.Model
	TextInput            textinput.Model
	NameInput            textinput.Model
	QueryInput           textinput.Model
	TablesList           list.Model
	ColumnsTable         table.Model
	QueryResultsTable    table.Model
	DataPreviewTable     table.Model
	IndexesTable         table.Model
	RelationshipsTable   table.Model
	SelectedDB           DBType
	ConnectionStr        string
	DB                   *sql.DB
	Err                  error
	Tables               []string
	TableInfos           []TableInfo
	SelectedTable        string
	Schemas              []SchemaInfo
	SelectedSchema       string
	SchemasList          list.Model
	IsLoadingSchemas     bool
	SavedConnections     []SavedConnection
	EditingConnectionIdx int
	QueryResult          string
	Width                int
	Height               int

	// Loading states
	IsTestingConnection bool
	IsConnecting        bool
	IsSavingConnection  bool
	IsLoadingTables     bool
	IsLoadingColumns    bool
	IsExecutingQuery    bool
	IsLoadingPreview    bool

	// Export states
	IsExporting        bool
	LastQueryColumns   []string
	LastQueryRows      [][]string
	LastPreviewColumns []string
	LastPreviewRows    [][]string

	// Spinner for animations
	Spinner spinner.Model

	// Search functionality
	SearchInput        textinput.Model
	IsSearchingTables  bool
	IsSearchingColumns bool
	OriginalTableItems []list.Item
	OriginalTableRows  []table.Row
	SearchTerm         string

	// Query history functionality
	QueryHistory     []QueryHistoryEntry
	QueryHistoryList list.Model
	IsViewingHistory bool

	// Row detail functionality
	SelectedRowData        []string
	SelectedRowIndex       int
	RowDetailCurrentPage   int
	RowDetailItemsPerPage  int
	RowDetailSelectedField int
	IsViewingFullText      bool
	FullTextScrollOffset   int
	FullTextLinesPerPage   int

	// Full text view pagination
	FullTextCurrentPage   int
	FullTextItemsPerPage  int
	FullTextSelectedField int

	// Individual field detail view
	SelectedFieldName           string
	SelectedFieldValue          string
	SelectedFieldIndex          int
	FieldDetailScrollOffset     int
	FieldDetailHorizontalOffset int
	FieldDetailLinesPerPage     int
	FieldDetailCharsPerLine     int

	// Field editing
	FieldTextarea      textarea.Model
	IsEditingField     bool
	OriginalFieldValue string

	// Index detail view
	SelectedIndexName       string
	SelectedIndexType       string
	SelectedIndexColumns    string
	SelectedIndexDefinition string

	// Data preview pagination
	DataPreviewCurrentPage  int
	DataPreviewItemsPerPage int
	DataPreviewTotalRows    int

	// Data preview horizontal scrolling
	DataPreviewScrollOffset int        // Current column offset
	DataPreviewVisibleCols  int        // Number of columns visible at once
	DataPreviewAllColumns   []string   // Store all column names
	DataPreviewAllRows      [][]string // Store all row data

	// Data preview filtering
	DataPreviewFilterActive bool            // Whether filter mode is active
	DataPreviewFilterValue  string          // Current filter text
	DataPreviewFilterInput  textinput.Model // Filter input field

	// Data preview sorting
	DataPreviewSortColumn    string        // Column to sort by
	DataPreviewSortDirection SortDirection // Current sort direction
	DataPreviewSortMode      bool          // Whether in column selection mode for sorting
}

// Message types for Bubble Tea
type ConnectResult struct {
	DB     *sql.DB
	Driver string
	Err    error
	Tables []string
	Schema string
}

type TestConnectionResult struct {
	Success bool
	Err     error
}

type ColumnsResult struct {
	Columns [][]string
	Err     error
}

type QueryResult struct {
	Columns  []string
	Rows     [][]string
	Err      error
	RowCount int
}

type DataPreviewResult struct {
	Columns   []string
	Rows      [][]string
	Err       error
	TotalRows int
}

type IndexesResult struct {
	Indexes [][]string
	Err     error
}

type RelationshipsResult struct {
	Relationships [][]string
	Err           error
}

type ClearResultMsg struct{}
type ClearErrorMsg struct{}

type ExportResult struct {
	Success  bool
	Err      error
	Filename string
	Format   string
}

type TestAndSaveResult struct {
	Success bool
	Err     error
	DB      *sql.DB
	Driver  string
	Tables  []string
	Schema  string
}

type FieldValueResult struct {
	Value string
	Err   error
}

type ClipboardResult struct {
	Success bool
	Err     error
}

type FieldUpdateResult struct {
	Success  bool
	Err      error
	ExitEdit bool
	NewValue string
}
