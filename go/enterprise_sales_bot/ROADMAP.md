# Roadmap

[Output]

ipe)
- [x] Automated Provisioning
- [x] Autonomous Development Module (Functional)
- [x] Executive Sync Protocol Integration
- [x] Autonomous PR Handling & Merging (Persistent)
- [x] Automated CI/CD for Codebase Updates
- [x] Real-time CI Status Monitoring & Merging Guardrails
- [x] Self-Service Deployment Pipeline (UI & Backend)
- [x] Deployment Health Monitoring

---

## COMPLETED: Phase 6 — Production Hardening & Reliability

### 6.1 Test Coverage & Quality
- [x] **Fix CRLF line-ending test failure** in `internal/gitres/resolve_test.go` (`TestResolveConflictTheirs`)
- [x] **Add connection pool configuration** to `db.NewDB()` (max open/idle conns, lifetime)
- [x] **Add graceful shutdown** with drain timeouts for all background workers (scraper, enricher, researcher, CRM, comm, autodev, deploy)
- [x] **Add integration tests with ephemeral DB** for `enrichment/worker`, `researcher`, `crm/worker`, and `communication/manager` (currently only `db` and `e2e` have integration tests)
- [x] **Add web dashboard handler tests** for `/` route, webhook endpoint, and form actions (currently only `/health` is tested)
- [x] **Add negative/error-path unit tests** for `db/repository.go` (currently only happy-path integration tests exist)
- [x] **Add test coverage reporting** to CI pipeline (e.g., `go test -coverprofile` + artifact upload)

### 6.2 Database & Data Integrity
- [x] **Add `contacts.email` NULL constraint fix** — currently `UNIQUE` allows multiple NULL emails which may cause duplicate contact issues
- [x] **Add `interactions.success` index** for efficient `ListSuccessfulInteractions` queries
- [x] **Add `deals(current_state)` index** for efficient `ListDealsByState` queries
- [x] **Add migration for `audit_log` table** to track state transitions with who/when/why metadata
- [x] **Add `deleted_at` soft-delete column** to companies, contacts, and deals for GDPR compliance
- [x] **Add database migration runner** to application startup (auto-apply pending migrations on boot)

### 6.3 Configuration & Environment
- [x] **Add structured configuration** — replace scattered `os.Getenv()` calls with a typed config struct (e.g., `Config{DatabaseURL, GitHubToken, CRMAPIKey, SyncInterval, ...}`)
- [x] **Add `.env` file loading** via `godotenv` or equivalent for local development
- [x] **Add config validation at startup** — fail fast with clear error messages if required env vars are missing
- [x] **Add configurable worker intervals** — currently hardcoded in `main.go`, should be overridable via env vars

### 6.4 Logging & Observability
- [x] **Add structured JSON logging** — replace all `log.Printf` with a leveled, structured logger (e.g., `slog` or `zerolog`)
- [x] **Add Prometheus metrics endpoint** (`/metrics`) — expose counters for leads discovered, interactions processed, deals won/lost, PRs merged, etc.
- [x] **Add correlation/request IDs** to all log lines for traceability
- [x] **Add pprof endpoint** for production debugging (`/debug/pprof/`)

### 6.5 Error Handling & Resilience
- [x] **Add retry with exponential backoff** to all external API calls (GitHub, CRM, Stripe)
- [x] **Add circuit breaker** for external integrations (CRM, Stripe, GitHub API)
- [x] **Add dead-letter tracking** — persist failed interactions/updates for manual review instead of silently dropping errors
- [x] **Add health check dependencies** — detailed health endpoint should report each worker's last successful run timestamp

---

## Phase 7 — Real Integrations & Multi-Channel Outreach

### 7.1 Real Enrichment Providers
- [x] **Implement Apollo.io API enrichment source** (replace mock)
- [x] **Implement Hunter.io email finder** as secondary enrichment source
- [x] **Implement LinkedIn Sales Navigator scraper** for contact discovery (headless)
- [x] **Add enrichment source fallback chain** — if primary fails, try secondary automatically

