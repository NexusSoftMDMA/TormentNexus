# TormentNexus Autonomous Sales Pipeline Architecture

This system is an asynchronous, event-driven orchestration layer written in Go to automate B2B lead generation, enrichment, hyper-personalized outreach, and billing for the **TormentNexus AI Hypervisor**.

---

## THE PRODUCT: TormentNexus AI Hypervisor

> **Every interaction, every dossier, every outreach email this sales bot generates is about selling TormentNexus. Understanding the product deeply is non-negotiable.**

### What TormentNexus Is

TormentNexus is a **local-first cognitive control plane** — a Go+TypeScript modular monolith that coordinates multi-agent LLM workflows, Model Context Protocol (MCP) tool routing, provider failover, memory persistence, and operator observability. It is the **Operating System for AI models**.

In one sentence: **TormentNexus is the substrate where a single local system seamlessly coordinates the most critical parts of AI-driven software development: tools, models, sessions, context, subagents, and full visibility across the entire stack.**

TormentNexus is not just an aggregator; it is a **decision system and universal bridge**.

### Architecture: Go Modular Monolith + TypeScript Control Plane

TormentNexus runs as two cooperating processes:

- **Go Sidecar** (port 4300): The authoritative execution kernel — 232 Go files, 446 HTTP handler methods, 18K-line server. Handles orchestration, progressive MCP routing, L1/L2 memory management, and LLM waterfall routing.
- **TypeScript Control Plane** (port 4100): The observation deck — 583 TS files, 4343-line MCPServer.ts core. tRPC middleware + Next.js dashboard.
- **Next.js Dashboard** (port 3000): 91 pages, 709 TS/TSX files. Real-time operator observability.
- **SQLite Storage**: `tormentnexus.db` (35MB), 14,726 agent memories, 13,478 contexts, 11,024 populated MCP servers.

### The Five Core Pillars

#### 1. Progressive MCP Tool Routing & Parity

Models should never be overwhelmed with a 50,000-token tool dump. TormentNexus employs a multi-layered, progressive disclosure system:

- **Semantic Search:** Local vector embeddings match the active prompt against a global MCP directory.
- **The Router:** Only the top highly relevant tool schemas are injected into the active LLM context.
- **Universal Parity:** Byte-for-byte identical tool signatures for Claude Code, Codex, Gemini CLI, Cursor, Windsurf, Kiro, and GitHub Copilot CLI.

This is TormentNexus's **strongest selling point** — no other system provides deterministic cross-harness tool parity.

#### 2. Dual-Tier Memory Architecture (L1/L2)

Context is finite; memory must be infinite.

- **L1 - Session Scratchpad:** Ephemeral, lightning-fast memory tied directly to the active session.
- **L2 - The Vault:** Permanent semantic storage in SQLite with `sqlite-vec` for vector search. Saves exact transcripts and LLM-compressed heuristics.
- **Context Harvesting:** Every session autonomously queries the L2 Vault to pull in relevant historical heuristics.

14,726 memories currently stored, all surviving restarts.

#### 3. The Resilient LLM Waterfall

Uptime is non-negotiable. TormentNexus's inference client natively catches 429s (Rate Limits) and 5xx (Server Errors), seamlessly cascading the exact payload down a prioritized chain:

1. **NVIDIA NIM** / Primary APIs
2. **OpenRouter** (Secondary aggregator fallback)
3. **Local LM Studio / Ollama** (Ultimate offline fallback)

Provider catalog includes: Google, Anthropic, OpenAI, DeepSeek, OpenRouter, GitHub Copilot.

#### 4. Multi-Agent Swarm & P2P Mesh

TormentNexus coordinates specialized models inside shared chatrooms via the Agent-to-Agent (A2A) protocol:

- **Role Rotation:** Models take turns acting as Planner, Implementer, Tester, and Critic.
- **Consensus & Debate:** Agents autonomously bid on tasks, share context via a neural transcript, and debate implementations until consensus is reached.
- **PairOrchestrator:** Enforces the `Planner → Reviewer → Implementer → Reviewer → Critic` collaboration cycle.
- **Council/Debate:** Collaborative debate manager with rotation rooms and human veto service.

