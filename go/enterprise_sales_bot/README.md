🛠️ ALPHA SOFTWARE UNDER CONSTRUCTION — Use at your own risk. Backwards compatibility not guaranteed.

```text
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║                     ██╗   ██╗███╗   ██╗██████╗ ███████╗██████╗              ║
║                     ██║   ██║████╗  ██║██╔══██╗██╔════╝██╔══██╗             ║
║                     ██║   ██║██╔██╗ ██║██║  ██║█████╗  ██████╔╝             ║
║                     ██║   ██║██║╚██╗██║██║  ██║██╔══╝  ██╔══██╗             ║
║                     ╚██████╔╝██║ ╚████║██████╔╝███████╗██║  ██║             ║
║                      ╚═════╝ ╚═╝  ╚═══╝╚═════╝ ╚══════╝╚═╝  ╚═╝             ║
║                                                                              ║
║                     ██████╗ ██████╗ ███╗   ██╗███████╗████████╗██████╗      ║
║                    ██╔════╝██╔═══██╗████╗  ██║██╔════╝╚══██╔══╝██╔══██╗     ║
║                    ██║     ██║   ██║██╔██╗ ██║███████╗   ██║   ██████╔╝     ║
║                    ██║     ██║   ██║██║╚██╗██║╚════██║   ██║   ██╔══██╗     ║
║                    ╚██████╗╚██████╔╝██║ ╚████║███████║   ██║   ██║  ██║     ║
║                     ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═╝  ╚═╝     ║
║                                                                              ║
║                     █████╗ ██╗     ██████╗ ██╗  ██╗ █████╗                  ║
║                    ██╔══██╗██║     ██╔══██╗██║  ██║██╔══██╗                 ║
║                    ███████║██║     ██████╔╝███████║███████║                 ║
║                    ██╔══██║██║     ██╔═══╝ ██╔══██║██╔══██║                 ║
║                    ██║  ██║███████╗██║     ██║  ██║██║  ██║                 ║
║                    ╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝                 ║
║                                                                              ║
║                    ╔══════════════════════════════════════╗                  ║
║                    ║     ⚠️  ALPHA SOFTWARE  ⚠️           ║                  ║
║                    ║  EXPECT BREAKING CHANGES & BUGS     ║                  ║
║                    ║  NOT READY FOR PRODUCTION USE       ║                  ║
║                    ╚══════════════════════════════════════╝                  ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

# TormentNexus Autonomous Sales Pipeline

A fully autonomous B2B sales pipeline written in Go. It discovers potential enterprise customers, researches their technical bottlenecks, sends hyper-personalized outreach emails, negotiates deals, invoices won deals via Stripe, and even **modifies its own source code** to improve itself. It runs without human intervention — a software salesperson that never sleeps, writes its own PRs, and learns from its successes.

### ▶️ Video Overviews

| Video | Description | Direct Download |
|---|---|---|
| [**HyperNexus: Ultimate Agent**](https://notebooklm.google.com/notebook/0a540934-3f43-4c52-91e0-ebc622071409/artifact/12652b93-e2f0-4b35-bc6a-31c2c8fa51f0?utm_source=nlm_web_share&utm_medium=google_oo&utm_campaign=art_share_1&utm_content=&utm_smc=nlm_web_share_google_oo_art_share_1_) | AI-generated deep-dive into the HyperNexus enterprise AI orchestration platform, architecture, and ultimate agent capabilities (182 sources) | [MP4 →](https://hypernexus.site/hypernexus_ultimate_agent.mp4) |
| [**TormentNexus: AI Control Plane**](https://notebooklm.google.com/notebook/0a540934-3f43-4c52-91e0-ebc622071409/artifact/282cd73a-c108-4fc7-bf6e-1dcdac446a54?utm_source=nlm_web_share&utm_medium=google_oo&utm_campaign=art_share_1&utm_content=&utm_smc=nlm_web_share_google_oo_art_share_1_) | AI-generated overview of the TormentNexus AI Control Plane — the local-first model hypervisor with MCP routing, memory, and autonomous capabilities (182 sources) | [MP4 →](https://tormentnexus.site/tormentnexus_ai_control_plane.mp4) |

## Table of Contents

- [What It Does](#what-it-does)
- [Architecture](#architecture)
- [The 7-State Lead Lifecycle](#the-7-state-lead-lifecycle)
- [Module-by-Module Breakdown](#module-by-module-breakdown)
- [The Self-Improving Loop](#the-self-improving-loop)
- [Integration Status](#integration-status)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Database & Migrations](#database--migrations)
- [Testing](#testing)
- [Repository Management — EXECUTIVE PROTOCOL](#repository-management--executive-protocol)
- [CI / Validation](#ci--validation)
- [Known Issues](#known-issues)
- [Contributing](#contributing)

---

## What It Does

In plain English, this is a Go program that:

1. **Finds companies** that might need an AI orchestration product — by scanning GitHub for MCP servers and job boards for hiring signals
2. **Finds decision-makers** at those companies — via Apollo/enrichment APIs with name, role, email, and GitHub handle
3. **Stalks their GitHub repos and blogs** to find technical pain points — like serial processing bottlenecks in orchestration logic
4. **Sends them personalized emails** that reference their specific bottlenecks — grounded in real TormentNexus documentation
5. **Handles their replies autonomously** — answering technical questions, quoting pricing ($5K–$50K/yr based on company size), handling objections (one rebuttal, then escalate to human)
6. **Closes deals** when the lead is qualified enough — creating real Stripe invoices with 30-day payment terms
7. **Syncs everything** to an external CRM bidirectionally — with retry logic and exponential backoff
8. **Reads its own TODO list and implements features** — by writing code, creating PRs, and auto-merging them after CI passes
9. **Manages its own git repository** — syncing, reconciling branches, and resolving merge conflicts via the Dual-Direction Intelligent Merge Engine
10. **Serves a web dashboard** where a human can watch all of this happen, manually trigger actions, and monitor performance metrics

---

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                              main.go                                  │
│                                                                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────────────┐   │
│  │ Scraper  │  │ Enricher │  │Researcher│  │   Communication    │   │
│  │ (1h tick)│  │ (1h tick)│  │ (1h tick)│  │     Manager        │   │
│  │          │  │          │  │          │  │   (30m tick)       │   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────────┬───────────┘   │
│       │              │              │                 │               │
│  ┌────▼──────────────▼──────────────▼─────────────────▼──────────┐   │
│  │                        PostgreSQL                              │   │
│  │    companies → contacts → interactions → deals                │   │
│  │                       pull_requests                           │   │
│  └───────────────────────────────────────────────────────────────┘   │
│                                                                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────────────────────┐   │
│  │ CRM Sync │  │ AutoDev  │  │       Web Dashboard :8080        │   │
│  │  (30m)   │  │ (1h tick)│  │ /login  /  /health  /webhook    │   │
│  └──────────┘  └──────────┘  └──────────────────────────────────┘   │
│                                                                       │
│  ┌─────────────────────┐  ┌──────────────────────────────────────┐  │
│  │ Deploy Worker (cfg) │  │ Target Discovery Worker (2h)         │  │
│  └─────────────────────┘  └──────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────┘
```

