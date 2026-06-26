# TormentNexus — Project Roadmap

## Project Purpose

TormentNexus is a Cognitive Kernel / Universal AI Control Plane for multi-agent workflows, MCP tools, and context-aware memory. It assimilates MCP servers from public catalogs, generates Go-native implementations, and provides a unified API surface via a Go sidecar + Next.js dashboard.

## Current State (June 22, 2026)

### Running Services (all 7 ports active)
- **freellm proxy**: port 4000 — LLM routing for all workers
- **Go sidecar** (new): port 7778 — Fiber HTTP API server
- **Dashboard** (Next.js): port 7779 — production mode
- **TS control plane**: port 4100 — Node.js tRPC bridge
- **Go sidecar** (legacy): port 4300 — old module
- **LM Studio**: port 1234 — local AI models
- **Go sidecar** (old zombie): port 8080 — can't kill (system process)

### Workers (all running silently via pythonw)
- **Swarm v7**: 5 workers, freellm-only code generation (GEN → REVIEW → FIX pipeline)
- **Watchdog**: monitors all services, restarts dead workers
- **BobbyBookmarks sync**: hourly bookmark merge (10,820 bookmarks imported)
- **Trends analyzer**: 6-hour analysis cycle (uses LM Studio)

### Data
- **12,201 MCP servers** tracked (1,932 implemented, 10,269 pending)
- **40,262 rows** imported from bobbybookmarks (bookmarks, atlas, embeddings, etc.)
- **7,064 Go tool files** (3,815 with real handler implementations)
- **545 bookmarks** imported from bookmarks_remaining.txt

### Services Registered (auto-start on boot)
- TormentNexusSidecar
- TormentNexusDashboard
- TormentNexusWatchdog

## Key Decisions Made

1. **freellm-only for swarm**: Generation, review, and fix all use `free-llm` model with 3600s timeout. LM Studio removed as fallback to prevent curl pileup.
2. **LM Studio for trends**: Trends analyzer uses local LM Studio (port 1234) instead of freellm — lighter workload.
3. **No more free-llm-fallback**: Removed because it returned empty responses 90% of the time.
4. **pythonw for silent workers**: All Python workers run with `pythonw.exe` + `CREATE_NO_WINDOW` to prevent console popups.
5. **Git LFS narrowed**: Only specific large DB files (`provider_metrics.db`, `tormentnexus.db`, `catalog.db`) are LFS-tracked — small DBs stay under regular git.
6. **Ports moved**: Orchestrator 8080→7778, Dashboard 3000→7779, tRPC 4100→7779.

## Completed Milestones

- [x] Git LFS fixed from `*.db` wildcard to specific files
- [x] All subprocess calls hidden (curl, git, go, taskkill, wmic — all silent)
- [x] Duplicate watchdog spawning fixed (pythonw detection + kill duplicates)
- [x] Core services registered as Windows scheduled tasks
- [x] Full data import from ../bobbybookmarks (40k+ rows)
- [x] Assimilation DB rebuilt with Go file ↔ server name mapping
- [x] Dashboard production build + dev mode working on 7779

## Planned Work (TODO — 13 items)

1. **ChunkHound / Probe Integration** — implement remaining MCP search tools as native Go handlers
2. **Session Import** — format works (228 sessions parsed), needs orchestrator POST endpoint for actual restoration
3. **Catalog DB Sync** — index new skills into `catalog.db` for unified search
4. **Git LFS** — clean up remaining LFS tracking issues
5. **Submodule Removal** — systematic removal of redundant submodules
6. **Skill Evolution** — win-rate tracking and auto-retirement for 3,000+ skills
7. **P2P Memory** — gossip protocol for decentralized context sharing
8. **L3 Cold Archive** — long-term compressed memory tier
9. **Fleet-Wide Intelligence** — cross-machine memory sharing
10. **Wails Native Runtime** — Go-native desktop shell
11. **Deep Link Protocol** — `tormentnexus://` protocol
12. **Compliance Boundary** — SSO/RBAC/Audit enterprise wrapper
13. **New Native Tools** — browser-use and browsermcp specialized logic

## Key Blockers

- **Model quality**: The `free-llm` model produces Go code that fails review 90% of the time. Adding better models to freellm is the single biggest improvement.
- **No DONE tasks**: 1,932 implemented (from file matching), but 0 full pipeline completions from the swarm since model downgrade.
- **Port 8080 zombie**: freellm.exe holds port 8080 as a system process — can't kill without admin rights.
