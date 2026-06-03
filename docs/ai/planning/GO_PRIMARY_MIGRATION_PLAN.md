# Go-Primary Migration Plan

## Objective
Make the Go runtime the **primary TormentNexus control plane** and port **every backend feature/function** from the current TypeScript runtime into Go, leaving only the most crucial TypeScript pieces in place temporarily:
- UI clients
- compatibility bridges
- any TS-only surface that has not yet reached validated Go parity

This plan is intentionally **truthful**:
- Go is already viable for meaningful fallback and several native subsystems.
- Go is **not yet** at 100% parity.
- The target state is a **Go-authoritative backend**, with TypeScript reduced to compatibility and UI roles until fully retired where practical.

---

## Authoritative target state

### Go owns by default
The Go runtime should become authoritative for:
- startup / lifecycle / lock / control-plane port ownership
- MCP inventory, routing, runtime server mediation, and write/config flows
- provider routing and quota-aware model execution
- session import/export, persistence, deduplication, and memory extraction
- SQLite-backed persistence and maintenance flows
- background workers / ingestion daemons
- workflows, supervision, orchestration, and councils
- submodule sync / repo orchestration / automation endpoints
- SSE/event streaming used by extensions and native clients

### TypeScript remains temporarily for
- web dashboard UI
- desktop/electron UI
- any still-unported compatibility router or UI helper surface
- any specialist integration that is not yet ported and validated in Go

### End-state rule
No backend surface should remain TypeScript-owned once the Go equivalent is:
1. implemented
2. validated
3. wired into startup
4. truthful in operator-facing status/docs

---

## Current parity classification model
Each major surface should be tracked as one of:
- **Native Go** — implemented and authoritative in Go
- **Bridge-first** — Go route exists but still defers to TS first
- **Bridge-only** — Go is only a transport shim to TS
- **Missing** — no real Go surface yet
- **Retirable TS** — TS implementation can be demoted or removed once callers are switched

---

## Migration rules

### 1. No false parity claims
Do not mark a surface complete until:
- the Go implementation exists
- tests exist where practical
- the startup path can prefer Go for that surface
- docs/status reflect reality

### 2. Port stateful/fragile surfaces first
Prioritize anything currently blocked by:
- `better-sqlite3`
- Node-native addon fragility
- hosted quota failures in TS-only service paths
- startup-time background jobs that can abort or degrade operator confidence

### 3. Go-first startup wins over passive sidecar framing
The launcher should evolve from:
- `TS primary + Go fallback`

to:
- `Go primary + TS compatibility sidecar/supplement`

### 4. TypeScript must shrink, not grow
New backend logic should default to Go unless a TS implementation is strictly required to unblock UI/client behavior.

---

## Priority workstreams

## Workstream A — Startup and control-plane ownership
### Goal
Make Go the default runtime started by operator entrypoints.

### Tasks
- Add explicit runtime selection in CLI/startup (`go`, `node`, `auto`)
- Flip startup defaults to **prefer Go**
- Make lock files and port ownership Go-authoritative
- Run TS compatibility services only when required
- Keep occupied-port behavior non-destructive and truthful

### Exit criteria
- `start.bat` and `tormentnexus start` prefer Go by default
- default startup validation is Go-primary rather than full-TS-workspace-first
- Go owns the primary control-plane port
- TS startup becomes optional/compatibility-oriented

