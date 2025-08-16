# Repository Guidelines

## Project Structure & Modules
- `cmd/dbx`: CLI entrypoint (`main.go`). Prefer this for new work.
- `internal/`: application packages
  - `config/` (storage, persistence)
  - `database/` (DB queries and adapters)
  - `handlers/` (update logic)
  - `models/` (core types and state)
  - `styles/` (theme/styling)
  - `ui/` (Bubble Tea models, views)
- Root `main.go`: legacy entry; keep changes minimal here.

## Build, Test, and Development
- Install deps: `go mod tidy`
- Build binary: `go build -o dbx ./cmd/dbx` (or `go build .` if using root)
- Run locally: `go run ./cmd/dbx` or `./dbx`
- Format: `go fmt ./...`
- Vet: `go vet ./...`
- Tests: `go test ./...` (table-driven tests preferred)

## Coding Style & Naming Conventions
- Formatting: standard Go tooling (`go fmt`); keep imports organized.
- Packages: short, all-lowercase names (no underscores or camelCase).
- Files: `snake_case.go` by feature (e.g., `operations.go`, `storage.go`).
- Exports: `CamelCase` for exported, `lowerCamel` for unexported.
- Errors: wrap with context (`fmt.Errorf("op failed: %w", err)`); return early.
- Boundaries: UI in `ui`, DB logic in `database`, shared types in `models`.

## Testing Guidelines
- Framework: standard `testing` package.
- Location: alongside code as `*_test.go` (e.g., `internal/database/operations_test.go`).
- Style: table-driven tests; prefer deterministic fixtures over live DB.
- Run: `go test ./...` (consider `-run` to scope while iterating).

## Commit & Pull Request Guidelines
- Commits: imperative mood, concise subject (≤72 chars), details in body if needed.
  - Examples: `Implement PostgreSQL schema selection`, `Fix input focus in connection view`.
- PRs: clear description, linked issues, test notes/steps, and screenshots for UI changes.
- Scope: keep PRs focused; update `README.md` when behavior or commands change.

## Security & Configuration Tips
- Do not commit credentials or real connection strings; use placeholders.
- Avoid committing build artifacts (e.g., `dbx`). Add to `.gitignore` if needed.
- Prefer environment variables or local config for sensitive values during development.

