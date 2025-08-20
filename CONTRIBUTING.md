# Contributing to dbx

Thanks for your interest in improving dbx! Contributions of all kinds are welcome — bug reports, docs, tests, and features.

## Getting Started
- Prerequisites: Go 1.24.5 or later.
- Clone and build:
  - `git clone https://github.com/danielcaldera/dbx`
  - `cd dbx`
  - `go mod tidy`
  - `go build -o dbx ./cmd/dbx`
- Run: `go run ./cmd/dbx`
- Tests: `go test ./...`

## Development Guidelines
- Style: use `go fmt ./...`; keep imports organized.
- Static checks: run `go vet ./...` before committing.
- File size: keep source files under 500 lines.
- Naming: follow package and file conventions used in this repo.
- Architecture: keep UI in `ui`/`views`, DB in `database`, shared types in `models`.
- Errors: wrap with context, e.g., `fmt.Errorf("op failed: %w", err)`.

## Testing
- Prefer table-driven tests with the standard `testing` package.
- Keep tests deterministic; use fixtures over live DB connections.
- Place tests next to code as `*_test.go`.
- Run `go test ./...` and scope with `-run` when needed.

## Commits & PRs
- Branch: create a feature/fix branch from `main`.
- Commits: concise, imperative subject (≤72 chars), e.g.,
  - `Implement PostgreSQL schema selection`
  - `Fix input focus in connection view`
- PRs: include clear description, linked issues, test notes/steps, and screenshots for UI changes.
- Scope: keep PRs focused; update `README.md` when behavior or commands change.

## Issue Reports
- Describe the problem, steps to reproduce, expected behavior, and environment.
- Attach logs or screenshots where helpful.

## Code of Conduct
Please be respectful and constructive in all interactions. We foster a welcoming, inclusive environment for contributors of all backgrounds and experience levels.

## License
By contributing, you agree that your contributions are licensed under the MIT License. See `LICENSE`.