#### 5. Truth Over Hype Dashboards

TormentNexus's dashboards reflect actual SQLite database rows and active Go goroutine states. No mocked UI scaffolds. Monitor telemetry, traffic inspection, working-set capacity, and LLM routing histories in real-time.

### TormentNexus Technical Details (Sales-Relevant)

#### Runtime Ports

| Service | Port | Purpose |
|---|---|---|
| Next.js Dashboard | 3000 | Web observation deck |
| Socket.io | 3001 | Real-time swarm signals |
| tRPC Bridge | 4100 | TypeScript Control Plane API |
| TormentNexus Go Kernel | 4300 | Authoritative native sidecar |

#### Tech Stack

- **Go 1.26+** for the kernel sidecar (state, memory, routing, MCP sync, orchestration)
- **TypeScript 5.x** / Node.js 24+ / pnpm v10 for the control plane
- **Next.js 16 / React 19 / Tailwind CSS 4** for the dashboard
- **SQLite + sqlite-vec** for dependency-free, hyper-fast local vector search
- **tRPC** for type-safe internal API communication
- **SSE (Server-Sent Events)** for real-time event streaming from Go sidecar
- **JSON-RPC** for standard MCP server communication

#### Repository Structure

```
tormentnexus/
├─ apps/                    # Operator-facing applications (web, maestro, mobile)
│  ├─ web/                  #   Next.js dashboard (91 pages)
│  ├─ maestro/              #   Electron desktop shell
│  └─ cloud-orchestrator/   #   Nested cloud orchestrator sub-workspace
├─ packages/                # Shared libraries and TypeScript control plane
│  ├─ core/                 #   Main TS control plane, tRPC routers, services
│  ├─ cli/                  #   CLI entrypoint (31 commands)
│  ├─ ui/                   #   Shared React UI components
│  ├─ ai/                   #   Model/provider SDK integration layer
│  ├─ memory/               #   Memory storage, retrieval, embeddings
│  ├─ mcp-registry/         #   MCP metadata and registry
│  ├─ mcp-client/           #   MCP client integration
│  └─ tormentnexus-supervisor/ # Windows supervisor bridge
├─ go/                      # Go sidecar (experimental → stabilizing)
│  └─ internal/             #   35+ Go packages (ai, buffer, codeexec, config,
│                           #   controlplane, ctxharvester, eventbus, git,
│                           #   healer, httpapi, llm, memory, mcp, orchestration,
│                           #   providers, skillregistry, tools, vault, workflow...)
├─ docs/                    # Project documentation
├─ data/                    # Local knowledge assets, BobbyBookmarks
└─ tormentnexus.db          # SQLite database (35MB, 11K MCP servers)
```

#### Go Sidecar Internal Packages (35+)

| Package | Purpose |
|---|---|
| `ai` | AI provider abstractions |
| `buffer` | Stream/event buffering |
| `codeexec` | Multi-language code execution sandbox |
| `config` | Configuration management |
| `controlplane` | TS bridge communication |
| `ctxharvester` | Context harvesting & compaction |
| `eventbus` | High-frequency Swarm event broker |
| `git` / `gitservice` | Git operations and repository management |
| `healer` | Self-healing: error → diagnosis → fix → verify loop |
| `httpapi` | 446 HTTP handler methods |
| `llm` | LLM waterfall routing (NVIDIA → OpenRouter → LM Studio) |
| `memory` / `memorystore` | L1/L2 memory with sqlite-vec semantic search |
| `mcp` | MCP server management and progressive tool routing |
| `orchestration` | PairOrchestrator, Swarm, Council coordination |
| `providers` | Provider catalog (Google, Anthropic, OpenAI, DeepSeek, etc.) |
| `skillregistry` | SKILL.md discovery, CRUD, search |
| `tools` / `toolregistry` | Tool discovery, ranking, progressive disclosure |
| `vault` | Secure persistence for sessions, memories, secrets |

### Feature Maturity Matrix

#### FULLY OPERATIONAL (Stable)

