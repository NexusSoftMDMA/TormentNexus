# Changelog

## [1.0.0-alpha.62] - 2026-05-19

### Added
- Protocol Scaffolding: Implemented the basic `hypercode://` handler in the Go kernel to support session attachment.
- Next.js Dashboard Routes: Added dashboard routes for Blocks, Claude Chrome, Claude Cloud, Copilot, and OpenAI Codex.
- L2 Vault Visualization: Wired the `vaultRecords` query to the Next.js frontend to show persistent heal history on the healer dashboard.

### Changed
- Standardized documentation identity to Nexus Kernel & HyperCode.
- Replaced git merge conflict markers across multiple internal Kotlin and Markdown files with unified content logic.

## [1.0.0-alpha.61] - 2026-05-17

### Added
- **Autonomous Healer Loop (The Immune System)**:
  - New `HealerService` in the Go kernel with a multi-turn `diagnose -> fix -> verify -> retry` loop.
  - Integration with `CodeExecutor` for native, sandboxed verification (tsc, vitest, go test).
  - L2 Vault persistence: All healing events and extracted facts are saved as long-term memory for fleet-wide intelligence sharing.

### Changed
- Standardized documentation identity to Nexus Kernel & HyperCode.
- Updated `VERSION.md`, `ROADMAP.md`, and `TODO.md` to reflect Phase 5 active sprint goals.
- Unified `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` as the single source of truth for all AI agents.
- Resolved merge conflict markers and aligned role guidelines across `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `GPT.md`, and `copilot-instructions.md`.

## [1.0.0-alpha.60] - 2026-05-16

### Added
- Fully integrated Go-native `MemoryManager` into the core TS control plane.
- Wires up `sqlite-vec` storage backend, replacing the deprecated `@borg/claude-mem` implementation.
- Dual-tier cache invalidation for the L1/L2 memory boundaries.

### Changed
- Shifted authority of MCP configuration sync entirely to the Go sidecar.
- Removed legacy TS synchronization scripts for VSCode and Cursor.
