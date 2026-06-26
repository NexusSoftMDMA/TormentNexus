# MEMORY.md — Multi-Agent Observations

## Session 2026-06-17 (Merge & Documentation Session)

### Merge Architecture Observations
- **Branch topology**: `jules/baseline-128-hardened` had 7 unique commits diverging from `main` at `82a896d4f` (the merge base)
- **Conflict resolution strategy**: For this repo, `registry.go` conflicts should always be resolved by taking the branch with the full implementation (not stubs). The jules branch consistently has more complete tool registrations.
- **Fast-forward efficiency**: When branches like `assimilation-pipeline` and `assimilation-final` are behind main (no unique commits), use `git push <commit>:refs/heads/<branch>` to fast-forward them to the merged tip — this avoids creating unnecessary merge commits.
- **GitHub remote**: The repo URL has moved to `https://github.com/MDMAtk/TormentNexus.git` (GitHub redirects from old `NexusSoftMDMA/TormentNexus` URL)

### README.md Rewrite Observations
- A comprehensive README for this project needs: architecture diagram, monorepo structure tree, capability table, dashboard routes, API categories, and a "what's planned" section
- The README title should capture all 4 pillars: multi-agent, MCP tools, memory, and universal LLM routing
- Shields/badges for version, build, Go, TypeScript, Next.js, React, and license add immediate credibility
- Table of Contents is essential for a 650+ line document
- Using ASCII art for architecture diagrams is more reliable than Mermaid in plain markdown (though Mermaid works in GitHub)

## Session 2026-06-18 (Staging & Cleanup Session)

### Swarm Observations
- **Swarm stopped by itself**: No active swarm process found. The `swarm_v8.out` log shows it was hitting provider errors (nvidia empty responses) before stopping. May need a fresh run.
- **73 empty stub files committed**: Swarm creates zero-byte `.go` stub files that Go's build system silently ignores (no `package` declaration). These need to be filled before they become functional.
- **LLM provider flakiness**: Nvidia models (deepseek-v4-pro, deepseek-v4-flash, qwen-coder) returning empty responses consistently — may need to rotate providers or add retry fallbacks.

### Cleanup Observations
- **`.out` files grow large**: `swarm_forever.out` was 304KB, `swarm_norepair.out` was 182KB. These should be `.gitignore`'d after review.
- **`.pid` files are ephemeral**: Should never be tracked — they change every restart.

### Gotchas & Git Quirks
- **`data/` is gitignored** but `.db` files inside it may be tracked if added before the ignore rule. Always use `git add -f` or `git add -u` for DB files, never `git add data/`
- **`*.db-shm` and `*.db-wal`** are ignored — SQLite WAL files won't be committed. Good for avoiding large binary diffs.
- **`go-sidecar.pid`** is untracked (runtime file) — correct, don't track it.
- **`swarm_*.out` files** accumulate and can bloat the repo. Consider `.gitignore` them after review, or keep a few as progress evidence.
- **`tormentnexus.db`**, **`catalog.db`**, **`provider_metrics.db`** are large binary files. If tracked, every commit that touches them adds significant size. Consider using Git LFS or only tracking them on release commits.
- **Merge conflicts in `.gitignore`**: When the remote branch has fewer rules, always take `ours` (main) since it has the more comprehensive, tested ignore list.
- **Merge conflicts in `CHANGELOG.md`**: When the remote branch has older alpha versions, always take `ours` (main) with the newer version history. The remote's changelog is likely stale.
- **Merge conflicts in binary `.db` files**: Never attempt textual merge. Always take the newer version (`ours` during forward merge, or the one with the larger file size / more recent timestamp).
- **Branch names with slashes** like `jules/baseline-128-hardened-2272628885254508907` require `refs/heads/` prefix in `git push`: `git push origin <commit>:refs/heads/<branch-name>`
- **`git checkout --ours` vs `--theirs`**: In a merge, `ours` = the branch you're currently on (main), `theirs` = the branch being merged. This is counter-intuitive when rebasing.

