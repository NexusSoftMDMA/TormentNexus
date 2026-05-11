# Roadmap

_Last updated: 2026-04-08, version 1.0.0-alpha.12_

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

## Completed (v1.0.0-alpha.1)

Deliverables officially achieved and stabilized in the `1.0.0-alpha.1` milestone:

### 1. Stabilize the core wedge
- **MCP control plane**: Resolved the split-brain config cache, guaranteeing native tools and manual JSON configs synchronize without destructive DB wipes.
- **Provider routing and billing visibility**: Deployed fully functional `CoreModelSelector` fallback logic (handling `EMERGENCY_FALLBACK` and `budget_forced_local` states), seamlessly wired to the TRPC router and the active dashboard.
- **Session supervision**: Confirmed `SessionSupervisor` auto-restart functionality via active regression testing, validating isolated PTY recovery.
- **Memory inspection and continuity**: Deployed Vector Memory manipulation across `search_memory` and `add_memory` MCP tools, connecting universal memory directly to LLMs.
- **System observability**: Deployed robust `McpTrafficInspector` memory ring buffers, actively parsing latency thresholds.

### 2. Dashboard convergence
- completed high-value data-binding work for MCP Server Health, Integrations, and Traffic inspection.
- removed or clearly labeled misleading states (e.g. `borg mcp import` now gracefully errors on missing files instead of printing fake success messages).
- improved empty states and setup guidance (e.g. "Welcome to Borg! Let's get started" first-run banners now guide operators correctly).

### 3. Extension and runtime reliability
- fixed storage access failures across the SQLite database connections.
- reduced workspace build failures (re-aligned `pnpm-workspace.yaml` boundaries for proper transitive closure calculation).
- formalized internal daemon architecture (`borgd`, `hypermcpd`, `hyperingest`, `hyperharnessd`) and placed boundaries in `packages/core/src/daemons/`.
- successfully assimilated `bobbybookmarks` deduplication engines, absorbing 12,000+ validated external URLs and 900+ verified MCP catalogs into the robust internal `.borg` store.

### 4. Release discipline
- unified version story across all packages, dashboards, and CLIs (`1.0.0-alpha.1`).
- synchronized LLM instructions (merged all `CLAUDE.md`, `GEMINI.md`, etc., to point to `docs/UNIVERSAL_LLM_INSTRUCTIONS.md`).
- improved documentation accuracy (`GO_SIDECAR_API.md` explicitly defines what Go natively executes vs proxies).

## Next

### A. MCP operator improvements (BETA)
- ✅ Tool grouping and ranked search (multi-signal scoring)
- ✅ Progressive tool disclosure (6 permanent meta-tools)
- ✅ Auto-load with confidence thresholds
- ✅ Working set with LRU + idle-first eviction
- ✅ Profile-based tool boosting (web-research, repo-coding, etc.)
- ✅ Tool semantic search / Tool RAG via search_tools meta-tool
- ✅ Search-and-use in one shot (auto_call_tool)
- ✅ Eviction history and telemetry tracking
- ✅ Observability dashboard (inspector page)
- ✅ Catalog ingestion (5 adapters: Glama, Smithery, MCP.run, npm, GitHub Topics)
- [ ] Supervisor tool prediction — preemptively inject tool ads based on conversation context
- [ ] Progressive skill disclosure (same architecture as tool disclosure)

### B. Tool parity with CLI harnesses (BETA)
- ✅ Claude Code parity: Read, Write, Edit, MultiEdit, Bash, Glob, Grep, LS, WebFetch
- ✅ Codex CLI parity: shell, apply_diff, create_file, view_file, list_directory, search_files
- ✅ Gemini CLI parity: read_file, write_file, edit_file, list_directory, search
- ✅ OpenCode/Pi parity: read, write, edit, bash, glob, grep, ls, web_fetch
- [ ] Cursor/Windsurf/Kiro-specific tool signatures
- [ ] Goose/Crush/Codebuff/Amp-specific tool signatures
- [ ] Wire tool parity aliases into MCP server tool surface
- [ ] Go-native parity tools (port ToolParityAliases to Go)

### C. Dashboard completeness (BETA)
- ✅ Health page with real server health and crash tracking
- ✅ Tools/Catalog page with real tool inventory and Always On toggling
- ✅ Submodules page with git status and heal action
- ✅ Observability page with real metrics (calls, error rate, latency)
- ✅ Inspector page with working set, telemetry, eviction history
- ✅ Session management and workspace tracking
- [ ] Verify all 69 sub-pages show real data (in progress)
- [ ] Polish empty states and loading indicators
- [ ] Mobile-responsive layout

### D. Session and memory continuity (EXPERIMENTAL)
- [ ] Auto-detect sessions from all AI harnesses (Claude Code, Codex, Gemini CLI, etc.)
- [ ] Session import with LLM extraction of valuable memories
- [ ] Memory subsystem plugin architecture
- [ ] Context harvesting from file changes
- [ ] Memory browser extension integration