### Tech Stack

- **Language:** Go 1.24 using standard concurrency (goroutines, channels)
- **Database:** PostgreSQL 13+ with strict relational schema and atomic state transitions
- **External APIs:** GitHub (`go-github`), Stripe (`stripe-go`), generic REST CRM
- **Core Pattern:** Multi-agent autonomous orchestration, background task workers, state logging

### Module Architecture

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
| `internal/auth` | Session-based dashboard auth | — |
| `internal/config` | Centralized environment config | — |
| `pkg/agents` | Target discovery worker | — |
| `pkg/config` | Safety guardrails | — |

### Background Workers

| Worker | Package | Interval | Purpose |
|---|---|---|---|
| Scraper | `internal/scraper` | 1h | Discover new leads from job boards & GitHub |
| Enricher | `internal/enrichment` | 1h | Enrich companies with contact data |
| Researcher | `internal/researcher` | 1h | Build technical dossiers via crawling |
| CRM Sync | `internal/crm` | 30m | Bidirectional CRM reconciliation |
| Target Discovery | `pkg/agents` | 2h | GitHub MCP server discovery |
| Communication | `internal/communication` | 30m | Process inbound + trigger outbound |
| AutoDev | `internal/autodev` | 1h | Self-code, PR, and merge cycle |
| Deploy Sync | `internal/deploy` | Configurable | Background repo synchronization |
| Deploy Monitor | `internal/deploy` | Configurable | Health check monitoring |