| Feature | Details |
|---|---|
| **Progressive MCP Tool Routing** | 6 meta-tools, ranked search, auto-load, LRU eviction, profiles |
| **MCP Decision System** | Multi-signal scoring, silent high-confidence auto-load, deferred binary startup |
| **LLM Waterfall** | NVIDIA → OpenRouter → LM Studio/Ollama cascade on 429/5xx |
| **Provider Catalog** | Google, Anthropic, OpenAI, DeepSeek, OpenRouter, GitHub Copilot routing |
| **MCP Config Management** | CRUD for mcp.jsonc + DB tools, sync targets, export/import |
| **MCP Catalog Ingestion** | 5 adapters: Glama, Smithery, MCP.run, npm, GitHub Topics |
| **MCP Server Pool** | Multi-process supervision with PID tracking |
| **Session Import/Export** | Claude Code, Cursor, Aider, Windsurf, Copilot format detection |
| **CLI (31 commands)** | Full lifecycle for MCP, sessions, providers, knowledge, swarm |
| **Dashboard (91 pages)** | Health, inspector, tools, catalog, memory, swarm, council, etc. |

#### BUILT BUT IMMATURE (Beta / Experimental)

| Feature | Gap |
|---|---|
| **Tiered Memory (L1/L2)** | No heat-based promotion, no adaptive forgetting, no L3 cold archive |
| **Knowledge Graph** | GraphNode/GraphEdge interfaces only — actual implementations are `undefined` |
| **Context Harvester** | No semantic compaction via LLM, rudimentary grooming |
| **Healer (Self-Healing)** | No closed-loop execute-fix-verify-retry, no StopHook, no IdleHealer |
| **Swarm Controller** | No real consensus, no task bidding, no completion detection |
| **Pair Orchestrator** | Not wired to actual agent sessions |
| **Council/Debate** | Self-evolution only adjusts weights — no prompt/skill evolution |
| **Skill Registry** | No /evolve command, no win-rate tracking, no auto-retirement |
| **Darwin (Self-Modification)** | No integration with skills or memory, runs in isolation |
| **Code Executor** | WASM sandbox scaffolded but uses exec.Command fallback |

#### NOT BUILT (Vision / TODO)

Progressive Skill Disclosure, WASM Sandbox, Graph Memory + HITL Gates, Context Re-Injection, Token Budget Manager, Memory Consolidation, Browser Extension, Free-Tier Fallback Chain, OAuth Login, Cursor/Windsurf/Kiro Parity

### Built-In Tool Parity (Evidence Lock)

TormentNexus maintains byte-for-byte tool signature parity across all major AI coding harnesses. 27 golden fixtures, 6 L2 platforms:

| Platform | Tool Parity | Fixture Count |
|---|---|---|
| Claude Code | ✅ Ready for L3 lock | 3 |
| GitHub Copilot CLI | ✅ Ready for L3 lock | 4 |
| Codex CLI | ✅ Ready for L3 lock | 3 |
| Cursor | ✅ Ready for L3 lock | 3 |
| Gemini CLI | ✅ Ready for L3 lock | 2 |
| Kiro | ✅ Ready for L3 lock | 2 |

Tool equivalence examples:
- `shell_execution`: `bash()` (Copilot/Codex/Gemini), `Bash()` (Claude), `Shell()` (Cursor)
- `file_read`: `view()` (Copilot), `read` (Codex), `Read()` (Claude/Cursor), `file-read` (Gemini)
- `file_write`: `edit()/create()` (Copilot), `write` (Codex), `Edit()/Write()` (Claude/Cursor)

### Licensing

- **Community Edition**: Free personal use under BSL 1.1 / AGPLv3 (self-hosted)
- **Enterprise Core**: Commercial license with SSO/OIDC, Role-Based Access Control, audit trail, Ed25519 signed enterprise tokens

### BobbyBookmarks & Ecosystem Intelligence

TormentNexus integrates a 13,503-bookmark ecosystem intelligence database. Cross-system analysis with Atlas (7,944 entries) reveals:

- **4,664 shared repos** between TormentNexus and Atlas
- **11,024 populated MCP servers** in the TormentNexus catalog
- **Ecosystem gap analysis**: Verification (18% saturation), Self-Modification (15%), Memory (15%) are critically underbuilt across the entire AI engineering ecosystem — these are TormentNexus's differentiation opportunities

### Key Selling Points for Outreach

