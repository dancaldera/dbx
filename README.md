# DBX - Database Explorer

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go)](https://go.dev/dl/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/danielcaldera/dbx)](https://goreportcard.com/report/github.com/danielcaldera/dbx)

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
├── main.go                     # Legacy main entry point
├── cmd/dbx/                    # CLI entrypoint (preferred)
├── internal/
│   ├── config/                 # Configuration and file storage
│   ├── database/               # Database operations and adapters
│   ├── models/                 # Core types and interfaces
│   ├── styles/                 # UI theming
│   ├── ui/                     # Bubble Tea model initialization
│   └── views/                  # UI view rendering
```

DBX follows a clean, modular architecture with well-separated concerns across configuration management, database operations, type definitions, UI styling, and component organization.


## Installation

### Prerequisites
- Go 1.24.5 or later

### Build from source
```bash
git clone <repository-url>
cd dbx
go mod tidy
go build -o dbx ./cmd/dbx
```

## Usage

Run the application:
```bash
./dbx
```

Or run directly with Go:
```bash
go run ./cmd/dbx
```

### Navigation Controls

| Key | Action |
|-----|--------|
| **↑/↓** | Navigate through lists and tables |
| **Enter** | Select item or confirm action |
| **Esc** | Go back to previous screen |
| **q** or **Ctrl+C** | Quit application |

### Data Operations

| Key | Action |
|-----|--------|
| **p** | Preview table data (first 10 rows) |
| **r** | Run custom SQL query |
| **i** | View indexes and constraints |
| **f** | View foreign key relationships |
| **Ctrl+F** or **/** | Search tables or columns |
| **Ctrl+H** | Access query history |

### Connection Management

| Key | Action |
|-----|--------|
| **s** | Save database connection |
| **e** | Edit saved connection |
| **d** | Delete saved connection or query |
| **F1** | Test database connection |
| **F2** | Validate, save and connect |

### Export Options

| Key | Action |
|-----|--------|
| **Ctrl+E** | Export to CSV |
| **Ctrl+J** | Export to JSON |

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
go build -o dbx ./cmd/dbx
```

### Run
```bash
go run ./cmd/dbx
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

## MVP Roadmap

### Core Features (Must Have)
- [x] 1. Connection persistence - Save and load database connections
- [x] 2. Query execution - Run custom SQL queries within the TUI
- [x] 3. Result display - Show query results in paginated table format
- [x] 4. Table data preview - Show sample rows from selected tables
- [x] 5. Export functionality - Export query results to CSV/JSON
- [x] 6. Connection testing - Validate connections before saving
- [x] 7. Error handling improvements - Better error messages and recovery
- [x] 8. Search functionality - Search tables and columns by name
- [x] 9. Query history - Save and recall previously executed queries

### Enhanced Database Support
- [x] 10. Database schema support - Handle multiple schemas in PostgreSQL
- [x] 11. View support - Browse and explore database views
- [x] 12. Index information - Display table indexes and constraints
- [x] 13. Foreign key relationships - Show table relationships
- [ ] 14. Stored procedures - List and describe stored procedures/functions
- [ ] 15. SSL/TLS support - Secure connections with certificate validation

### Advanced Features
- [ ] 16. Multi-connection support - Work with multiple databases simultaneously
- [ ] 17. Data filtering - Basic filtering on table data preview
- [ ] 18. Bookmark tables - Mark frequently accessed tables
- [ ] 19. Connection string validation - Real-time validation with hints
- [ ] 20. Configuration file - Settings for defaults and preferences
- [ ] 21. Performance monitoring - Basic query execution time tracking

### Extended Database Support
- [ ] 22. Redis support - Connect to and explore Redis data structures
- [ ] 23. MongoDB support - Browse MongoDB collections and documents
- [ ] 24. Query builder - Visual query construction for SQL databases
- [ ] 25. Database migration tools - Simple schema migration utilities

### Nice to Have Features
- [ ] 26. Keyboard shortcuts help - In-app help screen
- [ ] 27. Loading indicators - Progress bars for long-running operations
- [ ] 28. Plugin system - Extensible architecture for custom features
- [ ] 29. Themes support - Multiple color schemes and styling options
- [ ] 30. Cloud database support - AWS RDS, Google Cloud SQL, etc.

## Code Quality Standards

- **File size limit**: All source files must be under 500 lines
- **Separation of concerns**: Clear module boundaries and responsibilities
- **Testing**: Tests alongside code as `*_test.go` files
- **Formatting**: Follow standard Go formatting (`go fmt`)

## License

MIT License © 2025 Daniel Caldera. See `LICENSE` for details.