---

## The 7-State Lead Lifecycle

Every lead goes through a rigid pipeline enforced by a PostgreSQL `ENUM` type. No lead can skip states — all transitions are atomic database updates.

```
Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won
                                                                     ↘ Closed_Lost
```

| State | Meaning | What triggers it |
|---|---|---|
| `Discovered` | Company identified, no contacts yet | Scraper finds a new company |
| `Researched` | Decision-maker contacts found + technical dossier compiled | Enricher finds contacts |
| `Outreach_Sent` | First personalized email sent | Communication Manager sends initial outreach |
| `Engaged` | Prospect replied | Inbound message received from contact |
| `Negotiating` | Active deal discussion | 3+ interactions or qualification score >70 |
| `Closed_Won` | Deal won, invoice created | Qualification ≥80 + FollowUp intent → auto-close |
| `Closed_Lost` | Deal lost | Escalation or manual closure |

---

## Module-by-Module Breakdown

### 1. Scraper (`internal/scraper`) — Lead Discovery

Scans job boards and GitHub for companies hiring for AI orchestration roles. Creates a `Company` record and a `Deal` in `Discovered` state.

- **Interface:** `LeadSource.Discover(ctx, keywords) → []Company`
- **`MockJobBoardSource`** — returns hardcoded companies (AI Dynamics Corp, Neural Systems Inc)
- **`GitHubJobSource`** — returns verified high-value targets
- Searches for keywords: "AI Engineer", "LLM Orchestration", "Agentic Workflows"

### 2. Enricher (`internal/enrichment`) — Contact Discovery

Finds decision-makers (name, role, email, GitHub handle) at discovered companies. Advances deal from `Discovered` → `Researched`. Syncs found contacts to CRM with 3-attempt retry and exponential backoff.

- **Interface:** `EnrichmentSource.Enrich(ctx, company) → []Contact`
- **`MockApolloSource`** — returns hardcoded contacts per domain

### 3. Researcher (`internal/researcher`) — Technical Dossier Building

Crawls GitHub repos and engineering blogs for each contact's handle. Identifies technical bottlenecks (e.g., "high-latency serial state updates"). Builds a `technical_dossier` and stores it on the Deal. Pushes updated dossier to CRM.

- **Interface:** `Crawler.Crawl(ctx, target) → string` + `DossierProcessor.Process(findings) → string`
- **`GitHubCrawler`** — makes real GitHub API calls if `GITHUB_TOKEN` is set, otherwise falls back to intelligent simulated findings containing "BOTTLENECK DETECTED"
- **`BlogCrawler`** — simulated (returns hypothetical blog insights about state management)
- **`PromptFormatter`** — constructs outreach prompts combining TormentNexus docs + dossier

### 4. Communication Manager (`internal/communication`) — The Sales Brain

This is the most sophisticated module. Four sub-components work together:

#### a. Intent Classifier (`classifier.go`)

Classifies inbound messages into: `Technical`, `Pricing`, `Objection`, `MeetingRequest`, `FollowUp`, `Spam`, `Unknown`

