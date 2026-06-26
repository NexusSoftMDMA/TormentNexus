# Contributing to Arc Relay

## Prerequisites

- **Go 1.24+**
- **Docker** (for managing stdio/HTTP MCP servers)
- **GCC** and **SQLite dev headers** (CGO is required for `go-sqlite3`)
  - Debian/Ubuntu: `apt install gcc libsqlite3-dev`
  - macOS: `xcode-select --install` (SQLite is bundled)
  - Alpine: `apk add gcc musl-dev sqlite-dev`

## Building and Running

```bash
# Build both binaries
make build-all

# Run with the example config
make run

# Or build and run manually
CGO_ENABLED=1 go build -o arc-relay ./cmd/arc-relay
./arc-relay --config config.example.toml

# Build the CLI (pure Go, no CGO)
CGO_ENABLED=0 go build -o arc-sync ./cmd/arc-sync
```

## Running Tests

```bash
make test          # Full test suite (requires CGO)
make test-cli      # CLI tests only (no CGO)
```

## Linting

```bash
make lint          # go vet
```

## Code Style

- Run `gofmt` on all Go files (most editors do this automatically)
- Follow standard Go conventions: https://go.dev/doc/effective_go
- Keep error handling explicit - no swallowed errors in new code
- Use stdlib `testing` package - no external test frameworks

## Pull Requests

1. Fork the repository and create a feature branch from `main`
2. Make your changes and ensure `make test` and `make lint` pass
3. Write a clear PR description explaining what changed and why
4. Keep PRs focused - one feature or fix per PR

## AI Contributors

If you're using an AI coding tool (Claude Code, Cursor, Copilot), see [AGENTS.md](AGENTS.md) for project structure and key abstractions.

## Reporting Issues

- Search existing issues before opening a new one
- Include steps to reproduce, expected behavior, and actual behavior
- For bugs, include your Go version (`go version`), OS, and Arc Relay version
