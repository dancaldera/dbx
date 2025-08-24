# DBX - Database Explorer

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go)](https://go.dev/dl/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/dancaldera/dbx)](https://goreportcard.com/report/github.com/dancaldera/dbx)

A terminal-based database explorer built with Go and Bubble Tea. DBX provides an interactive TUI for connecting to and exploring database structures across PostgreSQL, MySQL, and SQLite databases.

## Features

- **Multi-database support**: PostgreSQL, MySQL, and SQLite
- **Interactive TUI**: Clean, keyboard-driven interface
- **Connection management**: Save, edit, and switch between database connections
- **Schema exploration**: Browse tables, views, columns, indexes, and relationships
- **Data preview**: Quick view of table contents without writing SQL
- **SQL query execution**: Run custom queries with formatted results
- **Export functionality**: Export data to CSV and JSON formats
- **Query history**: Save and recall previously executed queries

## Architecture

```
dbx/
├── main.go                     # Main entry point with update logic
├── internal/
│   ├── config/                 # Configuration and file storage
│   ├── database/               # Database operations and adapters
│   ├── models/                 # Core types and interfaces
│   ├── styles/                 # UI theming
│   ├── utils/                  # Helper functions and utilities
│   └── views/                  # UI view rendering
```

DBX follows a clean, modular architecture with well-separated concerns across configuration management, database operations, type definitions, UI styling, utility functions, and view rendering. The main application logic and Bubble Tea update handlers are implemented in `main.go` using the `appModel` wrapper pattern.

### Utils Package

The `utils` package provides centralized helper functions organized by domain:

- **Database utilities**: Schema detection, SQL generation, primary key finding
- **Data processing**: Sorting parameters, table info creation, list management
- **UI utilities**: Table creation, column width calculation, list updates
- **Mathematical utilities**: Min/max functions, pagination calculations
- **Type inference**: Data type detection, datetime parsing, value sanitization
- **Timeout utilities**: Async command helpers with timeouts


## Installation

### Prerequisites
- Go 1.24.5 or later

### Build from source
```bash
git clone <repository-url>
cd dbx
go mod tidy
go build -o dbx main.go
```

## Usage

Run the application:
```bash
./dbx
```

Or run directly with Go:
```bash
go run main.go
```

### Navigation Controls

Global

- **↑/↓**: Navigate lists and tables
- **Enter**: Select or confirm
- **Esc**: Go back
- **q/Ctrl+C**: Quit

DB Type Selection

- **enter**: Select database type
- **s**: Open saved connections
- **q**: Quit

Saved Connections

- **enter**: Connect
- **esc**: Back

Connection Form

- **Enter**: Save and connect
- **F1**: Test connection
- **Tab**: Switch fields
- **Esc**: Back

Schemas

- **enter**: Select schema
- **esc**: Back

Tables

- **enter**: Preview data
- **v**: View columns
- **f**: Relationships
- **esc**: Disconnect

Columns

- **↑/↓**: Navigate
- **esc**: Back to tables

Data Preview

- **hjkl**: Navigate cells
- **←/→**: Previous/next page
- **enter**: Row details
- **/**: Filter
- **s**: Sort columns
- **r**: Reload
- Filter active: **enter** apply, **esc** cancel
- Sort mode: **↑/↓** select column, **enter** toggle asc/desc, **esc** exit
- **esc**: Back

Row Details

- Field list: **↑/↓** navigate, **enter** view field, **e** edit, **esc** back
- Field detail: **↑↓/jk** scroll, **←→/hl** horizontal scroll, **esc** back
- Edit field: **Ctrl+S** save, **Ctrl+K** clear, **Esc** cancel

Query Runner

- **Enter**: Execute query
- **Tab**: Switch focus
- **↑/↓**: Navigate results
- **Ctrl+E**: Export CSV
- **Ctrl+J**: Export JSON
- **Esc**: Back to tables

Query History

- **enter**: Use query
- **d**: Delete
- **esc**: Back

### Connection Strings

#### PostgreSQL
```
postgres://username:password@localhost:5432/database_name?sslmode=disable
```

#### MySQL
```
username:password@tcp(localhost:3306)/database_name
```

#### SQLite
```
/path/to/your/database.db
```

## Workflow

1. **Select Database Type**: Choose from PostgreSQL, MySQL, or SQLite
2. **Enter Connection String**: Provide the appropriate connection string for your database
3. **Browse Tables**: View all available tables in the connected database
4. **Explore Data**: Preview table data (first 10 rows) or view column structure
5. **Run Queries**: Execute custom SQL queries with formatted table results
6. **Export Data**: Export query results or table previews to CSV/JSON format

## Database Features

### Schema Support
- **PostgreSQL**: Full schema support with automatic detection and selection interface
- **MySQL**: Database-level organization (no schema selection needed)
- **SQLite**: Uses default `main` schema

### Views and Tables
- Views and tables are listed with clear labels (no icons)
- Full column structure browsing for both tables and views
- Data preview functionality for quick content inspection

### Indexes and Constraints
- Complete index information (primary keys, unique indexes, regular indexes)
- Constraint details (foreign keys, primary keys, check constraints)
- Visual organization with type, affected columns, and full definitions

### Foreign Key Relationships
- Comprehensive relationship visualization across all tables
- Shows source/target tables and columns with constraint names
- Cross-database support for all relationship types

### Connection Validation
- Pre-save connection testing with 10-second timeout
- Database-specific error messages with troubleshooting hints
- Format validation for connection strings

### Export Capabilities
- **CSV**: Comma-separated values with headers
- **JSON**: Array of objects format
- Automatic timestamped filenames
- Export from query results or table previews

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- Database drivers for PostgreSQL, MySQL, and SQLite

## Development

### Build
```bash
go build
```

### Run
```bash
go run main.go
```

### Run Production
```bash
chmod +x ./dbx
./dbx
```

### Development Commands
```bash
# Install dependencies
go mod tidy

# Format code
go fmt ./...

# Static analysis
go vet ./...

# Run tests
go test ./...
```

## Contributing

Contributions are welcome! Please review `CONTRIBUTING.md` for guidelines on setting up your environment, coding style, testing, and submitting pull requests.

## Code Quality Standards

- **File size limit**: All source files must be under 500 lines
- **Separation of concerns**: Clear module boundaries and responsibilities
- **Testing**: Tests alongside code as `*_test.go` files
- **Formatting**: Follow standard Go formatting (`go fmt`)

## License

MIT License © 2025 Daniel Caldera. See `LICENSE` for details.
