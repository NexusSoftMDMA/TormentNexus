# Go Backend Parity Matrix

## Purpose

Track the migration from **TypeScript-primary** backend ownership to **Go-primary** backend ownership.

Status values:

- **Native Go** — implemented and intended to become authoritative in Go
- **Bridge-first** — Go route exists and attempts TS first, then falls back natively
- **Bridge-only** — Go surface exists but only proxies to TS
- **Missing** — no meaningful Go backend parity yet
- **TS-only critical** — still materially owned by TS and blocks Go-primary startup parity

---

## 1. Runtime / startup / lifecycle

| Surface | Current status | Notes |
|---|---|---|
| Go HTTP control plane | Native Go | `go/cmd/tormentnexus serve` works and exposes large `/api/*` surface |
| Lock/status introspection | Native Go | runtime/config/lock endpoints exist |
| Primary launcher ownership | Native Go | start.bat defaults to Go-primary when binary exists ("./tormentnexus.exe"). Use --runtime node for TS-primary. |
| Non-destructive occupied-port behavior | Native Go | start.bat Go-primary path retries on port conflict, falls back to 7777, reports status clearly. |
| Go-first default startup | Native Go | start.bat --runtime auto (default) detects Go binary and uses Go-primary path. Falls back to TS if Go unavailable. Use --runtime go or --runtime node to force. |

---

## 2. MCP platform

| Surface | Current status | Notes |
|---|---|---|
| MCP inventory snapshot | Bridge-first / Partial Native Go | Go inventory now reads live `mcp.jsonc`, local `tormentnexus.db`, and a Go-owned persisted inventory cache file (`mcp_inventory_cache.json`) in the Go config dir; JSONC save/mutation flows actively resync that cache from live sources, canonical metadata-tool normalization is shared between JSONC metadata and inventory generation, and fallback inventory views can combine persisted runtime-overlay cache data with live runtime-registry overlay while surfacing per-layer cache freshness/source metadata |
| Tool listing/search/call | Bridge-first with Partial Native Go fallback | native aggregation and ranking exist, Go fallback prefers persisted local inventory cache data for MCP tool list/search plus secondary `/api/tools*` cache-backed recovery when local DB rows are unavailable, and cache-backed responses now surface source/freshness metadata plus distinct persisted-vs-live runtime overlay counts and age/staleness heuristics per layer; fallback records expose a stable nested `provenance` object marked as the primary contract, including explicit `legacyMirrorFields` metadata, and low-risk search/list surfaces plus `/api/tools/get` have now trimmed redundant top-level provenance mirrors |
| Runtime server list/status | Bridge-first / Partial Native Go fallback | fallback runtime-server responses now carry server-level origin-layer and layer freshness metadata, can recover from persisted runtime-overlay cache when live summary detection is unavailable, and expose the same nested `provenance` schema used by tool/configured-server fallback records, marked as the primary contract with explicit `legacyMirrorFields`; low-risk list-style runtime-server responses have now begun trimming redundant top-level provenance mirrors |
| Configured server CRUD | Partial Native Go | Go has native JSONC-backed configured-server create/update/delete plus JSONC-first read fallback, and configured-server read surfaces now expose record-level origin-layer and metadata freshness/provenance fields across JSONC-backed and DB-backed fallback payloads, including the same nested `provenance` schema used by other MCP fallback records and marked as the primary contract with explicit `legacyMirrorFields`; configured-server list and get fallback responses have now begun trimming redundant top-level provenance mirrors while the remaining highest-risk config/detail surfaces stay compatibility-heavy for now; broader ecosystem authority cleanup still remains |
| Runtime server add/remove/mutation | Partial Native Go | Go has a native runtime-server registry for add/remove/list fallback behavior, and the durable subset of successfully probed runtime server/tool metadata now syncs into the canonical inventory cache; full transport/lifecycle parity is still incomplete |
| Metadata refresh/cache management | Partial Native Go | native JSONC inspection/cache normalization exists, probeable STDIO servers can attempt live `tools/list` refresh, local JSONC write paths resync the Go inventory cache from live sources so metadata clear/refresh actions do not leave stale inventory cache state behind, and canonical metadata-tool mapping is now shared between JSONC metadata handling and inventory/cache generation; broader multi-transport live discovery parity is still incomplete |
| Telemetry/history/write surfaces | Partial Native Go | Go now has native persisted local MCP working-set state, eviction history, and tool-selection telemetry fallback behavior via `mcp_state.json` in the Go config dir; richer multi-session/runtime parity is still incomplete |
| MCP config import/export/client sync | Bridge-only | not yet full native authority |

---

## 3. Providers / model routing / quotas