### 7.2 Real Communication Channels
- [x] **Implement SMTP email sender** for outbound outreach (replace mock)
- [x] **Implement IMAP/POP3 email polling** for inbound message ingestion
- [x] **Implement LinkedIn message sending** via headless automation
- [x] **Implement GitHub Issue/PR comment outreach** as a technical hook channel
- [x] **Implement dual-brand social media posting worker agent** (Reddit, Bluesky, LinkedIn, Twitter/X)
- [x] **Add channel preference logic** — route outreach via the contact's preferred channel
- [x] **Add outreach cadence management** — configurable follow-up schedule (e.g., Day 1 → Day 3 → Day 7)

### 7.3 Real LLM Integration
- [x] **Implement OpenAI/Anthropic LLM provider** (replace mock)
- [x] **Add provider fallback chain** for LLM calls (primary → secondary → tertiary)
- [x] **Add token budget tracking** per deal/contact to control costs
- [x] **Add prompt versioning** — store and track prompt templates with A/B testing capability
- [x] **Add response quality scoring** — auto-evaluate generated responses before sending

### 7.4 Real CRM Integration
- [x] **Implement Salesforce CRM adapter** (replace generic REST mock)
- [x] **Implement HubSpot CRM adapter** as alternative
- [x] **Add CRM field mapping configuration** — map local fields to CRM-specific schema

---

## Phase 8 — Intelligence & Autonomous Evolution

### 8.1 Advanced Lead Intelligence
- [x] **Implement real GitHub repository analysis** — detect tech stack, architecture patterns, and bottlenecks from actual source code
- [x] **Implement real technical blog/RSS ingestion** — parse engineering blogs for hiring signals and pain points
- [x] **Add competitor intelligence** — track when targets evaluate or adopt competing solutions
- [x] **Add intent signal aggregation** — combine hiring signals, GitHub activity, blog posts, and job postings into a unified intent score

### 8.2 Autonomous Development Improvements
- [x] **Replace hardcoded `LocalAgent.ProposeSolution`** with LLM-powered code generation
- [x] **Add rollback mechanism** — if verification fails, revert to pre-change state
- [x] **Add PR feedback loop** — use `GetPRComments` to refine the agent's code generation accuracy
- [x] **Add task dependency resolution** — respect ordering between tasks (e.g., DB migration before feature code)
- [x] **Add concurrent task execution** — process independent tasks in parallel goroutines

### 8.3 Advanced Sales Strategy
- [x] **Add multi-touch outreach sequences** — define and execute cadenced sequences across channels
- [x] **Add A/B testing for outreach templates** — track conversion per template variant
- [x] **Add objection handling library** — curated rebuttals indexed by objection type with success rates
- [x] **Add human-in-the-loop approval workflow** — require explicit approval for deals above a configurable threshold
- [x] **Add deal forecasting** — predict close probability and expected revenue using historical patterns

### 8.4 Self-Improving Prompts v2
- [ ] **Add A/B prompt testing** — compare outreach generated with vs. without successful examples
- [x] **Add interaction sentiment analysis** — auto-classify sentiment of inbound messages to refine strategy
- [ ] **Add prompt performance tracking** — measure response quality over time as few-shot examples accumulate
- [ ] **Add negative example injection** — learn from failed outreach (flagged `success=false`) to avoid repeated patterns

---

## Phase 9 — Security, Compliance & Scale

### 9.1 Security
- [ ] **Add rate limiting** on all HTTP endpoints (dashboard, webhook, health)
- [x] **Add authentication** to the web dashboard (OAuth2 or API key)
- [ ] **Add CSRF protection** for dashboard form submissions
- [ ] **Add input sanitization** for all user-supplied data (webhook payloads, form inputs)
- [ ] **Add secrets encryption at rest** — encrypt GITHUB_TOKEN and API keys in config/storage
- [ ] **Add GDPR data export endpoint** — `/api/v1/export/{company_id}` for right-to-portability
- [ ] **Add GDPR data deletion endpoint** — `/api/v1/delete/{company_id}` for right-to-erasure
- [ ] **Add webhook IP allowlisting** — restrict GitHub webhook processing to known GitHub IPs

