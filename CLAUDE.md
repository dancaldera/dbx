# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is **DBX**, a terminal-based database explorer built in Go. It's a TUI (Terminal User Interface) application that allows users to connect to different types of databases (PostgreSQL, MySQL, SQLite) and explore their structure interactively.

## Architecture

### Modular Architecture

The application has been refactored from a single-file monolith into a well-organized modular structure:

```
dbx/
├── main.go                     # Main application entry point
├── cmd/dbx/                    # Command-line application (alternative)
├── internal/
│   ├── config/                 # Configuration and file storage
│   │   └── storage.go          # Connection and history persistence
│   ├── database/               # Database operations
│   │   ├── operations.go       # Core database functions
│   │   └── columns.go          # Column, index, and relationship queries
│   ├── models/                 # Data models and types
│   │   ├── types.go            # Core type definitions
│   │   └── model.go            # Bubble Tea model interface
│   ├── styles/                 # UI theming and styles
│   │   └── theme.go            # Magenta theme definitions
│   ├── ui/                     # UI initialization
│   │   └── init.go             # Model initialization
│   └── handlers/               # Event handlers (optional)
│       └── update.go           # Update logic handlers
├── go.mod
└── go.sum
```

### Core Components

- **Modular design** with clear separation of concerns
- **Package-based organization** for maintainability and testing
- **State machine pattern** with main view states:
  - `dbTypeView`: Database type selection screen
  - `connectionView`: Connection string input screen
  - `schemaView`: Schema selection for PostgreSQL (when multiple schemas exist)
  - `tablesView`: Display available tables in connected database
  - `columnsView`: Show column details for selected table
  - `queryView`: SQL query execution interface
  - `queryHistoryView`: Browse and reuse previous queries

### Package Responsibilities

- **`main.go`**: Application entry point and Bubble Tea event loop
- **`internal/config`**: Configuration file management, connection/history persistence
- **`internal/database`**: All database operations, queries, and connection management
- **`internal/models`**: Type definitions, data structures, and interfaces
- **`internal/styles`**: UI theming, colors, and styling definitions
- **`internal/ui`**: UI component initialization and setup

### Key Dependencies

- `github.com/charmbracelet/bubbletea`: TUI framework and event handling
- `github.com/charmbracelet/bubbles`: Pre-built UI components (list, table, textinput)
- `github.com/charmbracelet/lipgloss`: Styling and layout
- Database drivers: `github.com/lib/pq` (PostgreSQL), `github.com/go-sql-driver/mysql` (MySQL), `github.com/mattn/go-sqlite3` (SQLite)

### Application Flow

1. User selects database type from list OR selects saved connection
2. User enters connection string specific to chosen database type (if not using saved connection)
3. Application connects and queries available tables
4. User can select tables to view their column structure
5. Navigation between states using keyboard shortcuts (Enter, Esc, s for saved connections, n for new connection)

### Connection Persistence

The application now supports saving and loading database connections:
- **Config Location**: `~/.dbx/connections.json`
- **Save Connection**: Press `s` in connection view after entering connection string
- **Load Connections**: Press `s` from main menu to view saved connections
- **Auto-connect**: Select saved connection with Enter to automatically connect

### Schema Support

For PostgreSQL databases with multiple schemas:
- **Automatic Detection**: Application detects available schemas on connection
- **Schema Selection**: If multiple schemas exist, shows schema selection screen
- **Schema Context**: All table operations work within the selected schema
- **Default Behavior**: Single schema databases (or MySQL/SQLite) skip schema selection

### View Support

The application supports database views across all database types:
- **View Detection**: Automatically detects and lists database views alongside tables
- **Visual Indicators**: Views display with 👁️ emoji, tables with 📊 emoji
- **Cross-Database**: PostgreSQL, MySQL, and SQLite view support with database-specific queries

## Development Commands

### Build and Run
```bash
# Build the application
go build .

# Run directly 
go run .

# Install dependencies
go mod tidy

# Update dependencies
go mod download
```

### Code Quality
```bash
# Format code
go fmt ./...

# Vet for issues
go vet ./...

# Run tests (currently no tests exist)
go test ./...
```

## Known Issues

None currently. The application builds and runs successfully.

## Database Connection Examples

- **PostgreSQL**: `postgres://user:password@localhost/dbname?sslmode=disable`
- **MySQL**: `user:password@tcp(localhost:3306)/dbname`  
- **SQLite**: `/path/to/database.db`

## Navigation Controls

- `↑/↓`: Navigate lists and tables
- `Enter`: Select/confirm action
- `Esc`: Go back to previous view
- `s`: Access saved connections (from main menu) or save current connection (from connection view)
- `n`: Start new database connection
- `p`: Preview table data (first 10 rows)
- `r`: Run custom SQL query
- `i`: View indexes and constraints (from columns view)
- `f`: View foreign key relationships (from tables view)
- `Ctrl+H`: Access query history (from query view)
- `d`: Delete saved connection (in saved connections view) or query history entry (in query history view)
- `Ctrl+F` or `/`: Search tables or columns by name
- `F1`: Test database connection (without saving)
- `F2`: Validate, save and connect to database
- `Ctrl+E`: Export query results or table preview to CSV
- `Ctrl+J`: Export query results or table preview to JSON
- `q` or `Ctrl+C`: Quit application