### Progress note
- `start.bat` now defaults to a Go-primary startup build profile for `auto`/`go` runtime modes: it validates the Go control plane and CLI without requiring a full workspace build first
- `start.bat` now also probes whether Go-primary startup dependencies are already present and can skip `pnpm install` when the current workspace is already ready
- `start.bat` now also probes whether the Go-primary startup build artifacts are already current and can skip the startup build step for repeat launches
- `start.bat` now prints explicit install/build phase summaries so operators can see whether each phase ran or was skipped and why
- `start.bat` now exports those install/build decisions to the CLI runtime so startup provenance can be persisted and queried later
- `start.bat` now launches through the built CLI entrypoint directly instead of relying on `pnpm start` for the final handoff path
- the CLI Go runtime launcher now prefers the prebuilt `go/tormentnexus(.exe)` binary and only falls back to `go run ./cmd/tormentnexus` when the binary is absent or source launch is explicitly forced
- startup output now reports whether Go is running via the prebuilt binary, via source fallback, or whether Node compatibility runtime is active due to explicit selection or Go fallback
- the CLI now also prints a concise startup mode summary block describing which surfaces are actually active/compatibility-only in the chosen runtime
- `tormentnexus status` now exposes persisted startup provenance from the local startup lock when available
- the TypeScript `startupStatus` API surface now also exposes that persisted startup provenance, making the same truth available to dashboard and API consumers
- the dashboard startup-readiness UI now renders a visible `Startup mode` section backed by that persisted/API-visible provenance
- the dashboard Health, Integrations, System, MCP System, and Orchestrator pages now also surface startup/runtime provenance so operators can see launch truth across the major runtime views
- the web local-compat startup fallback now also carries `startupMode` from the local lock, reducing dependence on a live TS startup snapshot for this runtime truth
- the web tRPC compat layer now prefers Go-native `/api/startup/status` and `/api/runtime/status` when the TypeScript `startupStatus` procedure is unavailable, so dashboard startup truth degrades to native Go state instead of only local lock/config guesses
- the web tRPC compat layer now also prefers Go-native `/api/mcp/status` when the TypeScript `mcp.getStatus` procedure is unavailable, improving dashboard MCP/system/router truth across both legacy MCP bridge batches and richer local dashboard fallback mode
- the web tRPC compat layer now also prefers Go-native `/api/billing/provider-quotas` and `/api/billing/fallback-chain` when the TypeScript billing procedures are unavailable, replacing empty provider/fallback placeholder data with native Go provider-routing previews in both legacy bridge batches and local dashboard fallback mode
- the web tRPC compat layer now also prefers Go-native `/api/cli/harnesses` for `tools.detectCliHarnesses` when the TypeScript harness-detection procedure is unavailable, replacing empty degraded-mode harness detections with native Go harness inventory in local dashboard fallback mode
- the web tRPC compat layer now also prefers Go-native `/api/sessions` for `session.list` when the TypeScript session-list procedure is unavailable, replacing empty degraded-mode session inventories with native Go-discovered session rows in both legacy bridge batches and local dashboard fallback mode
- the web tRPC compat layer now also derives `session.catalog` from Go-native `/api/cli/harnesses` when the TypeScript session-catalog procedure is unavailable, preserving session-creation harness metadata instead of collapsing the catalog to empty in degraded mode
- `start.bat` now captures startup probe exit codes with runtime `!ERRORLEVEL!` instead of parse-time `%ERRORLEVEL%`, eliminating contradictory Go-primary build messaging during startup and making the install/build phase summaries truthful again
- Go-primary startup installs now default to `pnpm install --ignore-scripts` unless `TORMENTNEXUS_STARTUP_INSTALL_SCRIPTS=1` is set, reducing startup dependence on unrelated workspace postinstall hooks when only the Go control plane + built CLI lane is needed
- the web tRPC compat layer now also prefers Go-native `/api/tools/detect-execution-environment` when the TypeScript execution-environment procedure is unavailable, normalizing native shell/tool/harness posture into the existing dashboard contract instead of returning an all-zero synthetic placeholder
- degraded `startupStatus.checks.executionEnvironment` now reuses that same normalized Go-native execution summary, keeping dashboard-home/system readiness summaries aligned with the AI Tools page during TypeScript outage/degraded mode
- the web tRPC compat layer now also prefers Go-native `/api/tools/detect-install-surfaces` when the TypeScript install-surface procedure is unavailable, preserving browser-extension / VS Code / MCP-sync artifact summaries instead of collapsing install-surface pages to `[]` in degraded mode
- the web tRPC compat layer now also prefers Go-native `/api/sessions/imported/maintenance-stats` when the TypeScript imported-maintenance procedure is unavailable, preserving imported-session archive/retention counters and backfilling degraded `startupStatus.checks.importedSessions` when startup telemetry omits that block
- the web tRPC compat layer now also prefers Go-native MCP inspector state for `/api/mcp/working-set`, `/api/mcp/tool-selection-telemetry`, and `/api/mcp/preferences` when the corresponding TypeScript procedures are unavailable, preserving working-set rows, tool-selection telemetry history, and tool-preference controls instead of falling back to synthetic empty placeholders
- the web tRPC compat layer now also prefers Go-native `/api/api-keys`, `/api/shell/history/system`, `/api/memory/agent-stats`, and `/api/expert/status` when the corresponding TypeScript procedures are unavailable, preserving operator-facing API-key metadata, shell-history lines, compact agent-memory stats, and expert offline status instead of synthetic placeholders
- the Go HTTP layer now owns persisted local fallback writes/searches for agent-memory-backed facts, observations, user prompts, session summaries, pivot search, timeline windows, cross-session links, and direct `/api/agent-memory/*` inventory/export/stats/handoff/pickup mutations through `.tormentnexus/agent_memory/memories.json` when the TypeScript runtime is unavailable
- the Go HTTP layer now also owns a more truthful degraded-memory context path: generic `memory.query` can merge local SQLite-backed memory rows with persisted `.tormentnexus/memory/contexts.json` entries, `memory.saveContext` can persist new local saved-context entries there during TypeScript outage, `memory.getContext` can return locally persisted inline context bodies when present, `memory.deleteContext` can remove local context-registry entries instead of hard-failing, and the local memory interchange cluster now truthfully supports structured format listing plus import/convert fallback across `json`, `json-provider`, `jsonl`, `csv`, and `sectioned-memory-store`
- the web tRPC compat layer now also prefers Go-native `/api/tools` and `/api/tools/search` when `tools.list` and `mcp.searchTools` are unavailable, preserving tool inventory and search results instead of synthetic empty catalog/search placeholders
- the web tRPC compat layer now also prefers Go-native `/api/mcp/traffic` and `/api/server-health/check` when `mcp.traffic` and UUID-backed `serverHealth.check` are unavailable, preserving router traffic rows and truthful server health counters instead of synthetic empty traffic and config-only health inference
- the web tRPC compat layer now also prefers Go-native `/api/sessions/supervisor/state` when `session.getState` is unavailable, preserving truthful session-state signals like active auto-drive and current goal instead of the synthetic session-state placeholder
- the Go-native `/api/runtime/status` surface now also exposes startup provenance, making the native backend self-describing rather than depending on the TS compatibility surface for that truth
- this dashboard propagation cluster is now complete; the next focus is reducing remaining TS compatibility dependence by switching more runtime-heavy dashboard/system reads onto Go-native truth where equivalent native surfaces already exist
- Go-primary startup no longer has to hard-skip the web dashboard: `tormentnexus start --runtime auto|go` can now launch the Next.js dashboard in a compatibility-backed mode against the live Go control plane, while still warning explicitly that some mutation-heavy surfaces remain compatibility-dependent during the migration
- the shared Next.js compat route now also maps the session dashboard's key supervisor reads/mutations onto Go `/api/sessions/supervisor/*` routes when `/trpc` is unavailable, making Go-primary dashboard startup materially more usable for supervised-session workflows instead of only launching the shell UI
- the shared Next.js compat route now also maps the memory dashboard’s key read/export/import/convert flows onto Go `/api/memory/*` routes when `/trpc` is unavailable, so the dashboard can inherit the newer Go-native saved-context/interchange fallbacks instead of keeping that cluster artificially TS-dependent
- the shared Next.js compat route now also maps the adjacent operator-facing `agentMemory.*` cluster onto Go `/api/agent-memory/*` routes when `/trpc` is unavailable, allowing handoff/pickup, recent memory, and intake flows to inherit the existing Go-native agent-memory ownership instead of remaining artificially TS-dependent
- the shared Next.js compat route now also maps the Project Constitution dashboard cluster onto Go `/api/project/*` routes when `/trpc` is unavailable, and Go now owns truthful local `project.updateContext` file writes, so that page no longer remains artificially TS-dependent for basic read/save behavior
- the shared Next.js compat route now also maps the Skills dashboard/library cluster onto Go `/api/skills/*` routes when `/trpc` is unavailable, so local skill list/read/assimilate behavior can flow through to the UI instead of remaining artificially TS-dependent
- the shared Next.js compat route now also maps the MCP Settings dashboard cluster onto Go `/api/config/*` and `/api/mcp/servers/*` routes when `/trpc` is unavailable, so config rows, client sync targets, export previews, and sync execution can remain truthful in degraded mode
- the shared Next.js compat route now also maps key MCP inspector/search/system mutations onto Go `/api/mcp/*` routes when `/trpc` is unavailable, reducing degraded-mode dependence on TS-era MCP mutation contracts for operator runtime control
- Go now also owns native fallback behavior for API key and workspace-secret writes in the HTTP layer, and the shared Next.js compat route exposes `secrets.list` plus API key/secret admin mutations during `/trpc` outage, making the governance/admin dashboard cluster more usable in Go-primary degraded mode
- Go now also owns native fallback behavior for the DB-backed `tools.setAlwaysOn` mutation through `/api/tools/always-on`, and the shared Next.js compat route exposes that always-on toggle during `/trpc` outage, improving MCP Catalog / MCP Inspector usability in Go-primary degraded mode
- Go now also owns native fallback behavior for policy CRUD through the HTTP layer, and the shared Next.js compat route exposes `policies.list` plus policy mutations during `/trpc` outage, making the Policies governance dashboard cluster more usable in Go-primary degraded mode
- Go now also owns native fallback behavior for tool-set create/delete through the HTTP layer, and the shared Next.js compat route exposes `toolSets.list` plus tool-set mutations during `/trpc` outage, making the Tool Sets dashboard cluster more usable in Go-primary degraded mode
- explicit Node compatibility mode still uses the full workspace build path and still defaults to a full install/build posture
- full builds remain available via `TORMENTNEXUS_FULL_BUILD=1`

