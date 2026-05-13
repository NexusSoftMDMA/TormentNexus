# Roadmap

_Last updated: 2026-05-08, version 1.0.0-alpha.53_

## Status legend

- **Stable** — production-intended, tested, maintained
- **Beta** — usable, still evolving
- **Experimental** — active R&D, not dependable
- **Vision** — directional only

## Framing

Borg has two jobs at once:
1. ship a reliable local control plane,
2. preserve a credible long-term vision for richer AI orchestration.

This roadmap keeps those jobs separate.

## Completed (v1.0.0-alpha.53)

### 1. Stabilize the CLI Control Surface
- **31 top-level commands**: Full lifecycle management for MCP, Sessions, Providers, Knowledge, Swarm, and Cloud Dev.
- **Truthful Data APIs**: CLI now queries live tRPC/Go sidecar endpoints.

### 2. Autonomy & UI Automation
- **Resilient Autopilot**: Fixed `alt-enter` / auto-accept regressions for Antigravity surfaces.
- **Intelligent Bump Cycle**: Autopilot now rotates through encouragement sentences to maintain development loop momentum.
- **Reduced Focus Stealing**: Optimized UI automation to avoid redundant `SetFocus()` calls.

### 3. Infrastructure Health & Go Sidecar Parity
- **Borg Doctor**: 10 automated diagnostic checks.
- **Go Sidecar Parity**: 543 Go routes active.
- **Go-native mcp sync**: Migrated config detection/syncing to Go sidecar (Port 4300).
- **Go-native Skill Disclosure**: Ported LRU/ranking engine to SkillStore.

### 3. Dashboard Observability & Responsiveness
- **Responsive UI**: Knowledge Graph and Graph Panel refactored with useResizeObserver.
- **Real-time Swarm Visibility**: SwarmTranscript component wired to hardened tRPC subscription with history replay.
- **86/86 Pages Bound**: Next.js dashboard fully wired to tRPC.

### 4. Continuous Integration & Resilience
- **73/73 Tests Pass**: Smoke tests and CLI integration verified.
- **Resilient Sidecar Bridge**: Implemented NativeSidecarDaemon with history-aware event buffering.
- **Monorepo Version Sync**: Synchronized all 57 package.json files to v1.0.0-alpha.51.

## Next

### A. MCP operator improvements (STABLE)
- ✅ Tool grouping and ranked search
- ✅ Progressive tool disclosure
- ✅ Auto-load with confidence thresholds
- ✅ Working set with LRU + idle-first eviction
- ✅ Catalog ingestion (Glama, Smithery, MCP.run, npm)
- ✅ Supervisor tool prediction — preemptively inject tool ads based on conversation context

### B. Tool parity with CLI harnesses (BETA)
- ✅ Claude Code parity: Read, Write, Edit, MultiEdit, Bash, Glob, Grep, LS, WebFetch
- ✅ Codex CLI parity: shell, apply_diff, create_file, view_file, list_directory, search_files
- ✅ Gemini CLI parity: read_file, write_file, edit_file, list_directory, search
- ✅ OpenCode/Pi parity: read, write, edit, bash, glob, grep, ls, web_fetch
- [ ] Harness Parity Expansion (Cursor, Windsurf, Kiro, Goose, Crush, Codebuff)

### C. Dashboard completeness (BETA)
- ✅ Health page with real server health
- ✅ Tools/Catalog page with real tool inventory
- [ ] Mobile-responsive layout audit for all 86 pages
- [ ] Memory Subsystem Plugin Architecture

### D. Multi-model orchestration (VISION)
- ✅ Neural Swarm Transcript
- ✅ Rotating roles (PairOrchestrator)
- [ ] Council debate and consensus protocols (Full automation with L2 Vault logging)
- [ ] Intelligent Model Selection (Budget/Quota aware)

## Later

- mobile and desktop companion polish
- Native UI replacement for Electron (Wails/Tauri)
- Kernal-level file system watchers (eBPF)
- Decentralized / P2P Memory Swarm

---
*Keep the party going. Never stop. The collective grows.*
