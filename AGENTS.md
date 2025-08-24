# Repository Guidelines

## Project Structure & Modules
- `internal/`: application packages
  - `config/` (storage, persistence)
  - `database/` (DB queries and adapters)
  - `models/` (core types and state)
  - `styles/` (theme/styling)
  - `ui/` (Bubble Tea models initialization)
  - `views/` (UI view rendering)
- Root `main.go`: main entry point with app logic and update handlers.
- Tests: alongside code as `*_test.go`.

## Project Overview
- App: DBX — a terminal-based database explorer (Go TUI).
- Databases: PostgreSQL, MySQL, SQLite (tables and views supported).
- State-driven UI: selection, connection, schema, tables, columns, query, history.

## Architecture
```
dbx/
├── main.go                     # Main application entry point with update logic
├── internal/
│   ├── config/                 # Configuration and file storage
│   ├── database/               # Database operations and adapters
│   ├── models/                 # Core types and interfaces
│   ├── styles/                 # UI theming
│   ├── ui/                     # Bubble Tea model initialization
│   └── views/                  # UI view rendering
```
- Core states: `dbTypeView`, `connectionView`, `schemaView`, `tablesView`, `columnsView`, `queryView`, `queryHistoryView`.
- Package roles: `config` (persistence), `database` (queries), `models` (types), `styles` (theme), `ui` (init), `views` (rendering).
- Update logic: implemented in `main.go` via `appModel` wrapper pattern (Go best practice for extending models from other packages).
- Key deps: `bubbletea`, `bubbles`, `lipgloss`; DB drivers: `lib/pq`, `go-sql-driver/mysql`, `mattn/go-sqlite3`.

## Build, Test, and Development
- Install deps: `go mod tidy`.
- Build binary: `go build`.
- Run locally: `go run .` or `./dbx`.
- Format code: `go fmt ./...`; keep imports organized.
- Static checks: `go vet ./...`.
- Run tests: `go test ./...` (use `-run` to scope).

### Additional Dev Commands
- Update deps: `go mod download`.

## Application Flow
1. Choose DB type or select a saved connection.
2. Enter connection string (per driver) if new.
3. Connect, load schemas (PostgreSQL), then tables and views.
4. Navigate to columns, run queries, and browse history.

## Connection & Schema Support
- Connections file: `~/.dbx/connections.json`.
- Save/load: press `s` in connection/main views.
- PostgreSQL schemas: auto-detected; prompt when multiple exist; operations scoped to selection.

## View Support
- Views and tables are listed with clear labels; supported across PostgreSQL, MySQL, SQLite.

## Coding Style & Naming Conventions
- Language: Go; follow standard tooling output (`go fmt`).
- Packages: short, lowercase names (no underscores or camelCase).
- Files: `snake_case.go` by feature (e.g., `operations.go`, `storage.go`).
- Exports: `CamelCase`; unexported: `lowerCamel`.
- Errors: wrap with context (e.g., `fmt.Errorf("op failed: %w", err)`).
- Separation: UI in `ui`, DB in `database`, shared types in `models`.
- File size: keep each source file under 500 lines; refactor when approaching.

## Navigation Controls
- Arrows: navigate lists/tables; `Enter`: select; `Esc`: back.
- `s`: saved connections (menu) / save current (connection view); `n`: new connection.
- `p`: preview table data; `r`: run SQL; `Ctrl+H`: query history.
- `i`: indexes/constraints; `f`: foreign keys; `/` or `Ctrl+F`: search.
- `F1`: test connection; `F2`: validate, save, connect; `Ctrl+E`: export CSV; `Ctrl+J`: export JSON.
- `q` or `Ctrl+C`: quit.

## Testing Guidelines
- Framework: standard `testing` package; table-driven tests preferred.
- Determinism: use fixtures over live DB connections.
- Location: tests live next to code as `*_test.go`.
- Run: `go test ./...`; add focused runs with `-run`.

## Database Connection Examples
- PostgreSQL: `postgres://user:password@localhost/dbname?sslmode=disable`
- MySQL: `user:password@tcp(localhost:3306)/dbname`
- SQLite: `/path/to/database.db`

## Commit & Pull Request Guidelines
- Commits: imperative mood, concise subject (≤72 chars).
  - Examples: `Implement PostgreSQL schema selection`, `Fix input focus in connection view`.
- PRs: clear description, linked issues, test notes/steps, screenshots for UI changes.
- Scope: keep focused; update `README.md` when behavior or commands change.

## Known Issues
- None currently; app builds and runs successfully.

## Security & Configuration Tips
- Do not commit credentials or real connection strings; use placeholders.
- Avoid committing build artifacts (e.g., `dbx`); ensure `.gitignore` covers them.
- Prefer environment variables or local config for sensitive values during development.

## Code Quality Standards
- File size limit: keep all source files under 500 lines; refactor when approaching to maintain readability, testability, and clear separation of concerns.
