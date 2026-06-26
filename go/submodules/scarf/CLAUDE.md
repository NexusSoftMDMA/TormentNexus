# Scarf — macOS GUI for the Hermes AI Agent

## Build

`./scripts/build-detached.sh` (no args) — builds into isolated DerivedData and launches a decoupled,
visually-distinct **Scarf Dev** copy; each run quits every running copy first so you get a single
instance, sparing only agent test copies under `/tmp`.

Raw compile-only:

```bash
xcodebuild -project scarf/scarf.xcodeproj -scheme scarf -configuration Debug build
```

<!-- memophant:begin -->
<!-- memophant:shim -->
> Agent instructions for this project live in [AGENTS.md](./AGENTS.md) — read it before
> starting work. Memophant manages a repo-resident memory system (`.memory/`, `wiki/`, `design/`,
> `code/`, `TASKS.md`) and a native MCP server (`memophant-mcp`) for read/write. Substance lives
> in those files, not here.
<!-- memophant:end -->