- **`MockIntentClassifier`** — keyword heuristic matching (e.g., "pricing"/"cost" → Pricing intent)
- **`LLMIntentClassifier`** — sends text to an LLM for classification (exists but not active by default)

#### b. RAG Response Generator (`responder.go`)

Generates hyper-personalized replies using three context sources:

1. **TormentNexus documentation** loaded from `borg/docs/ARCHITECTURE.md` (pseudo-RAG) — injected for technical intent
2. **Pricing context** from `CalculateQuote()` — Enterprise=$50K, Mid-Market=$15K, SMB=$5K
3. **Self-Improving Prompts loop** — injects successful past interactions as few-shot examples

#### c. Learning Sales Engine (`engine.go`) — The Decision Engine

- **`ScoreLead()`** — Scores 0–100 based on market cap tier (Enterprise=+50), dossier insights (BOTTLENECK=+30, INFRASTRUCTURE=+20), and interaction count
- **`QualifyLead()`** — Scores 0–100 based on lead score + engagement quality + intent signals (MeetingRequest=+25, Pricing=+15, FollowUp=+20, Objection=−10)
- **`Decide()`** — The core decision loop:
  - Qualified ≥80 + FollowUp intent → auto-close as `Closed_Won`
  - Engaged with 3+ interactions or qualification >70 → advance to `Negotiating`
  - Technical intent → `ActionRespond`
  - Pricing from high-value leads → `ActionRespond`; low-value → `ActionEscalate`
  - Objection → one autonomous rebuttal, then escalate
  - Spam → `ActionWait`
- **`RouteLead()`** — Determines internal routing: Lead Solutions Architect, Senior Account Executive, Technical Sales Engineer, or Standard Sales Representative

#### d. Manager (`manager.go`) — The Orchestration Loop

- Polls for `Researched` deals with no outbound interaction yet
- Auto-initiates outreach (sends "START_OUTREACH" to the pipeline)
- Processes inbound messages through classify → decide → respond
- **When a deal is won:** marks all past outbound interactions as `success=true` (feeding the Self-Improving Prompts loop), triggers `OrderProcessor`

### 5. Order Processor (`internal/sales`) — Deal Fulfillment

Triggered when a deal reaches `Closed_Won`:

1. Creates a Stripe invoice via `BillingClient`
2. Syncs the order to CRM
3. **Real Stripe integration:** `StripeBillingClient` uses `stripe-go` v81 to create invoices with 30-day payment terms

### 6. AutoDev Orchestrator (`internal/autodev`) — Self-Modifying Code

The bot **modifies its own source code** through this lifecycle:

1. **TaskManager** parses `TODO.md` for unchecked `- [ ]` items, prioritizes `[HIGH]` tasks
2. **LocalAgent** generates code proposals, writes files (with path traversal protection), runs `go build` and `go test` as verification
3. **Orchestrator loop** (every 1 hour):
   - Execute Executive Sync Protocol (git fetch, submodule update, branch reconciliation)
   - Check if working directory is clean
   - Get next task from `TODO.md`
   - `ProposeSolution()` → generates code
   - `ApplyChanges()` → writes files to disk
   - `Verify()` → runs `go build ./...` and `go test ./...`
   - `MarkCompleted()` → updates `TODO.md`
   - Bumps `VERSION` with build timestamp
   - Creates feature branch, commits, pushes
   - Creates GitHub PR via `PRManager`
   - **On next cycle:** checks if PR CI passed → auto-merges → cleans up branch

### 7. Deploy Worker (`internal/deploy`) — CI/CD & Self-Deployment

- **`CITracker`** — checks GitHub Actions CI status for branches
- **`WorkflowDispatcher`** — triggers GitHub Actions workflows remotely
- **`Deployer`** — periodic sync (git pull + submodule update), periodic health monitoring
- If GitHub webhook receives a push event → triggers sync + build

### 8. Git Operations (`internal/gitcheck` + `internal/gitres`)

