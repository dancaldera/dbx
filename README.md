# DBX - Database Explorer

A terminal-based database explorer built with Go and Bubble Tea. DBX provides an interactive TUI for connecting to and exploring database structures across multiple database types.

## Features

- **Multi-database support**: PostgreSQL, MySQL, and SQLite
- **Interactive TUI**: Clean, keyboard-driven interface
- **Database exploration**: Browse tables and view column details
- **Connection management**: Easy switching between database connections

## Installation

### Prerequisites
- Go 1.24.5 or later

### Build from source
```bash
git clone <repository-url>
cd dbx
go mod tidy
go build .
```

## Usage

Run the application:
```bash
./dbx
```

### Navigation
- **↑/↓**: Navigate through lists and tables
- **Enter**: Select item or confirm action
- **Esc**: Go back to previous screen
- **n**: Start new database connection
- **q** or **Ctrl+C**: Quit application

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
4. **Explore Columns**: Select a table to see its column structure, types, and constraints

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- Database drivers for PostgreSQL, MySQL, and SQLite

## Development

### Build
```bash
go build .
```

### Run
```bash
go run .
```

### Format code
```bash
go fmt ./...
```

### Vet for issues
```bash
go vet ./...
```

## MVP Roadmap

### Core Features (Must Have)
- [x] 1. Connection persistence - Save and load database connections
- [ ] 2. Query execution - Run custom SQL queries within the TUI
- [ ] 3. Result display - Show query results in paginated table format
- [ ] 4. Export functionality - Export query results to CSV/JSON
- [ ] 5. Connection testing - Validate connections before saving
- [ ] 6. Error handling improvements - Better error messages and recovery
- [ ] 7. Table data preview - Show sample rows from selected tables
- [ ] 8. Search functionality - Search tables and columns by name
- [ ] 9. Connection history - Recently used connections list

### Enhanced Database Support
- [ ] 10. Database schema support - Handle multiple schemas in PostgreSQL
- [ ] 11. View support - Browse and explore database views
- [ ] 12. Index information - Display table indexes and constraints
- [ ] 13. Foreign key relationships - Show table relationships
- [ ] 14. Stored procedures - List and describe stored procedures/functions

### User Experience
- [ ] 15. Configuration file - Settings for themes, defaults, and preferences
- [ ] 16. Keyboard shortcuts help - In-app help screen
- [ ] 17. Connection string validation - Real-time validation with hints
- [ ] 18. Themes support - Multiple color schemes and styling options
- [ ] 19. Responsive layout - Better handling of different terminal sizes
- [ ] 20. Loading indicators - Progress bars for long-running operations

### Advanced Features
- [ ] 21. Query history - Save and recall previously executed queries
- [ ] 22. Bookmark tables - Mark frequently accessed tables
- [ ] 23. Data filtering - Basic filtering on table data preview
- [ ] 24. SSL/TLS support - Secure connections with certificate validation
- [ ] 25. Multi-connection support - Work with multiple databases simultaneously

### Bonus Features (Nice to Have)
- [ ] 26. Query builder - Visual query construction
- [ ] 27. Database migration tools - Simple schema migration utilities
- [ ] 28. Performance monitoring - Basic query execution time tracking
- [ ] 29. Plugin system - Extensible architecture for custom features
- [ ] 30. Cloud database support - MongoDB, DynamoDB, etc.

## License

This project is licensed under the MIT License - see the [MIT License](https://opensource.org/licenses/MIT) for details.