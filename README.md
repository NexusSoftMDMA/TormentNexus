# Borg: The Cognitive Control Plane & Universal AIOS

![Version](https://img.shields.io/badge/version-1.0.0--alpha.45-blue)
![Build](https://img.shields.io/badge/build-passing-brightgreen)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)
![TypeScript](https://img.shields.io/badge/TypeScript-5.x-3178C6?logo=typescript)

**Borg** is the ultimate local-first control plane for multi-agent workflows, Model Context Protocol (MCP) tooling, provider routing, session continuity, and operator observability.

We are building the substrate where a single local system seamlessly coordinates the most critical parts of AI-driven software development: tools, models, sessions, context, subagents, and full visibility across the entire stack. Borg is not just an aggregator; it is a **decision system and universal bridge**.

---

## 🏗️ The Architecture (Modular Monolith)

Borg has evolved into a high-performance **Go (Golang) modular monolith** with a **TypeScript/Next.js frontend**.
* **The Go Sidecar (`go/internal/`)**: Go handles the heavy lifting—orchestration, progressive MCP routing, L1/L2 memory management, and LLM waterfall routing.
* **The Control Panel (`apps/web/`)**: A rich Next.js and React dashboard serving as your visual observation deck.
* **The Storage (`sqlite-vec`)**: Dependency-free, hyper-fast local vector search for omniscient memory and tool ranking.

## ✨ Core Pillars

### 1. Progressive MCP Tool Routing & Parity
Models should never be overwhelmed with a 50,000-token tool dump. Borg employs a multi-layered, progressive disclosure system:
* **Semantic Search:** Local vector embeddings match the active prompt against a global MCP directory.
* **The Router:** Only the top highly relevant tool schemas are injected into the active LLM context.
* **Universal Parity:** Byte-for-byte identical tool signatures for Claude Code, Codex, Gemini CLI, Cursor, and Windsurf.

### 2. Dual-Tier Memory Architecture (L1 / L2)
Context is finite; memory must be infinite.
* **L1 - Session Scratchpad:** Ephemeral, lightning-fast memory tied directly to the active session.
* **L2 - The Vault:** Permanent semantic storage in SQLite. Saves exact transcripts and LLM-compressed heuristics.
* **Context Harvesting:** Every session autonomously queries the L2 Vault to pull in relevant historical heuristics.

### 3. The Resilient LLM Waterfall
Uptime is non-negotiable. Borg’s inference client natively catches 429s (Rate Limits) and 5xx (Server Errors), seamlessly cascading the exact payload down a prioritized chain without crashing:
1. **NVIDIA NIM** / Primary APIS
2. **OpenRouter** (Secondary aggregator fallback)
3. **Local LM Studio / Ollama** (Ultimate offline fallback)

### 4. Multi-Agent Swarm & P2P Mesh
Borg coordinates specialized models inside shared chatrooms via the Agent-to-Agent (A2A) protocol.
* **Role Rotation:** Models take turns acting as Planner, Implementer, Tester, and Critic.
* **Consensus & Debate:** Agents autonomously bid on tasks, share context via a neural transcript, and debate implementations until consensus is reached.

### 5. Truth Over Hype Dashboards
Borg's dashboards reflect actual SQLite database rows and active Go goroutine states. No mocked UI scaffolds. Monitor telemetry, traffic inspection, working-set capacity, and LLM routing histories in real-time.

---

## 🚀 Quick Start

**Prerequisites:**
* Node.js 24+
* Go 1.26+
* pnpm v10

**Installation:**
```bash
# 1. Clone the repository
git clone [https://github.com/robertpelloni/borg.git](https://github.com/robertpelloni/borg.git)
cd borg

# 2. Install dependencies & rebuild SQLite bindings
pnpm install
pnpm rebuild better-sqlite3

# 3. Build the Go sidecar
cd go && go build -buildvcs=false ./cmd/borg && cd ..

# 4. Start the Borg Control Plane
pnpm run dev
```




---

# borg

**The local-first control plane for AI operations.**

> Status: **Pre-1.0 convergence**
> Focus: **stability, truthfulness, and operator trust**

borg helps operators run a fragmented AI tool stack from one local control plane. It is designed for people who already use multiple MCP servers, multiple model providers, and multiple coding or session workflows—and want one place to inspect, route, recover, and understand them.

## What borg is

borg is primarily four things:

1. **MCP control plane** — manage and inspect MCP servers and tool inventories from one local service.
2. **Provider routing layer** — handle quota-aware fallback across model providers.
3. **Session and memory substrate** — preserve continuity across work sessions.
4. **Operator dashboard** — make runtime state visible and diagnosable.

## Why this project exists

Modern AI work is messy:
- too many MCP servers,
- too many providers and quotas,
- too many half-connected tools,
- too little context continuity,
- and weak observability when something breaks.

borg exists to reduce that fragmentation without requiring a hosted backend.

## What is real today

### Stable
- **31 CLI Commands**: Full control plane (MCP, Sessions, Providers, Knowledge, Swarm, Cloud Dev).
- **Go Sidecar Bridge**: 543 REST API routes providing truthful telemetry for catalog, memory, and routing.
- **MCP Fleet Management**: Multi-process supervision with PID tracking (12/16 servers alive).
- **System Metrics & Inventory**: Real-time 32-core AMD64 host monitoring and 51-tool/49-harness mapping.
- **73/73 Tests Pass**: 100% success rate across smoke, CLI integration, and workflow suites.
- **Dashboard Convergence**: 86/86 pages bound to live tRPC/REST endpoints.
- **Server Resilience**: Verified 20-hour continuous uptime in production simulation.
- **Build & Typecheck**: All four compilation targets at zero errors.

### Beta
- **Session Supervision**: PID-tracked PTY recovery and process isolation.
- **Swarm Orchestration**: Debate protocols, consensus engines, and 3 active mission types.
- **Cloud Dev (Jules)**: Verified Google Autopilot integration with full session lifecycle.
- **Knowledge RAG**: 13,478 memory-backed nodes with graph/stats visibility.
- **Provider Routing**: Quota-aware fallback (Google → OpenAI) with fallback history.
- **Skill Store**: 4 registered skills with show/list/create workflows.
- **Memory Substrate**: 14,708 entries across long-term and working segments.

### Experimental
- borg assimilation via `submodules/borg` plus primary borg CLI harness registration
- Council or debate workflows
- Broader autonomous workflow layers
- Mobile and desktop parity layers
- Mesh and marketplace concepts

### Vision
- A definitive internal library of MCP servers and tool metadata aggregated from public lists and operator-added sources
- Continuous normalization, deduplication, and refresh of that MCP library inside borg
- Eventual operator-controlled access to any relevant MCP tool through one local control plane
- Operator-owned discovery, benchmarking, and ranking of the MCP ecosystem so borg knows what tools exist, how well they work, and when to trust them
- A universal model-facing substrate where any model, any provider, any session, and any relevant MCP tool can be coordinated through borg

## What borg is not yet

borg is **not yet** a fully hardened universal “AI operating system.” The most honest current description is:

> borg is an ambitious, local-first AI control plane with real implementation across MCP routing, provider management, sessions, and memory—plus a broader experimental layer around orchestration and automation.

## Current focus

The current release track centers on:
- core MCP reliability,
- provider routing correctness,
- practical memory usefulness,
- session continuity,
- and honest dashboard or operator UX.

Longer-term, borg should become the place where operators maintain a definitive internal MCP server library, benchmark the live tool ecosystem, and expose universal tool reach through one operator-owned control plane. That ambition is intentionally large, but it is still **Vision** work until the current control plane is more reliable.

## Orchestrator identities

borg currently presents three operator-facing orchestrator identities:

- `packages/cli` is the **cli-orchestrator** lane.
- `apps/maestro` is the desktop **electron-orchestrator** lane.
- `apps/cloud-orchestrator` is the web **cloud-orchestrator** lane.

The experimental Go workspace under `go/` is a sidecar **cli-orchestrator** coexistence port for read-parity and feasibility work, not a replacement fork and not yet the primary control-plane implementation.

Today, `electron-orchestrator` and `cli-orchestrator` do **not** yet have 100% feature parity. The desktop lane currently exposes the broader operator UX, while the Node-based CLI lane remains the cleaner control-plane foundation. borg should not drop either surface until parity gaps and operator workflows are intentionally closed. The Go lane should currently be described as **Experimental** read-only bridge replacement work, not as a completed daemon extraction.

## Quick start

### Requirements
- Node.js 22+ (tested on Node 24)
- Go 1.22+
- pnpm 10+

### Local development
```bash
make install   # pnpm install + rebuild native modules
make build     # Go binary + TS core + TS CLI
make typecheck # Verify all targets at 0 errors
make dev       # Start development server
```

### borg harness lane
```bash
borg session harnesses
borg session start ./my-app --harness borg
borg mesh status
```

`borg` is now borg's primary CLI harness identity, backed by the `submodules/borg` upstream. The upstream now exposes a Go/Cobra CLI with a default TUI REPL plus a `pipe` command, and borg now surfaces borg's source-backed tool inventory from `submodules/borg/tools/*.go` via `borg session harnesses` and the Go sidecar harness registry. borg's harness catalogs now also track the broader known external identities it already references elsewhere in the repo, including `aider`, `cursor`, `copilot`, `qwen`, `superai-cli`, `codebuff`, `codemachine`, and `factory-droid`, but those still expose install/runtime metadata only until borg has equally source-backed bridge contracts for them. borg's maturity remains **Experimental** while the cross-runtime adapter contract is still shallow.

The CLI mesh surface is now operator-visible through `borg mesh status`, `borg mesh peers`, `borg mesh capabilities [nodeId]`, and `borg mesh find --capability <name>`. These commands query the live local control plane through `BORG_TRPC_UPSTREAM` or the borg startup lock, so they report real mesh visibility instead of placeholder CLI output.

### Docker
```bash
docker compose up --build
```

## Repository shape

```text
apps/
  web/              Next.js dashboard
  borg-extension/   Browser extension surfaces (compatibility path)
  maestro/          electron-orchestrator desktop shell work (legacy path)
  vscode/           VS Code integration

packages/
  core/             Main control plane backend
  ai/               Provider/model routing
  cli/              cli-orchestrator entrypoints
  ui/               Shared UI package
  types/            Shared types

submodules/
  borg/        External borg harness upstream (experimental assimilation track)

go/
  cmd/borg/         Experimental sidecar Go cli-orchestrator port workspace

The Go port is intentionally isolated from the main Node/Next fork. It uses its own `.borg-go` config directory and can observe the primary borg lock state via `/api/runtime/locks`, summarize its interop visibility via `/api/runtime/status` including compact lock visibility/running counts, config-path health, total and available CLI tool/harness counts, provider totals plus configured/authenticated/executable counts and auth/task buckets, memory availability plus default-section and per-section entry breakdowns, discovered-session counts plus session-type, task, model-hint, and TypeScript supervisor-bridge visibility, and import-root plus import-source health including valid/invalid counts, aggregate estimated size, and compact source-type, model-hint, and error buckets, expose a self-describing route index via `/api/index`, inspect effective path wiring via `/api/config/status` including repo-level `borg.config.json` and `mcp.jsonc` presence, expose read-only provider credential visibility via `/api/providers/status`, expose provider catalog metadata via `/api/providers/catalog`, expose compact provider rollups via `/api/providers/summary`, preview intended task-type routing order via `/api/providers/routing-summary`, read the main fork's generated imported-instructions artifact via `/api/runtime/imported-instructions`, expose discovered session artifacts through `/api/sessions` and `/api/sessions/summary`, and bridge or selectively replace TypeScript read routes across `/api/sessions/supervisor/*`, `/api/sessions/imported/*`, `/api/mcp/*`, `/api/memory/*`, `/api/agent-memory/*`, `/api/graph/*`, `/api/context/*`, `/api/git/*`, `/api/tests/*`, `/api/metrics/*`, `/api/logs/*`, `/api/server-health/*`, `/api/settings/*`, `/api/tools/*`, `/api/tool-sets/*`, `/api/project/*`, `/api/shell/*`, `/api/agent/*`, `/api/commands/*`, `/api/skills/*`, `/api/workflows/*`, `/api/symbols/*`, `/api/lsp/*`, `/api/api-keys/*`, `/api/audit/*`, `/api/scripts/*`, `/api/links-backlog/*`, `/api/infrastructure/*`, `/api/expert/*`, `/api/policies/*`, `/api/secrets/*`, `/api/marketplace/*`, `/api/catalog/*`, `/api/oauth/*`, `/api/research/*`, `/api/pulse/*`, `/api/session-export/*`, `/api/browser-extension/*`, `/api/open-webui/*`, `/api/code-mode/*`, `/api/submodules/*`, `/api/suggestions/*`, and `/api/plan/*`. Some of those reads now have truthful local Go fallbacks backed by the same SQLite database, local config files, or deterministic local defaults, but many orchestration-heavy routes remain bridge-only by design. Its current role is to validate a Go-native cli-orchestrator path, grow honest read-only local truth where practical, and avoid overstating daemon-extraction maturity before the underlying contracts are stable.
```

## Recommended binary-to-package evolution

The repo does **not** yet ship the full recommended borg binary family, but the current workspace already suggests the right extraction seams.

### Control plane

- Future binaries: `borg`, `borgd`
- Current likely sources:
  - `packages/cli`
  - `packages/core`
  - `packages/ai`
  - `packages/types`
  - `packages/tools`
  - `go/cmd/borg`
  - `go/internal/controlplane`, `go/internal/httpapi`, `go/internal/providers`

### MCP layer

- Future binaries: `borgmcpd`, `hypermcp-indexer`
- Current likely sources:
  - `packages/mcp-client`
  - `packages/mcp-registry`
  - `packages/mcp-router-cli`
  - MCP-related surfaces inside `packages/core`
  - `go/internal/httpapi` and future Go MCP-specific packages as extraction work continues

### Memory and ingestion layer

- Future binaries: `borgmemd`, `borgingest`
- Current likely sources:
  - `packages/memory`
  - `packages/claude-mem`
  - session and import flows inside `packages/core`
  - `go/internal/memorystore`
  - `go/internal/sessionimport`

### Harness layer

- Future binaries: `borgharnessborgharness`, `borgharnessborgharnessd`
- Current likely sources:
  - `packages/agents`
  - `packages/adk`
  - `packages/borg-supervisor`
  - `packages/browser`
  - `packages/search`
  - harness registration and supervisor flows in `packages/core`
  - `go/internal/harnesses`

### Client surfaces

- Future apps/binaries: `borg-web`, `borg-native`
- Current likely sources:
  - `apps/web`
  - `apps/maestro`
  - `apps/maestro-go`
  - `apps/mobile`
  - `packages/ui`

### Extraction rule

Keep shared contracts, config, auth, logging, and transport schemas in reusable packages first. Extract a new binary only after the package seam is clear enough that process separation improves reliability or operator clarity instead of just adding more moving parts.

### First extraction seams to prefer

If work proceeds incrementally, the first concrete seams should be:

1. `borgd`
   - pull top-level control-plane routing, operator health/status APIs, lock/config coordination, and provider-routing orchestration toward a cleaner daemon-owned boundary
   - keep CLI, web, and native surfaces as clients of that boundary
2. `borgmcpd`
   - pull MCP registry state, runtime-server lifecycle, working-set state, tool inventory/search/call mediation, and probe/test flows toward a dedicated service boundary
   - keep scrape/probe refresh and offline metadata enrichment as `hypermcp-indexer` worker responsibilities rather than interactive daemon logic

These seams are preferred first because they already have visible operator-facing surfaces, clear uptime concerns, and strong pressure to separate control-plane truth from client UX.

## Design principles

1. **Local first** — default to local state and operator control.
2. **Truth over hype** — label maturity honestly.
3. **Interoperability over reinvention** — unify tools where possible.
4. **Visibility over magic** — make system state inspectable.
5. **Continuity over novelty** — prioritize recovery, routing, and memory.

## Contributing

For now, compatibility paths, package names, and the `borg` CLI command remain unchanged while the visible branding shifts to borg.

Use `pnpm` v10 and verify changes before claiming success:

```bash
pnpm -C packages/core exec tsc --noEmit
pnpm -C apps/web exec tsc --noEmit --pretty false
pnpm run test
```

Also review:
- `AGENTS.md`
- `ROADMAP.md`
- `TODO.md`
- `VISION.md`

## Documentation map

- `VISION.md` — long-term direction
- `ROADMAP.md` — now/next/later
- `TODO.md` — active worklist
- `AGENTS.md` — contributor and agent rules
- `CHANGELOG.md` — release history

## License

MIT