- **`gitcheck`** — `IsClean`, `IsSynced`, `SyncRemote`, `UpdateSubmodules`, `CheckoutAndCommit`, `PushBranch`, `DeleteBranch`, `CheckConflicts`, `GenerateSubmoduleInventory`
- **`gitres`** — **Dual-Direction Intelligent Merge Engine**:
  - Forward-merges feature branches into main
  - Reverse-merges main back into features to prevent drift
  - Then pushes reconciled main to origin

### 9. CRM Sync (`internal/crm`)

Bidirectional reconciliation: pulls lead updates from CRM, pushes negotiating/closed deals to CRM.

- **`RestCRMClient`** — real REST API client with Bearer auth, 6 methods (PushDeal, GetLeadUpdates, ValidateAccount, SyncInteraction, SyncContacts, FetchDealDetails)
- **`MockCRMClient`** — for testing
- All CRM calls include 3-attempt retry with exponential backoff

### 10. Target Discovery (`pkg/agents`) — GitHub Scanning

**Real GitHub API integration** using `go-github`:

- Searches GitHub for repositories matching `"model-context-protocol OR mcp-server language:Go language:TypeScript"`
- Creates companies from discovered repos with their language as tech stack
- Deduplicates against existing database entries

### 11. Web Dashboard (`internal/web`)

Session-based authentication (password from `ADMIN_PASSWORD` env var, default "admin"). Serves:

- **`/`** — HTML dashboard: recent deals table, performance metrics (total leads, win rate, outreach success), autonomous task board from `TODO.md`, active PRs, self-service deployment buttons (Sync/Build), system health
- **`/health`** → `"OK"`
- **`/health/detailed`** → JSON with DB connectivity + worker liveness
- **`/api/v1/webhook/github`** — GitHub webhook endpoint with HMAC-SHA256 signature verification
- **`/login`** — session authentication form

### 12. LLM Abstraction (`internal/llm`)

`LLMProvider` interface with `Generate(ctx, Prompt) → string`. Currently only `MockLLMProvider` returns `[MOCK LLM RESPONSE based on: ...]`. The `Prompt` struct carries `System`, `User`, and `MaxTokens` fields — ready for a real provider (OpenAI, Anthropic) to be plugged in.

### 13. Config (`internal/config`) — Centralized Environment Config

Typed `Config` struct loaded from env vars at startup with validation. Eliminates scattered `os.Getenv()` calls.

### 14. Safety (`pkg/config`) — Guardrails

`SafetyConfig` with `MaxDailyPRs=5`, tone constraint "Helpful Peer (Developer-to-Developer)", and opt-out disclaimer "Automated optimization discovery. Reply 'opt-out' to blacklist."

---

## The Self-Improving Loop

The system's most interesting architectural feature — a positive feedback loop that improves outreach quality over time:

```
  Deal reaches Closed_Won
          │
          ▼
  Communication Manager marks all past
  OUTBOUND interactions as success=true
          │
          ▼
  RAGResponseGenerator queries
  ListSuccessfulInteractions(limit=3)
          │
          ▼
  Successful responses injected into
  LLM prompt as few-shot examples
          │
          ▼
  Future outreach shaped by
  what actually worked
```

---

## Integration Status

| Integration | Status | Implementation |
|---|---|---|
| GitHub API (target discovery) | ✅ Real | `pkg/agents/discovery.go` with `go-github` |
| GitHub API (CI tracking) | ✅ Real | `internal/deploy/github_tracker.go` |
| GitHub API (PR management) | ✅ Real | `internal/gitcheck/pr.go` with `go-github` |
| Stripe billing | ✅ Real | `internal/billing/billing.go` with `stripe-go` |
| REST CRM client | ✅ Real | `internal/crm/crm.go` with generic REST |
| Intent classifier | ⚠️ Hybrid | `MockIntentClassifier` + `LLMIntentClassifier` (exists but mock is default) |
| LLM provider | ❌ Mock | `internal/llm/llm.go::MockLLMProvider` |
| Enrichment (Apollo) | ❌ Mock | `internal/enrichment/worker.go::MockApolloSource` |
| Job board scraper | ❌ Mock | `internal/scraper/scraper.go::MockJobBoardSource` |
| Email sending | ❌ Not implemented | Outbound is logged but not sent |
| Email receiving | ❌ Not implemented | Inbound is simulated by polling DB |

