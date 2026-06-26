# Arc Relay - AI Assistant Context

Arc Relay is an open-source MCP (Model Context Protocol) control plane written in Go. It provides auth, policy controls, traffic interception, and archiving for AI tool calls.

## Project Structure

- `cmd/arc-relay/` - Server binary (CGO required for SQLite)
- `cmd/arc-sync/` - CLI tool for syncing MCP servers to projects (pure Go)
- `internal/middleware/` - Bidirectional middleware pipeline (archive, sanitizer, sizer, alerter)
- `internal/proxy/` - MCP proxy backends (stdio bridge, HTTP, remote with OAuth)
- `internal/store/` - SQLite persistence layer with migrations
- `internal/web/` - HTTP handlers and server-rendered HTML templates
- `internal/config/` - TOML config + ARC_RELAY_* env var overrides
- `internal/cli/` - CLI packages (config, sync, project detection)

## Building

```bash
CGO_ENABLED=1 go build ./cmd/arc-relay    # Server (needs gcc, libsqlite3-dev)
CGO_ENABLED=0 go build ./cmd/arc-sync     # CLI (pure Go, cross-platform)
go test ./...                              # Run all tests
go vet ./...                               # Static analysis
```

## Key Patterns

- Middleware implements `ProcessRequest` and `ProcessResponse` on `mcp.Request`/`mcp.Response`
- Request middleware runs in order; response middleware runs in reverse order
- Config: TOML file + env var overrides (ARC_RELAY_*, ARC_SYNC_*)
- Auth: session cookies (web UI) + Bearer API keys (proxy)
- Archive dispatcher uses SQLite queue with retry backoff and circuit breaker
