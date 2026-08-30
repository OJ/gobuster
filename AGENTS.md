# Repository Guidelines

## Project Structure & Module Organization

`main.go` is the executable entry point. Command-line wiring and mode-specific argument handling live under `cli/` (for example, `cli/dir/` and `cli/dns/`). Shared scanning infrastructure is in `libgobuster/`; implementations are split into top-level packages such as `gobusterdir/`, `gobusterdns/`, `gobusterfuzz/`, `gobustervhost/`, `gobusters3/`, `gobustergcs/`, and `gobustertftp/`. Keep mode options, results, and execution logic within the corresponding package. Tests sit beside production code as `*_test.go`. Demo recordings and their source scripts are under `vhs/`; CI definitions are in `.github/workflows/`.

## Build, Test, and Development Commands

- `task build` runs dependency cleanup, formatting, vetting, and then builds a static Linux AMD64 `gobuster` binary.
- `task test` runs `go test -race -cover ./...` after dependency and formatting checks.
- `task check` applies `go fmt`, applies `gofumpt`, and runs `go vet ./...`.
- `task lint` runs `golangci-lint` across all packages and verifies that `go.mod`/`go.sum` remain tidy.
- `task windows` cross-compiles `gobuster.exe`; `task linux` builds the default Linux target.
- `go run . dir -u https://example.com -w wordlist.txt` runs a local development build.

Task commands may update module or formatting files. Review the resulting diff before committing.

## Coding Style & Naming Conventions

Follow idiomatic Go and let `gofumpt` determine formatting (tabs for Go indentation). Package names are short and lowercase. Exported identifiers use `PascalCase`; internal identifiers use `camelCase`. Match existing mode types such as `Options`, `Result`, and `GobusterDir`, and keep CLI concerns separate from scanning logic. Add context to returned errors and avoid unnecessary global state.

## Testing Guidelines

Use Go's standard `testing` package and name files `*_test.go` with functions such as `TestParseExtensions`. Prefer table-driven tests for option parsing and multiple edge cases. Run `task test` before opening a pull request; the race detector and coverage report are part of the normal suite. There is no fixed coverage threshold, but new behavior and regressions should receive focused tests.

## Commit & Pull Request Guidelines

History favors brief, imperative subjects (for example, `fix go vet` or `update dependencies`); use a clear scope when useful, such as `dns: handle wildcard timeout`. Keep commits focused and avoid mixing refactors with behavioral fixes. Pull requests should explain the problem and solution, list validation commands, link relevant issues, and call out user-visible CLI changes. Include terminal output or updated VHS assets when output formatting changes; screenshots are otherwise unnecessary.
