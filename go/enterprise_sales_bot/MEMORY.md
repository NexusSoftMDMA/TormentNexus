# Memory: Architectural Observations & Design Preferences

## Current State

- The project is at v0.5.1 with Hermes Agent LLM integration, real-time quote generation, and stabilized CI pipelines.
- Core modules: scraper, enricher, researcher, communication, CRM, billing, deploy, autodev.
- **LLM provider is now REAL** via Hermes Agent gateway (WSL) — replaces MockLLMProvider.
- **Intent classifier is now REAL** via LLMIntentClassifier when Hermes is available.
- Remaining mocks: enrichment (Apollo), job board scraper, email send/receive, billing (Stripe).

## Architectural Traits

- **Event-Driven:** Designed to be asynchronous and event-driven via background worker goroutines.
- **Interface-Based:** External integrations (scrapers, CRM, billing, LLM, email) are abstracted behind interfaces for easier mocking and rotation.
- **Rigid State Management:** Lead transitions are handled via an atomic state machine in PostgreSQL with a 7-state enum.
- **Automation First:** Every feature is built with the intent of being fully autonomous.
- **Self-Development Loop:** The system includes an `autodev` module that autonomously selects tasks from `TODO.md`, proposes changes, and verifies them via a branch-push-PR-merge lifecycle.
- **Autonomous Continuous Delivery:** Codebase updates initiated by the bot trigger automated GitHub Action workflows for testing and deployment to ensure system stability.
- **Self-Learning Sales Engine:** The `communication` package features a `LearningSalesEngine` that analyzes interaction history and lead context to decide on autonomous responses, state transitions, or human escalation.
- **Prompt Optimization Feedback Loop:** The `RAGResponseGenerator` implements a feedback loop by injecting successful past interactions (flagged upon `StateClosedWon`) into the prompt context.
- **Dual-Direction Merge Engine:** Reconciles autonomous feature branches by forward-merging into main and reverse-merging main back into features to prevent drift.

## Known Technical Debt

- **CRLF Test Failure:** `internal/gitres/resolve_test.go::TestResolveConflictTheirs` fails on Windows due to `\r\n` vs `\n` mismatch.
- **Unstructured Logging:** All modules use `log.Printf` — no structured JSON logging, no log levels, no correlation IDs.
- **No DB Migration Runner:** Migrations must be applied manually; they are not auto-applied on startup.
- **No Rate Limiting:** HTTP endpoints accept unlimited requests.
- **No Pagination:** Dashboard hardcodes `LIMIT 20` for deals.
- **Missing Indices:** `interactions.success` and `deals.current_state` lack database indices.
- **Hardcoded Worker Intervals:** Background worker intervals are configurable via env vars but not via config file.
- **Hermes Dependency:** LLM calls depend on Hermes gateway being running in WSL. If Hermes is down, bot falls back to mock.

## Design Preferences

- **Go (Golang):** Preferred for the orchestration layer due to its performance and concurrency model.
- **PostgreSQL:** Used for reliable relational data storage and state tracking.
- **Headless Scrapers:** Required for robust data extraction from modern web platforms.
- **Atomic Commits:** Prefer small, descriptive commits that correspond to specific features or fixes.
- **Interface-Driven Design:** All external dependencies should be behind Go interfaces for testability and swappability.
- **CI-Gated Merging:** No code reaches main without passing all tests.

## Integration Status

| Integration | Status | Implementation |
|---|---|---|
| GitHub API (target discovery) | ✅ Real | `pkg/agents/discovery.go` with `go-github` |
| GitHub API (CI tracking) | ✅ Real | `internal/deploy/github_tracker.go` |
| GitHub API (PR management) | ✅ Real | `internal/gitcheck/pr.go` with `go-github` |
| Stripe billing | ✅ Real | `internal/billing/billing.go` with `stripe-go` |
| REST CRM client | ✅ Real | `internal/crm/crm.go` with generic REST |
| **LLM provider** | **✅ Real** | **`internal/llm/hermes.go::HermesLLMProvider` via Hermes Agent gateway** |
| **Intent classifier** | **✅ Real** | **`LLMIntentClassifier` via Hermes (keyword mock fallback)** |
| Enrichment (Apollo) | ❌ Mock | `internal/enrichment/worker.go::MockApolloSource` |
| Job board scraper | ❌ Mock | `internal/scraper/scraper.go::MockJobBoardSource` |
| Email sending | ❌ Not implemented | Outbound is logged but not sent |
| Email receiving | ❌ Not implemented | Inbound is simulated by polling DB |