---

## Getting Started

### Prerequisites

- **Go:** version 1.24 or later
- **PostgreSQL:** version 13 or later
- **Git:** for version control and submodule management
- **GitHub Token:** A Personal Access Token (PAT) with `repo` permissions (recommended)

### Quick Start (Docker)

```bash
docker compose up --build
```

Dashboard at `http://localhost:8080`.

### Quick Start (Local)

1. **Set environment variables:**

   ```bash
   export DATABASE_URL="postgres://user:pass@localhost:5432/sales_bot?sslmode=disable"
   export GITHUB_TOKEN="ghp_xxxx"           # optional, enables real GitHub integration
   export GITHUB_REPOSITORY="owner/repo"     # optional, enables CI tracking and AutoDev
   export DEPLOY_SYNC_INTERVAL="1h"          # optional, enables background sync
   ```

2. **Apply migrations:**

   ```bash
   migrate -path migrations/ -database "$DATABASE_URL" up
   ```

3. **Initialize submodules:**

   ```bash
   git submodule update --init --recursive
   ```

4. **Build and run:**

   ```bash
   go build -o bin/sales_bot ./cmd/sales_bot
   ./bin/sales_bot
   ```

   Or use the provided scripts:

   ```batch
   build.bat
   start.bat
   ```

### Command-Line Flags

| Flag | Description |
|---|---|
| `--reconcile` | Run branch reconciliation and exit |
| `--inventory` | Generate submodule inventory table and exit |

---

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | `postgres://postgres:postgres@localhost:5432/sales_bot?sslmode=disable` | PostgreSQL connection string |
| `GITHUB_TOKEN` | No | — | GitHub PAT for API access (enrichment, CI, PRs) |
| `GITHUB_REPOSITORY` | No | — | `owner/repo` for CI tracking and AutoDev |
| `GITHUB_WEBHOOK_SECRET` | No | — | HMAC secret for webhook verification |
| `CRM_BASE_URL` | No | — | REST CRM API base URL (enables real CRM) |
| `CRM_API_KEY` | No | — | REST CRM API key |
| `DEPLOY_SYNC_INTERVAL` | No | `1h` | Duration string (e.g., `1h`, `15m`) for background sync |
| `ADMIN_PASSWORD` | No | `admin` | Dashboard login password |
| `PORT` | No | `8080` | HTTP dashboard port |
| `ENVIRONMENT` | No | `development` | Runtime environment label |
| `GO_TEST_MODE` | No | — | Set to `true` to skip git operations in tests |
| `SKIP_AUTODEV_SYNC` | No | — | Set to `true` to skip git sync in AutoDev |
| `SKIP_AUTODEV_TESTS` | No | — | Set to `true` to skip test verification in AutoDev |

---

## Database & Migrations

All schema changes go through the `migrations/` directory. Relational integrity enforced with explicit foreign keys:

```
companies(id) → contacts(company_id) → interactions(contact_id)
companies(id) → deals(company_id)
                pull_requests(id)
```

### Current Schema (4 migrations)

| Migration | Description |
|---|---|
| `000001` | Initial schema: companies, contacts, interactions, deals + `lead_state` enum + `updated_at` triggers |
| `000002` | Add `technical_dossier` column to deals |
| `000003` | Create `pull_requests` table for AutoDev tracking |
| `000004` | Add `success` boolean to interactions (prompt optimization loop) |

### Known Schema Debt

- `contacts.email` UNIQUE constraint allows multiple NULLs — needs partial index or NOT NULL
- Missing indices on `interactions.success` and `deals.current_state` for query performance
- No `audit_log` table for state transition history
- No `deleted_at` soft-delete columns for GDPR compliance