### E. Provider and model management (STABLE)
- ✅ Provider fallback chains with automatic model switching
- ✅ Gemini 2.5 Flash free-tier fallback
- ✅ Billing/cost dashboard
- [ ] Free-tier provider chain (OpenRouter, Google AI Studio)
- [ ] OAuth login for Claude Max/Pro, Copilot Premium, ChatGPT Plus
- [ ] Intelligent model selection based on credits/quotas/budgets

### F. Multi-model orchestration (VISION)
- [ ] Multi-model chatroom with shared context
- [ ] Rotating implementer/tester/planner roles
- [ ] Council debate and consensus protocols
- [ ] Autonomous supervisor until completion criteria met

### G. Browser extension (VISION)
- [ ] Chrome/Firefox extension for MCP injection into web chats
- [ ] Session/memory export from web interfaces
- [ ] Browser debug console integration
- [ ] Browser history ingestion into memory

### H. Go parity (BETA)
- ✅ Go sidecar with 543 API routes
- ✅ Go version sync via ldflags
- ✅ Go-native handlers for council, billing, catalog, tools, etc.
- [ ] Port remaining TypeScript handlers to Go
- [ ] Go-native MCP router
- [ ] Make Go the default runtime
- stronger import/export clarity
- better working-set management
- groundwork for benchmarking, ranking, and operator review loops across discovered MCP servers
- keep registry-source adapters truthful about source drift, partial availability, and non-fatal ingestion failures instead of treating stale registries as healthy empty catalogs

### B. Memory quality
- better retrieval tuning
- stronger import/export ergonomics

### C. Session workflow quality
- cleaner create/edit flows
- clearer isolation behavior
- converge primary CLI harness support around first-class Borg identities, starting with `borg`

### D. Provider routing polish
- clearer fallback history
- stronger quota truthfulness
- more actionable auth-state and routing config

### E. Architecture convergence
- converge the repo toward the recommended Borg binary family without splitting everything at once
- turn current packages into clearer extraction seams for `borgd`, `hypermcpd`, `hypermemd`, `hyperingest`, and `hyperharnessd`
- keep CLIs and GUIs as clients of daemon-owned state
- keep shared contracts and config stable before promoting them into cross-process APIs
- treat the current Go workspace as an **Experimental** coexistence lane for truthful read-parity and bridge-first replacement work, not as proof that the daemon boundaries are already extracted
- use the Go lane to validate which reads can be backed by the same SQLite/file/config truth sources before promoting any service boundary claims

## Later

- mobile and desktop companion polish after core stabilization
- safer public tool ingestion and sandboxing
- richer operator automation with guardrails
- deeper benchmark and latency reporting
- graduate selected daemon boundaries into real standalone binaries once uptime, isolation, or deployment needs justify them

## Vision

These remain exploratory until the control plane is stronger:
- advanced council or debate systems
- a definitive internal library of MCP servers aggregated from public lists and kept refreshed inside Borg
- benchmarking and comparative ranking across competing MCP server implementations
- eventual model reach to any relevant MCP tool through one operator-controlled control plane
- an operator-owned substrate spanning any model, any provider, any session, and any relevant tool
- multi-node federation or mesh-style coordination
- marketplace or community distribution systems
- performance-critical Rust components
- richer graph or 3D cognition visualizations

## Proposed `v1.0.0` bar

A credible `v1.0.0` should mean:
- reliable start and core workflow execution
- honest operator pages with real backend state
- dependable MCP inspection and config flows
- useful provider fallback visibility
- meaningful session and memory continuity
- documentation that does not overclaim

## Explicit de-scope until v1 ships

Do not prioritize these ahead of core convergence:
- economy or payment ideas
- broad P2P claims
- major rewrites without measured need
- large net-new surfaces while core ones remain rough
# Roadmap

_Last updated: 2026-04-02_

## Status legend

- **Stable** — production-intended, tested, maintained
- **Beta** — usable, still evolving
- **Experimental** — active R&D, not dependable
- **Vision** — directional only

## Framing

borg has two jobs at once:
1. ship a reliable local control plane,
2. preserve a credible long-term vision for richer AI orchestration.

This roadmap keeps those jobs separate.

## Completed (v1.0.0-alpha.1)

Deliverables officially achieved and stabilized in the `1.0.0-alpha.1` milestone:

### 1. Stabilize the core wedge
- **MCP control plane**: Resolved the split-brain config cache, guaranteeing native tools and manual JSON configs synchronize without destructive DB wipes.
- **Provider routing and billing visibility**: Deployed fully functional `CoreModelSelector` fallback logic (handling `EMERGENCY_FALLBACK` and `budget_forced_local` states), seamlessly wired to the TRPC router and the active dashboard.
- **Session supervision**: Confirmed `SessionSupervisor` auto-restart functionality via active regression testing, validating isolated PTY recovery.
- **Memory inspection and continuity**: Deployed Vector Memory manipulation across `search_memory` and `add_memory` MCP tools, connecting universal memory directly to LLMs.
- **System observability**: Deployed robust `McpTrafficInspector` memory ring buffers, actively parsing latency thresholds.