---

## Workstream B — SQLite/stateful service migration
### Goal
Remove backend dependency on TS `better-sqlite3` for critical runtime behavior.

### Highest-priority surfaces
- Link crawler / HyperIngest backlog processing
- imported session persistence and maintenance
- transcript deduplication and retention maintenance
- debate history persistence
- any startup-time service that still touches TS SQLite directly

### Exit criteria
- startup no longer depends on TS SQLite bindings for core control-plane behavior
- Go `modernc.org/sqlite` path is authoritative for stateful backend services

---

## Workstream C — MCP full parity in Go
### Goal
Make Go the authoritative MCP backend.

### Remaining parity targets
- configured server CRUD
- runtime mutation flows
- metadata refresh / cache persistence
- telemetry / call history / health persistence
- config import/export and client sync surfaces
- tool preference / working-set mutation surfaces

### Exit criteria
- all MCP read/write surfaces are available natively in Go
- TS MCP routers become optional compatibility layers only

---

## Workstream D — Providers, quotas, and model execution
### Goal
Make Go authoritative for provider execution and routing.

### Tasks
- complete quota/routing parity in Go
- ensure OpenRouter free defaults are mirrored in Go-first execution paths
- move background/service LLM work behind Go routing
- remove TS-only paid-default fallback behavior

