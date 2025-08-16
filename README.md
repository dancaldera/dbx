# DBX - Database Explorer

A terminal-based database explorer built with Go and Bubble Tea. DBX provides an interactive TUI for connecting to and exploring database structures across multiple database types.

## Architecture

DBX has been refactored from a monolithic structure into a clean, modular architecture with well-separated concerns:

- **Configuration Management**: Persistent storage for connections and query history
- **Database Operations**: Abstracted database interactions with support for PostgreSQL, MySQL, and SQLite
- **Models & Types**: Clean type definitions and data structures
- **UI Styling**: Consistent theming and visual design
- **Modular Components**: Reusable and testable code organization

## Features

- **Multi-database support**: PostgreSQL, MySQL, and SQLite
- **Interactive TUI**: Clean, keyboard-driven interface
- **Database exploration**: Browse tables and view column details
- **Connection management**: Save, edit, delete, and switch between database connections
- **Data preview**: Quick view of table contents without writing SQL
- **SQL query execution**: Run custom queries with formatted table results

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
- **p**: Preview table data (first 10 rows)
- **r**: Run custom SQL query
- **i**: View indexes and constraints (from columns view)
- **f**: View foreign key relationships (from tables view)
- **Ctrl+F** or **/**: Search tables (in tables view) or columns (in columns view)
- **Ctrl+H**: Access query history (in query view)
- **d**: Delete saved connection (in saved connections view) or query history entry (in query history view)
- **F1**: Test database connection (in connection view)
- **F2**: Validate, save and connect to database (in connection view)
- **s**: Save database connection (from tables view)
- **e**: Edit saved connection (in saved connections view)
- **d**: Delete saved connection (in saved connections view)
- **Ctrl+E**: Export query results or table preview to CSV
- **Ctrl+J**: Export query results or table preview to JSON
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
4. **Explore Data**: Preview table data (first 10 rows) or view column structure
5. **Run Queries**: Execute custom SQL queries with formatted table results
6. **Export Data**: Export query results or table previews to CSV/JSON format

## Schema Support

DBX provides comprehensive schema support for PostgreSQL databases:

### PostgreSQL Schema Features
- **Automatic Detection**: Detects all available schemas in PostgreSQL databases
- **Schema Selection**: When multiple schemas exist, presents a selection interface
- **Schema Navigation**: Browse tables within specific schemas
- **Schema Information**: Displays schema names in table listings for non-public schemas

### Schema Workflow
1. **Single Schema**: If only one schema exists (usually `public`), proceeds directly to table listing
2. **Multiple Schemas**: Shows schema selection screen with descriptions
3. **Schema Selection**: Choose schema to explore its tables and structure
4. **Schema Context**: All subsequent table operations work within the selected schema

### Supported Schemas
- **PostgreSQL**: Full schema support with selection interface
- **MySQL**: Uses database-level organization (no schema selection needed)
- **SQLite**: Uses default `main` schema

## View Support

DBX provides comprehensive view support across all supported database types:

### Database View Features
- **View Detection**: Automatically detects and displays database views alongside tables
- **Visual Distinction**: Views are marked with 👁️ emoji, tables with 📊 emoji
- **Cross-Database Support**: Works with PostgreSQL, MySQL, and SQLite views

### View Workflow
1. **View Listing**: Views appear in the table listing with distinct visual indicators
2. **View Selection**: Select views the same way as tables using arrow keys and Enter
3. **View Columns**: View column structure the same way as tables
4. **View Preview**: Preview view data the same way as table data

## Export Functionality

DBX supports exporting query results and table previews to multiple formats:

### Export Formats
- **CSV**: Comma-separated values format with headers
- **JSON**: JavaScript Object Notation format as an array of objects

### Export Usage
- **From Query Results**: After executing a query, press `Ctrl+E` for CSV or `Ctrl+J` for JSON
- **From Table Preview**: When viewing table data preview, press `Ctrl+E` for CSV or `Ctrl+J` for JSON

### Export Files
- Files are saved in the current working directory
- Automatic filename generation includes timestamps
- Query results: `query_result_YYYYMMDD_HHMMSS.csv/json`
- Table previews: `tablename_YYYYMMDD_HHMMSS.csv/json`

## Index and Constraint Information

DBX provides comprehensive index and constraint information for all supported database types:

### Index and Constraint Features
- **Index Display**: Shows all indexes on a table including primary keys, unique indexes, and regular indexes
- **Constraint Information**: Displays table constraints including foreign keys, primary keys, and check constraints
- **Cross-Database Support**: Works with PostgreSQL, MySQL, and SQLite databases
- **Visual Organization**: Displays index type, affected columns, and full definition

### Index and Constraint Workflow
1. **Navigate to Columns**: Select a table and view its column structure
2. **Access Index View**: Press `i` to view indexes and constraints for the current table
3. **Review Information**: Browse through all indexes and constraints with their details
4. **Navigate Back**: Press `Esc` to return to the columns view

### Supported Information
- **PostgreSQL**: Full index and constraint support including schema-aware queries
- **MySQL**: Index statistics and constraint information from information_schema
- **SQLite**: Pragma-based index information and foreign key constraints

### Index Display Columns
- **Index Name**: The name of the index or constraint
- **Type**: Index type (PRIMARY, UNIQUE, INDEX) or constraint type (FOREIGN KEY, etc.)
- **Columns**: Which table columns are affected by the index or constraint
- **Definition**: Full SQL definition or description of the index or constraint

## Foreign Key Relationships

DBX provides comprehensive foreign key relationship visualization across all supported database types:

### Foreign Key Features
- **Relationship Display**: Shows all foreign key relationships between tables in the database
- **Cross-Database Support**: Works with PostgreSQL, MySQL, and SQLite databases
- **Comprehensive Information**: Displays source table, source column, target table, target column, and constraint name
- **Easy Navigation**: Access from tables view to see all database relationships at once

### Foreign Key Workflow
1. **Navigate to Tables**: Connect to a database and view the tables listing
2. **Access Relationships**: Press `f` to view all foreign key relationships in the database
3. **Review Relationships**: Browse through all foreign key constraints with their details
4. **Navigate Back**: Press `Esc` to return to the tables view

### Supported Databases
- **PostgreSQL**: Full foreign key support using information_schema queries
- **MySQL**: Foreign key relationships from INFORMATION_SCHEMA.KEY_COLUMN_USAGE
- **SQLite**: Foreign key information using PRAGMA foreign_key_list for each table

### Relationship Display Columns
- **From Table**: The table that contains the foreign key
- **From Column**: The column in the source table that references another table
- **To Table**: The table being referenced by the foreign key
- **To Column**: The column in the target table being referenced
- **Constraint Name**: The name of the foreign key constraint

## Connection Validation

DBX includes comprehensive connection validation to ensure reliable database connections:

### Features
- **Pre-save validation**: Connections are automatically tested before being saved
- **Enhanced error messages**: Clear, database-specific error descriptions
- **Connection timeout**: 10-second timeout prevents hanging on unreachable servers
- **Format validation**: Connection string format is validated for each database type

### Validation Process
1. **F1 - Test Connection**: Manually test a connection without saving
2. **F2 - Validate & Save**: Automatically tests connection, then saves if successful
3. **Real-time feedback**: Loading indicators and clear success/error messages
4. **Smart error handling**: Database-specific error messages with troubleshooting hints

### Enhanced Error Messages
- **PostgreSQL**: Server connection, authentication, database existence, and timeout errors
- **MySQL**: Server connection, access denied, unknown database, and timeout errors  
- **SQLite**: File existence, permissions, and database lock errors

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

## License

This project is licensed under the MIT License - see the [MIT License](https://opensource.org/licenses/MIT) for details.