| Surface | Current status | Notes |
|---|---|---|
| Provider catalog/status summary | Native Go | provider catalog/status/routing summary endpoints exist |
| Native provider execution | Native Go | Anthropic/OpenAI/Google/DeepSeek/OpenRouter implemented |
| Auto-routing fallback | Native Go | provider selection and fallback logic exist |
| Quota-aware backend authority | Partial Native Go | routing exists, but TS still owns some live provider/quota surfaces |
| Billing/task routing admin mutations | Native Go with upstream fallback | All billing routes have native Go fallbacks. Read endpoints use providers package for real data. Write endpoints (setRoutingStrategy, setTaskRoutingRule) record mutations locally. |
| Background-service LLM execution authority | Native Go | Go has /api/llm/generate endpoint with WaterfallClient provider routing. Supports direct model selection and task-type routing. TS services can call this endpoint instead of invoking providers directly. |

---

## 4. Memory / persistence / SQLite-backed state

| Surface | Current status | Notes |
|---|---|---|
| Sectioned memory status/search | Bridge-first with Native Go fallback | native SQLite fallback exists |
| Memory search endpoints | Bridge-first with Partial Native Go fallback | Go now owns persisted local fallback capture/search for agent-memory-backed facts, observations, user prompts, session summaries, pivot search, timeline windows, cross-session links, a truthful generic `memory.query` fallback that can merge local SQLite-backed memory rows with persisted `.tormentnexus/memory/contexts.json` context-registry results, plus truthful local memory interchange format/import/convert fallback behavior across `json`, `json-provider`, `jsonl`, `csv`, and `sectioned-memory-store`; broader generic memory graph/vector parity is still incomplete |
| Agent memory runtime | Partial Native Go | Go now owns persisted local add/delete/clear/search/recent/filter/export/stats/handoff/pickup fallback behavior through `.tormentnexus/agent_memory/memories.json`, and the adjacent saved-context fallback now also supports local context creation, local context-body reads, and local registry deletion from `.tormentnexus/memory/contexts.json`; deeper vector-backed parity and mixed-runtime authority are still TS-first |
| Imported session storage | Partial Native Go | Go now has a real imported-session store plus native file-based and DB-backed ingest paths (`llm-cli` logs.db and TormentNexus DB artifacts), with transcript-hash dedup, archived transcript persistence, memory rows, and instruction-doc regeneration; some niche parser parity gaps still remain |
| Imported session docs/maintenance | Partial Native Go | Go now provides instruction-doc generation/listing and maintenance stats from the native store, though mixed-runtime cleanup is still ongoing |
| Transcript dedup / retention maintenance | Partial Native Go | transcript-hash dedup now exists in the native Go imported-session store, but broader retention/backfill ownership is still incomplete |
| Workspace/config/secret persistence | Mixed | some Go read surfaces, TS still owns many writes |
| Debate history persistence | Partial Native Go | Go now has a native debate-history store plus native council-history read/write fallbacks, but TS config/policy semantics and broader mixed-runtime cleanup are still incomplete |
| Windows/Node SQLite reliability | Go-preferred | major motivation for migration away from TS `better-sqlite3` |

---

## 5. Session lifecycle / supervision / import-export

| Surface | Current status | Notes |
|---|---|---|
| Session summary/discovery | Native Go / Bridge mix | scanner and summary routes exist |
| Session import scan | Partial Native Go | Go scanner/import routes exist |
| Session export | Native Go | native export path implemented |
| Session supervisor lifecycle | Bridge-first with Partial Native Go fallback | the public supervisor route family now tries TS first and falls back natively in Go for `list/get/create/start/stop/restart/logs/attach-info/health/execute-shell/restore`; the native Go fallback is now durable across Go runtime restarts via `.tormentnexus-go/session-supervisor.json` and can explicitly reload persisted inventory on demand, but full TS parity and shared authority are still incomplete |
| Session state/log parity | Bridge-first with Partial Native Go fallback | Go now has native persisted fallback ownership for the shared session-state core (`getState`, `updateState`, `clear`, `heartbeat`) via workspace `.tormentnexus-session.json`, plus native persisted supervisor lifecycle/log/attach/health/session-shell fallback behavior for public supervised sessions; execution-policy visibility and worktree/isolation behavior are now much closer to TS parity, but fuller memory-bootstrap parity is still incomplete |
| Session CRUD authority | Mixed / bridge-heavy | Go now owns more public supervisor CRUD/read behavior during TS outage and can restore its own fallback inventory durably; the shared Next.js compat route now also maps the session dashboard's key supervisor reads/mutations onto Go `/api/sessions/supervisor/*` routes when `/trpc` is unavailable, but full end-to-end authority and shared TS/Go restore semantics are still incomplete |

---

## 6. Workflows / orchestration / councils

| Surface | Current status | Notes |
|---|---|---|
| Native workflow engine | Native Go | DAG engine + built-ins implemented |
| Workflow API parity | Partial Native Go | native endpoints exist, not yet full TS parity |
| Council debate endpoint | Bridge-first with Native Go fallback | native Go council fallback exists and now persists native fallback debates into the Go debate-history store |
| Council history/persistence | Partial Native Go | Go history status/stats/list/get/delete/supervisor/clear/initialize fallbacks now use native persisted debate history when TS is unavailable |
| Swarm/squad/autodev/darwin | Partial Native Go / bridge mix | many Go routes still proxy to TS; `/api/skills/assimilate` now has a native Go fallback that creates a truthful local starter skill scaffold, Darwin routes now have a native persisted local mutation/experiment/status fallback, AutoDev routes now have a native persisted local loop manager fallback, squad routes now have a native persisted local squad/indexer state fallback, and swarm mission start/resume/history/risk/mesh plus local approve/decompose/update/debate/consensus/direct-message fallback semantics now have native persisted local fallback ownership, but richer director/worktree/LLM/research-driven squad, swarm, Darwin, AutoDev, and assimilation parity remain incomplete |
| Director config/status | Native Go with upstream fallback | status, config get/test/update, auto-drive start/stop all have native Go fallbacks. memorize/chat remain bridge-only (LLM-dependent). |