### Exit criteria
- provider routing decisions originate in Go
- TS provider execution is no longer authoritative

---

## Workstream E — Sessions, supervision, council, workflows
### Goal
Port orchestration-heavy backend features fully to Go.

### Targets
- session lifecycle parity
- supervisor/task lifecycle parity
- council debate history + orchestration parity
- workflow definition/execution/history parity
- session export/import parity

### Exit criteria
- Go owns orchestration and persistence
- TS becomes UI/client presentation only for these surfaces

---

## Workstream F — Dashboard/client transition
### Goal
Keep UI usable while backend ownership shifts to Go.

### Tasks
- identify which dashboard routes still require TS tRPC contracts
- replace those dependencies with Go HTTP/API contracts
- keep TS UI packages where necessary, but point them at Go-owned APIs

### Exit criteria
- the web dashboard can operate primarily against Go APIs
- TS backend routers are no longer required for normal operator workflows

---

## Concrete near-term execution order
1. Build and maintain a truthful TS→Go parity matrix for all major surfaces
2. Make launcher/runtime selection Go-first
3. Port remaining TS SQLite-backed startup services to Go
4. Finish MCP write/config parity in Go
5. Finish session/supervisor/orchestration parity in Go
6. Move dashboard/backend dependencies from TS routers to Go APIs
7. Retire or demote TS backend ownership surface-by-surface

---

## Immediate next implementation slice
The next coding slice after this plan should be:
1. add a parity matrix doc for backend surfaces
2. add Go-primary launcher/runtime selection
3. continue eliminating TS SQLite ownership from startup-time services

---

## Success definition
The migration is successful when:
- the default TormentNexus startup path runs Go first
- all major backend surfaces are native in Go
- TypeScript is reduced to crucial UI/compatibility roles only
- removing TS backend ownership does not reduce operator-visible functionality
- docs and status surfaces truthfully describe Go as the primary runtime