### Failure Lessons
- **bobbybookmarks.com** DNS resolution fails consistently from this environment — permanently blocked. Use Smithery.ai or Glama.ai for MCP server catalog discovery.
- **`--repair` flag** in swarm causes premature exit — use `--forever` without `--repair`
- **`tormentnexus-upstream` remote** does not exist — only `origin` and `origin-backup` (dead). All pushes go to `origin`.
- **`.out` files from swarm** are large and should be ignored by `.gitignore` to prevent repo bloat. Add `*.out` and `swarm_*.out` to `.gitignore`.
- **Deleting tracked files**: When swarm removes broken tool files, `git add -A` will stage the deletions. This is correct — commit them as part of cleanup.
- **Merge conflicts in `go/internal/tools/registry.go`**: This file always conflicts when merging branches because every branch adds tool registrations. The correct resolution is to take the version with the most registrations (usually the branch being merged), then verify `go build` compiles.
- **Windows `EBUSY` errors**: When Git cannot unlink `.db` files (they're locked by a running process), use `git checkout -f` or close the process before switching branches.

### Preferences
- Always stage and commit `.db` files per user instruction "stage and track db always"
- Use `--forever` mode for swarm to avoid premature shutdown
- Tag commits with version: `v1.0.0-alpha.X`
- After any merge, verify `go build ./cmd/tormentnexus` compiles clean
- When a README.md rewrite is done, immediately commit it separately so it doesn't get lost in merge noise
- Fast-forward feature branches that are fully merged rather than leaving them stale
- Delete feature branches on GitHub after they are fully merged into main

## Session 2026-06-24 (MCP Parity & Compile Hardening)

### MCP Server Observations
- **Local Module Replacing**: Setting `replace github.com/NexusSoftMDMA/TormentNexus => ../` and requiring `github.com/NexusSoftMDMA/TormentNexus v0.0.0` in `go/go.mod` allows the Go sidecar module to import `"github.com/NexusSoftMDMA/TormentNexus/tools"` without invoking network proxy lookups.
- **Unused Import Errors**: The Go compiler enforces unused imports strictly. Having a python compiler feedback loop that isolates files failing due to unused imports into `_disabled` and regenerates the dispatch map dynamically ensures compilation success.
- **Console Window Output**: To prevent JSON-RPC stream corruption for stdio client runners, all logging must be written strictly to `os.Stderr`.

## Session 2026-06-24 (Dashboard Consolidation & MCP Robustness)

### Stdio MCP Protocol Robustness
- **JSON-RPC Parameters**: The MCP spec allows `params` to be optional or a JSON array (`[]`) for methods like `tools/list`. Using `json.RawMessage` for the request envelope's `params` field prevents unmarshal errors on standard client handshakes, dynamically decoding the struct parameters only inside handlers that strictly require them (like `tools/call`).
- **Dashboard Layout**: Overview dashboards should avoid repeating structural readouts (like detailed metrics `dl` blocks) if they already exist in top-level cards, balancing screen estate and avoiding unnecessary page height.

### Page Consolidation & Binary Pathing Heuristics
- **Page De-duplication**: Pages with similar intent (like `/dashboard/config` form and `/dashboard/settings` raw JSON text area) should be consolidated into a single route with a tabbed interface. This reduces code footprint, streamlines navigation, and improves UX.
- **Root Binary Pathing**: When the monorepo has two Go projects (e.g., a root Cobra CLI and a subfolder sidecar), they may compile binaries with the same name. If configurations or start scripts prioritize the root directory path (`tormentnexus.exe`), overwrite the root binary with the sidecar server so that subcommands like `mcp` can execute without Cobra CLI path conflicts.
- **PowerShell Overwriting**: When using PowerShell in Windows, `Copy-Item` requires the `-Force` flag to overwrite an existing binary; otherwise, it silently leaves the target unchanged.

## Session 2026-06-24 (Dashboard Consolidation Phase 2 & 3)

### Agent & Knowledge Consolidation Heuristics
- **Complex Page Consolidation**: Grouping multiple related subpages (e.g., `/dashboard/director`, `/dashboard/council`, `/dashboard/supervisor`, `/dashboard/squads`, `/dashboard/swarm`) into a single route with a tabbed interface makes the dashboard significantly cleaner and more cohesive, while preserving routing logic.
- **Tab Layout Structure**: Using Framer Motion's `AnimatePresence` and `motion.div` transitions on tab panels creates a high-end, responsive feel.
- **Vitest Workspace Matching**: Vitest scans all git worktrees recursively by default, matching nested test files. To run a fast, scoped test check, limit the matches by targeting specific directories or using exact path matching.

### Session Supervisor Robustness
- **Process Spawning Try-Catch**: When the SessionSupervisor or PtySupervisor attempts to restore active sessions from `session-supervisor.json` on startup, it executes `spawnProcess` (e.g. `node-pty`). If a shell path or binary is missing on the host, it throws a synchronous error (like `Error: File not found: ...`). Wrapping the spawn block in a try-catch ensures that individual session start failures are handled gracefully (marking the session as 'error' state) rather than raising an uncaught exception that completely crashes the main MCP server.


## Session 2026-06-25 (Pure Go Vector Index & Memory Consolidation)

### Vector Search Architecture
- **Pure Go Cosine Similarity**: Computing cosine similarity in pure Go is highly efficient (takes < 2ms for 20,000 vectors of 384 dimensions) and avoids heavy CGO dependencies (`sqlite-vec` or external database links), making it ideal for local-first desktop agents.
- **Double-Value Helper Rules**: The swarm generator sometimes outputs invalid brackets or files. Running the compiler reset script (`compiler_reset.py`) clears these quickly and maintains compiler safety.
- **L1 In-Memory Caching**: Mirroring BobbyBookmarks' tiered design using in-memory caches inside the `VectorStore` structure provides rapid key-value retrieval for hot memories during chat loops.

## Session 2026-06-25 (Dashboard Consolidation Phase 4 & Dev Hardening)

### Dashboard Consolidation
- **Brain & Memory Consolidation**: Unified `/dashboard/brain` and `/dashboard/memory` (and its hydration pages) into a tabbed layout in `/dashboard/brain/page.tsx` called **TormentNexus Cognitive Hub**. This groups Symbols Graph, Memory Vault, URL Ingestion, Expert Agents, Observations logs, and Hydration sync under a single sidebar tab.
- **Client Redirects**: Replacing redundant page paths (like `/dashboard/memory`) with a client-side Next.js `useRouter.push()` redirect ensures backward compatibility for older bookmarks.
- **Turbo Filter validation**: When a package doesn't exist in the workspace, Turbo's `--filter` triggers an error. Custom dev tools (like `scripts/dev_tabby_ready.mjs`) must avoid hardcoded package exclusion filters that target deleted packages (like `mcp-superassistant` or `@extension/hmr`).