---

## 7. Background workers / ingestion

| Surface | Current status | Notes |
|---|---|---|
| BobbyBookmarks sync | Native Go exists + TS worker still present | Go sync implemented; TS worker still exists |
| Link backlog crawl/tag enrichment | Partial Native Go | native Go crawler utility, HTTP endpoint, and Go server-owned background worker lifecycle now exist; TS worker still remains in the mixed-runtime world |
| Session auto-import worker | Native Go | Go-native background worker in PreWarmCaches() scans and imports sessions periodically using sessionimport.IngestDiscoveredSessions. |
| Transcript maintenance jobs | Native Go | Go-native background worker reports maintenance stats via ImportedSessionStore.GetMaintenanceStats. Runs 5min after startup, then every 24h. |
| Background ingestion ownership | Mixed | key migration target area |

---

## 8. Git / repo / submodules / system

| Surface | Current status | Notes |
|---|---|---|
| Submodule listing/update | Native Go | native fallback implemented and tested |
| Runtime/system/submodule summary | Native Go | cloud/system routes exist |
| Git/system bridge coverage | Partial | some surfaces still bridge or remain TS-owned |

---

## 9. Streaming / extensions / operator APIs

| Surface | Current status | Notes |
|---|---|---|
| SSE broker | Native Go | implemented for extension parity |
| Browser extension support APIs | Partial Native Go / bridge mix | some parity exists, not complete |
| Dashboard backend contract | TS-only critical | web app still materially depends on TS-oriented surfaces/contracts, but Go-primary startup can now launch the Next.js dashboard in compatibility-backed mode against the live Go control plane instead of hard-skipping dashboard startup entirely; the shared compat route now also covers the supervised-session dashboard cluster, the memory dashboard’s key read/export/import/convert flows, the operator-facing `agentMemory.*` cluster, the Project Constitution dashboard cluster, the Skills dashboard/library cluster, the MCP Settings dashboard cluster (config rows plus client-config sync flows), key MCP inspector/runtime-control mutations, API key/secret admin actions, the DB-backed `tools.setAlwaysOn` toggle, policy dashboard CRUD, and tool-set dashboard mutations during `/trpc` outage |
| Native client/operator APIs | Growing Native Go | good progress, but not enough for full backend replacement |

---

## Highest-priority blockers to true Go-primary status

1. **Primary launcher still TS-centered**
   - Go is available, but not yet the default authoritative startup path.
2. **TS SQLite-backed startup services still matter**
   - imported-session maintenance
   - transcript dedup
   - debate history persistence
   - link crawler / HyperIngest ownership
3. **MCP write/config parity is incomplete in Go**
   - CRUD/mutation/cache/telemetry surfaces still lag.
4. **Dashboard still depends heavily on TS-era backend contracts**
   - Go can now host the primary control plane while startup launches the Next.js dashboard in compatibility-backed mode, and the supervised-session dashboard cluster plus key MCP runtime-control mutations, API key/secret admin actions, the DB-backed tool always-on toggle, policy dashboard CRUD, and tool-set dashboard mutations now map onto Go routes during `/trpc` outage, but the UI still does not yet rely on a fully Go-authoritative backend contract.
5. **Background-service LLM execution is still partly TS-owned**
   - despite OpenRouter-free default migration, execution ownership is not fully in Go.

---

## Recommended migration order

### Phase 1 — Control-plane authority

- implement Go-primary launcher/runtime selection
- make Go the default owner of the primary control-plane port
- keep TS as explicit compatibility supplement only

### Phase 2 — Stateful/startup-critical migrations

- port remaining SQLite-backed startup services to Go
- remove startup dependence on TS `better-sqlite3`
- move HyperIngest/link-crawl ownership to Go

### Phase 3 — MCP backend completion

- finish native Go CRUD/mutation/config/telemetry surfaces
- retire TS ownership of MCP backend state

### Phase 4 — Orchestration parity

- finish session/council/workflow/supervisor persistence and control parity
- reduce bridge-first behavior where Go implementations are stable

### Phase 5 — UI/backend contract migration

- point dashboard/native clients at Go-owned APIs
- demote TS backend to compatibility only

---

## Definition of done

Go is the primary version when:

- the default startup path is Go-first
- all major backend surfaces above are **Native Go**
- TS backend ownership is no longer required for normal operator workflows
- remaining TS code is limited to crucial UI/client compatibility roles