### 2. Dashboard convergence
- completed high-value data-binding work for MCP Server Health, Integrations, and Traffic inspection.
- removed or clearly labeled misleading states (e.g. `borg mcp import` now gracefully errors on missing files instead of printing fake success messages).
- improved empty states and setup guidance (e.g. "Welcome to borg! Let's get started" first-run banners now guide operators correctly).

### 3. Extension and runtime reliability
- fixed storage access failures across the SQLite database connections.
- reduced workspace build failures (re-aligned `pnpm-workspace.yaml` boundaries for proper transitive closure calculation).
- formalized internal daemon architecture (`borgd`, `borgmcpd`, `borgingest`, `borgharnessborgharnessd`) and placed boundaries in `packages/core/src/daemons/`.
- successfully assimilated `bobbybookmarks` deduplication engines, absorbing 12,000+ validated external URLs and 900+ verified MCP catalogs into the robust internal `.borg` store.

### 4. Release discipline
- unified version story across all packages, dashboards, and CLIs (`1.0.0-alpha.1`).
- synchronized LLM instructions (merged all `CLAUDE.md`, `GEMINI.md`, etc., to point to `docs/UNIVERSAL_LLM_INSTRUCTIONS.md`).
- improved documentation accuracy (`GO_SIDECAR_API.md` explicitly defines what Go natively executes vs proxies).

## Next

### A. MCP operator improvements
- better tool grouping and search
- stronger import/export clarity
- better working-set management
- groundwork for benchmarking, ranking, and operator review loops across discovered MCP servers
- keep registry-source adapters truthful about source drift, partial availability, and non-fatal ingestion failures instead of treating stale registries as healthy empty catalogs

### B. Memory quality
- better retrieval tuning
- stronger import/export ergonomics

### C. Session workflow quality
- cleaner create/edit flows
- clearer isolation behavior
- converge primary CLI harness support around first-class borg identities, starting with `borg`

### D. Provider routing polish
- clearer fallback history
- stronger quota truthfulness
- more actionable auth-state and routing config

### E. Architecture convergence
- converge the repo toward the recommended borg binary family without splitting everything at once
- turn current packages into clearer extraction seams for `borgd`, `borgmcpd`, `borgmemd`, `borgingest`, and `borgharnessborgharnessd`
- keep CLIs and GUIs as clients of daemon-owned state
- keep shared contracts and config stable before promoting them into cross-process APIs
- treat the current Go workspace as an **Experimental** coexistence lane for truthful read-parity and bridge-first replacement work, not as proof that the daemon boundaries are already extracted
- use the Go lane to validate which reads can be backed by the same SQLite/file/config truth sources before promoting any service boundary claims

## Later

- mobile and desktop companion polish after core stabilization
- safer public tool ingestion and sandboxing
- richer operator automation with guardrails
- deeper benchmark and latency reporting
- graduate selected daemon boundaries into real standalone binaries once uptime, isolation, or deployment needs justify them

## Vision

These remain exploratory until the control plane is stronger:
- advanced council or debate systems
- a definitive internal library of MCP servers aggregated from public lists and kept refreshed inside borg
- benchmarking and comparative ranking across competing MCP server implementations
- eventual model reach to any relevant MCP tool through one operator-controlled control plane
- an operator-owned substrate spanning any model, any provider, any session, and any relevant tool
- multi-node federation or mesh-style coordination
- marketplace or community distribution systems
- performance-critical Rust components
- richer graph or 3D cognition visualizations

## Proposed `v1.0.0` bar

A credible `v1.0.0` should mean:
- reliable start and core workflow execution
- honest operator pages with real backend state
- dependable MCP inspection and config flows
- useful provider fallback visibility
- meaningful session and memory continuity
- documentation that does not overclaim

## Explicit de-scope until v1 ships

Do not prioritize these ahead of core convergence:
- economy or payment ideas
- broad P2P claims
- major rewrites without measured need
- large net-new surfaces while core ones remain rough

8. **Agent Communication Protocols [P3]**
   - *Implementation:* Integrate A2A discovery for URL processing delegation.
   - *Implementation:* Integrate AG-UI generative UI for Kinetic HUD.

9. **Graph/RAG/Vector Intelligence [P1]**
   - *Implementation:* Build tag co-occurrence graph. Use DuckDB VSS for semantic search over existing dataset.

10. **Security & Guardrails [P2]**
    - *Implementation:* SafeExec execution blocking, PII masking, "AI Receipts" logging.

11. **Observability & Telemetry [P1]**
    - *Implementation:* Add OTEL spans to all extraction steps. Add session recording for JSON-RPC payloads.

12. **Model Intelligence & Cost Optimization [P0]**
    - *Implementation:* Tiered routing: Simple URLs (SLM) vs Complex URLs (Frontier). Auto-demote on ELO drop.

13. **Browser Automation [P2]**
    - *Implementation:* Accessibility Tree extraction for SPAs to achieve 90% token reduction vs DOM parsing.

14. **CI/CD & Git Workflows [P3]**
    - *Implementation:* Version DB schema via Git. Autonomous self-healing queues for low-quality extractions.