When crafting outreach, always anchor to these differentiators:

1. **Progressive Tool Disclosure** — "Your agents are drowning in 50K-token tool dumps. TormentNexus's semantic router injects only the 3 most relevant tools per request."
2. **Cross-Harness Parity** — "One config, identical tool signatures across Claude Code, Cursor, Codex, Gemini CLI, Copilot, and Windsurf. No vendor lock-in."
3. **LLM Waterfall** — "When OpenAI rate-limits you, TormentNexus cascades to OpenRouter, then to local Ollama. Zero downtime inference."
4. **Local-First Memory** — "14,726 memories that survive restarts. No cloud dependency. Your team's knowledge stays on your machines."
5. **Multi-Agent Swarm** — "Planner, Implementer, Tester, Critic — all in one chatroom with shared transcript and consensus-driven decisions."
6. **Self-Healing** — "When code breaks, TormentNexus diagnoses, fixes, and verifies autonomously. Every attempt persisted for fleet-wide learning."
7. **Truth Over Hype** — "Every dashboard widget shows real SQLite rows, real goroutine state. No mock data. No reassuring fiction."
8. **11K+ MCP Server Catalog** — "The largest indexed catalog of MCP servers, with semantic search and auto-discovery from 5 registries."
9. **Enterprise-Ready** — "RBAC, SSO/OIDC, Ed25519 signed tokens, audit trails. Self-hosted. No data leaves your network."

### Customer Pain Points TormentNexus Solves

| Pain Point | TormentNexus Solution |
|---|---|
| Agents overwhelmed by tool context | Progressive MCP routing with ranked search + LRU eviction |
| Vendor lock-in to single AI harness | Universal tool parity across 6+ harnesses |
| LLM API downtime kills productivity | 3-tier waterfall fallback (cloud → aggregator → local) |
| Lost context between sessions | L1/L2 memory with semantic search, 14K+ persisted memories |
| No visibility into agent decisions | Real-time dashboards with actual DB state, not mocks |
| Manual tool/server management | 5-adapter auto-ingestion, CRUD management, profile-based routing |
| Single-agent limitation | Swarm orchestration with Planner/Implementer/Tester/Critic rotation |
| Repetitive debugging cycles | Self-healing loop: diagnose → fix → verify → persist learnings |

### Competitive Landscape

TormentNexus occupies a unique position: **no other system connects memory, tool routing, provider failover, multi-agent coordination, and cross-harness parity in a single local-first control plane.**

The ecosystem data shows:
- Protocol/MCP layers: 66% saturated (overbuilt) — TormentNexus ROUTES existing servers, doesn't compete
- Agent Runtime: 52% saturated — TormentNexus's swarm/council is sufficient
- **Verification**: 18% saturation (underbuilt) — TormentNexus's healer differentiates
- **Self-Modification**: 15% saturation (underbuilt) — TormentNexus's Darwin + skill evolution differentiates
- **Memory**: 15% saturation (underbuilt) — TormentNexus's L1/L2 + sqlite-vec differentiates

### Roadmap (What's Coming)

| Phase | Focus | Key Features |
|---|---|---|
| Phase 6 (Current) | Enterprise Readiness | MCP client testing, enterprise wrapper integration, context optimization |
| Phase 7 | Memory Foundation | Heat-based tier promotion, L3 cold archive, LLM consolidation, memory→tool feedback loop |
| Phase 8 | Autonomy Loop | Closed-loop self-healing, planner-checker-revise loop, context re-injection |
| Phase 9 | Skill Intelligence | Progressive skill discovery, skill evolution with win-rate tracking, auto-retirement |
| Phase 10 | The Unique Value | Real knowledge graph, semantic entity extraction, graph-memory-informed HITL gates |

---

## Sales Bot Tech Stack

- **Language:** Go 1.24 using standard concurrency paradigms (goroutines, channels) for background workers.
- **Database:** PostgreSQL 13+ with strict relational schema and atomic state transitions.
- **External APIs:** GitHub (`go-github`), Stripe (`stripe-go`), generic REST CRM.
- **Core Pattern:** Multi-agent autonomous orchestrations, task workers, and state logging.

## Module Architecture

