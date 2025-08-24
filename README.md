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

- **Database utilities**: Schema detection, SQL generation, primary key finding, connection handling
- **Data processing**: Sorting parameters, table info creation, list management, result handling
- **UI utilities**: Smart table creation, intelligent column width calculation, enhanced data truncation
- **Mathematical utilities**: Min/max functions, pagination calculations
- **Type inference**: Content-aware data type detection (numeric, date, boolean, text), datetime parsing
- **Timeout utilities**: Async command helpers with timeouts and connection testing


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

- **hjkl/↑↓←→**: Navigate table and pages
- **enter**: Row details
- **/**: Filter data across all columns
- **s**: Sort mode - select column and cycle sort direction
- **r**: Reload table data
- **h/l**: Scroll columns horizontally when table is wider than screen
- Filter mode: **enter** apply filter, **esc** cancel
- Sort mode: **↑/↓** select column, **enter** cycle sort (off→asc→desc→off), **esc** exit
- **esc**: Back to tables

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

caldera/dbx/issues/new?template=feature_request.md) |
| 💬 **Questions** | [GitHub Discussions](https://github.com/dancaldera/dbx/discussions) |
| 📚 **Documentation** | [Wiki](https://github.com/dancaldera/dbx/wiki) |
| ✨ **Community** | [Discord Server](https://discord.gg/dbx) |

---

## 📦 Database Features

### 🏷️ Schema Support
- **PostgreSQL**: Full schema support with automatic detection and selection interface
- **MySQL**: Database-level organization (no schema selection needed)
- **SQLite**: Uses default `main` schema

### 📋 Tables & Views
- Enhanced data preview with smart column width distribution
- Improved sorting indicators directly in column headers (↑/↓)
- Intelligent data truncation for better readability
- Full column structure browsing for both tables and views

### 🔑 Indexes & Constraints
- Complete index information (primary keys, unique indexes, regular indexes)
- Constraint details (foreign keys, primary keys, check constraints)
- Visual organization with type, affected columns, and full definitions

### 🔗 Foreign Key Relationships
- Comprehensive relationship visualization across all tables
- Cross-database support for all relationship types
- Shows source/target tables and columns with constraint names

### 📤 Export Capabilities
- **CSV**: Comma-separated values with headers
- **JSON**: Array of objects format
- Automatic timestamped filenames
- Export from query results or table previews

---

## 🛠️ Development

### 🚀 Quick Development Setup

```bash
# Clone and setup
git clone https://github.com/dancaldera/dbx.git
cd dbx
go mod tidy

# Run in development mode
go run main.go

# Build for production
go build -o dbx main.go

# Run tests
go test ./...
```

### 📋 Development Commands

| Command | Purpose | Notes |
|---------|---------|-------|
| `go mod tidy` | Install dependencies | Run after cloning |
| `go fmt ./...` | Format code | Required before commit |
| `go vet ./...` | Static analysis | Catches common errors |
| `go test ./...` | Run tests | All tests must pass |
| `go build` | Build binary | Creates `dbx` executable |
| `go run main.go` | Development run | Hot reload for changes |

### 📝 Code Quality Standards

- **File size limit**: All source files must be under 500 lines
- **Separation of concerns**: Clear module boundaries and responsibilities
- **Testing**: Tests alongside code as `*_test.go` files
- **Formatting**: Follow standard Go formatting (`go fmt`)
- **Documentation**: Public functions must have Go doc comments

### 🎨 Architecture Patterns

- **TEA Pattern**: Model-Update-View architecture via Bubble Tea
- **Dependency Injection**: Components wired through initialization
- **Package Organization**: Domain-driven internal package structure
- **Error Handling**: Explicit error handling with user-friendly messages

---

## 📦 Dependencies

DBX is built on these excellent open-source libraries:

### 🎨 UI Framework
- [🫧 Bubble Tea](https://github.com/charmbracelet/bubbletea) `v1.3.6` - TUI framework
- [🥐 Bubbles](https://github.com/charmbracelet/bubbles) `v0.21.0` - TUI components
- [💄 Lipgloss](https://github.com/charmbracelet/lipgloss) `v1.1.0` - Styling and layout

### 🗄️ Database Drivers
- [🐘 PostgreSQL](https://github.com/lib/pq) `v1.10.9` - Pure Go Postgres driver
- [🐬 MySQL](https://github.com/go-sql-driver/mysql) `v1.9.3` - MySQL driver
- [📁 SQLite](https://github.com/mattn/go-sqlite3) `v1.14.28` - SQLite3 driver

### 🚀 Go Requirements
- **Go Version**: 1.24.5 or later
- **CGO**: Required for SQLite support
- **Build Tags**: None required

---

## 🤝 Contributing

We welcome contributions! Here's how to get started:

### 🚀 Quick Contribution Guide

1. **Fork & Clone**
   ```bash
   git clone https://github.com/your-username/dbx.git
   cd dbx
   ```

2. **Create Feature Branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```

3. **Make Changes**
   - Follow our [coding standards](#-code-quality-standards)
   - Add tests for new functionality
   - Update documentation as needed

4. **Test & Validate**
   ```bash
   go test ./...
   go fmt ./...
   go vet ./...
   ```

5. **Submit Pull Request**
   - Use our [PR template](https://github.com/dancaldera/dbx/blob/main/.github/pull_request_template.md)
   - Link related issues
   - Add screenshots for UI changes

### 📝 Contribution Areas

| Area | Examples | Difficulty |
|------|----------|------------|
| 🐛 **Bug Fixes** | Connection issues, UI glitches | 🟢 Beginner |
| 🎨 **UI/UX** | New themes, better layouts | 🟡 Intermediate |
| 🗄️ **Database Support** | New drivers, query features | 🔴 Advanced |
| 📚 **Documentation** | Guides, examples, translations | 🟢 Beginner |
| ⚡ **Performance** | Query optimization, caching | 🔴 Advanced |

See our [Contributing Guide](CONTRIBUTING.md) for detailed information.

---

## 📄 License

**MIT License** © 2025 Daniel Caldera

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software.

See [`LICENSE`](LICENSE) file for full details.

---

<div align="center">

## ⭐ Star History

[![Star History Chart](https://api.star-history.com/svg?repos=dancaldera/dbx&type=Date)](https://star-history.com/#dancaldera/dbx&Date)

### 🚀 Built with ❤️ by developers, for developers

**[Website](https://dbx.dev)** • **[Documentation](https://docs.dbx.dev)** • **[Discord](https://discord.gg/dbx)** • **[Twitter](https://twitter.com/dbx_dev)**

*Made with ☕ and Go in 2025*

</div>