---

## Testing

```bash
# Unit tests
go test ./... -v

# Integration tests (requires DATABASE_URL)
DATABASE_URL=postgres://user:pass@localhost:5432/sales_bot go test ./... -v -tags=integration

# E2E tests
DATABASE_URL=postgres://user:pass@localhost:5432/sales_bot go test ./tests/e2e/... -v

# Smoke test against a running instance
TARGET_URL="https://your-instance.com" go run scripts/smoke_test.go
```

---

## Repository Management — EXECUTIVE PROTOCOL

The system follows a strict protocol for repository synchronization and intelligent merging:

- **Upstream Tracking:** Always sync with the parent fork and update submodules recursively.
- **Intelligent Merge:** Use the dual-direction merge engine to reconcile feature branches with `main`.
  - Forward merge: Feature → Main
  - Reverse merge: Main → Feature (prevents drift)
- **Validation:** Every build must pass merge integrity tests in `internal/gitcheck`.
- **Automation:** Use `scripts/sync_repo.sh` for automated synchronization.
- **CI Gating:** AutoDev PRs are only merged after CI passes and staging validates successfully.

---

## CI / Validation

The CI pipeline (`.github/workflows/ci.yml`) runs:

- Version consistency check (`VERSION` vs `VERSION.md`)
- Integrity tests (`internal/gitcheck`)
- Conflict resolution tests (`internal/gitres`)
- Full project test suite (`go test ./...`)
- Build verification (`go build ./cmd/sales_bot`)

---

## Known Issues

- **CRLF Test Failure:** `TestResolveConflictTheirs` in `internal/gitres/resolve_test.go` fails on Windows due to `\r\n` vs `\n` line ending mismatch. Does not affect production functionality.

---

## Contributing

- Follow the coding and branch conventions documented in `AGENTS.md`.
- All PRs must:
  - Target a feature branch and include migration files (if DB changes).
  - Include tests (unit + integration where applicable).
  - Pass `internal/gitcheck` and CI.
  - Include a clear description of agent interfaces added/changed.

---

## Useful Scripts & Tools

| Script | Purpose |
|---|---|
| `scripts/sync_repo.sh` | Upstream sync and submodule updater (per EXECUTIVE PROTOCOL) |
| `scripts/smoke_test.go` | Production health verification |
| `scripts/crm_verify/` | CRM API interaction verification |
| `build.bat` / `start.bat` | Windows build and start scripts |
| `--reconcile` flag | Run branch reconciliation standalone |
| `--inventory` flag | Generate submodule inventory table |

---

### ▶️ Video Dossier — Evidence File

AI-generated video overviews based on 182 source documents each:

| Video | Link | Direct Download |
|---|---|---|
| **HyperNexus: Ultimate Agent** | [Watch on NotebookLM →](https://notebooklm.google.com/notebook/0a540934-3f43-4c52-91e0-ebc622071409/artifact/12652b93-e2f0-4b35-bc6a-31c2c8fa51f0?utm_source=nlm_web_share&utm_medium=google_oo&utm_campaign=art_share_1&utm_content=&utm_smc=nlm_web_share_google_oo_art_share_1_) | [Download MP4 (19.9MB) →](https://hypernexus.site/hypernexus_ultimate_agent.mp4) |
| **TormentNexus: AI Control Plane** | [Watch on NotebookLM →](https://notebooklm.google.com/notebook/0a540934-3f43-4c52-91e0-ebc622071409/artifact/282cd73a-c108-4fc7-bf6e-1dcdac446a54?utm_source=nlm_web_share&utm_medium=google_oo&utm_campaign=art_share_1&utm_content=&utm_smc=nlm_web_share_google_oo_art_share_1_) | [Download MP4 (19.9MB) →](https://tormentnexus.site/tormentnexus_ai_control_plane.mp4) |

---

## License & Contact

- Maintainer: robertpelloni
- For security issues, create an issue in this repo and mark it as high-priority.