### 9.2 Scale & Performance
- [ ] **Add PostgreSQL connection pooling** with configurable limits
- [ ] **Add Redis caching layer** for frequently accessed data (company lookups, performance metrics)
- [ ] **Add horizontal scaling support** — make workers stateless so multiple instances can run
- [ ] **Add message queue** (NATS/RabbitMQ) to decouple workers from direct DB polling
- [ ] **Add database read replicas** for dashboard queries to reduce load on primary
- [ ] **Add pagination** to dashboard deal list (currently hardcoded to 20)
- [ ] **Add worker performance profiling** — track and alert on slow processing cycles

### 9.3 Deployment & Operations
- [ ] **Add Kubernetes manifests** (Deployment, Service, ConfigMap, Secret)
- [ ] **Add Helm chart** for one-command cluster deployment
- [ ] **Add Terraform modules** for cloud infrastructure provisioning (AWS RDS, EKS, etc.)
- [ ] **Add blue-green deployment** strategy with automatic rollback on health check failure
- [ ] **Add database backup automation** — periodic `pg_dump` with S3/blob storage upload
- [ ] **Add log aggregation** — ship structured logs to ELK/Datadog/CloudWatch

---

## Phase 10 — Platform & Ecosystem

### 10.1 API & Extensibility
- [x] **Add REST API** for external pipeline management (`/api/v1/leads`, `/api/v1/deals`, `/api/v1/interactions`)
- [x] **Add webhook outbound** — notify external systems on deal state changes
- [x] **Add plugin system** — allow custom enrichment sources, classifiers, and responders via Go plugins or WASM
- [ ] **Add multi-tenant support** — isolate data and config per organization

### 10.2 TormentNexus-as-a-Service
- [ ] **Package the sales engine as a reusable service** for other B2B products
- [ ] **Add SaaS billing** with per-seat and per-outreach pricing tiers
- [ ] **Add onboarding wizard** — guide new users through target ICP definition and channel configuration
- [ ] **Add template marketplace** — community-contributed outreach templates and strategies

## Phase 11 — The Elite Enterprise Sales Agent Blueprint

To transform the system into a highly persuasive, context-aware revenue engine, we will implement the following elite sales strategies:

### 11.1 Core Psychological & Strategic Sales Layer
- [ ] **Challenger Sale Framework Integration:** Deliver "asymmetric insight" to prospects.
- [ ] **MEDDPICC Framework Tracking:** Build internal state machine tracking for Metrics, Economic Buyer, Decision Criteria/Process, Paper Process, Identified Pain, and Champion/Competitors.
- [ ] **SPIN Selling Discovery:** Dynamically balance conversational turns across Situation, Problem, Implication, and Need-payoff questions.

### 11.2 Advanced Multi-Agent Technical Architecture
- [ ] **Multi-Agent Orchestration Pattern:** Split the monolith LLM context into specialized nodes.
- [ ] **Sales Strategist Agent:** Evaluates user input to update the MEDDPICC scorecard and determine next micro-goals.
- [ ] **Context & Knowledge Agent (GraphRAG):** Implement a Knowledge Graph combining product documentation, case studies, and industry data.
- [ ] **Execution/Guardrail Agent:** Transform strategic goals into polished, compliant, and authoritative peer-to-peer responses.

### 11.3 Persuasion & Behavioral Tuning
- [ ] **Eliminate AI Tropes:** Remove generic assistant language in favor of an Enterprise Account Executive persona.
- [ ] **Asymmetrical Information Leverage:** Inject data points the prospect lacks into the prompt context.
- [ ] **Cost of Inaction (COI) Focus:** Shift framing from feature-benefits to loss aversion.
- [ ] **Tactical Empathy:** Implement mirroring and labeling techniques in response generation.