| Package | Purpose | Key Interfaces |
|---|---|---|
| `internal/scraper` | Lead discovery from job boards & GitHub | `LeadSource` |
| `internal/enrichment` | Contact enrichment (Apollo, Hunter) | `EnrichmentSource` |
| `internal/researcher` | Technical dossier building | `Crawler`, `DossierProcessor` |
| `internal/communication` | Inbound/outbound state machine | `IntentClassifier`, `ResponseGenerator`, `SalesStrategy`, `OrderProcessor` |
| `internal/crm` | Bidirectional CRM sync | `CRMClient` |
| `internal/billing` | Stripe invoicing & payment tracking | `BillingClient` |
| `internal/sales` | Order fulfillment for won deals | `OrderDB` |
| `internal/llm` | LLM provider abstraction | `LLMProvider` |
| `internal/autodev` | Autonomous code development | `Agent` |
| `internal/deploy` | CI tracking & deployment | `CITracker`, `WorkflowDispatcher` |
| `internal/gitcheck` | Git operations & PR management | `PRManager` |
| `internal/gitres` | Intelligent merge engine | — |
| `internal/db` | PostgreSQL data layer | — |
| `internal/web` | HTTP dashboard & API | — |
| `pkg/agents` | Target discovery worker | — |
| `pkg/config` | Safety guardrails | — |

## Extension Conventions

- All new worker engines or agent subclasses must implement the internal `Agent` interface (or module-specific equivalent).
- Add new background routines to `/pkg/agents/` or `/internal/`.
- Maintain state, run logs, and target histories inside the existing database configuration layer.
- Always include explicit mock testing endpoints and defensive execution loops.
- All external integrations must be abstracted behind Go interfaces for testability and swappability.

## System Guidelines

- **State Machine:** Enforce rigid, atomic state updates for all leads in the PostgreSQL database. The 7-state lifecycle is: `Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won / Closed_Lost`.
- **Integrations:** All scraper engines must utilize headless configuration profiles. External communication modules use abstract interfaces to allow mock testing.
- **Configuration:** All environment variables should be consolidated into a typed `Config` struct loaded at startup (see Phase 6 TODO).
- **Logging:** Migrate from `log.Printf` to structured JSON logging with `slog` (see Phase 6 TODO).
- **Error Handling:** External API calls must implement retry with exponential backoff and circuit breakers (see Phase 6 TODO).

## Database Schema Constraints

All data migrations must use strict relational mappings with full foreign key constraints tracking Companies -> Contacts -> Interactions -> Deals.

### Current Tables

| Table | Purpose | Key Columns |
|---|---|---|
| `companies` | Target organizations | `domain` (UNIQUE), `tech_stack[]`, `hiring_signals[]`, `market_cap_tier` |
| `contacts` | Decision-makers | `company_id` (FK), `email` (UNIQUE), `github_handle`, `linkedin_url` |
| `interactions` | Communication log | `contact_id` (FK), `channel`, `direction`, `success` (bool) |
| `deals` | Pipeline tracking | `company_id` (FK), `current_state` (enum), `quoted_pricing`, `technical_dossier` |
| `pull_requests` | AutoDev PR tracking | `id` (PK), `branch`, `status`, `task_description` |

### Known Schema Debt

- `contacts.email` UNIQUE constraint allows multiple NULLs — needs partial index or NOT NULL.
- Missing indices on `interactions.success` and `deals.current_state` for query performance.
- No `audit_log` table for state transition history.
- No `deleted_at` soft-delete columns for GDPR compliance.

## Autonomous Development & Repository Management Protocol

The system follows a strict "EXECUTIVE PROTOCOL" for repository synchronization and intelligent merging:

- **Upstream Tracking:** Always sync with the parent fork and update all submodules recursively.
- **Intelligent Merge:** Use the dual-direction merge engine to reconcile feature branches with `main`.
  - Forward merge: Feature → Main
  - Reverse merge: Main → Feature (prevents drift)
- **Validation:** Every build must pass the merge integrity tests defined in `internal/gitcheck`.
- **Automation:** Utilize `scripts/sync_repo.sh` for automated synchronization.
- **CI Gating:** AutoDev PRs are only merged after CI passes and staging validates successfully.
