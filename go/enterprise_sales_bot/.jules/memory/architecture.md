# TormentNexus: Architectural & Engineering Governance (v0.9.0)

## 1. System Vision & Architecture
TormentNexus is designed as a high-performance **Modular Monolith** that orchestrates the entire B2B sales lifecycle autonomously. It utilizes a state-driven pipeline to move leads from initial discovery to contract closure with minimal human intervention.

*   **State Machine:** Centralized in PostgreSQL, tracking transitions: `Discovered → Researched → Outreach_Sent → Engaged → Negotiating → Closed_Won/Lost`.
*   **Concurrency Model:** Uses Go’s goroutines for background workers (Scrapers, Enrichers, Researchers, CRM Workers) coordinated via polling intervals and state-based triggers.

## 2. Core Modules & Engineering Patterns
*   **Adapter Pattern (Integrations):**
    *   **CRM:** Unified interface supporting Salesforce and HubSpot.
    *   **LLM (Hermes Agent):** Agnostic gateway supporting NVIDIA, OpenRouter, and local providers.
*   **Strategy Pattern (Outreach):** Multi-channel support (Email, LinkedIn, GitHub) with priority-based routing.
*   **Autonomous Development (AutoDev):** A self-correcting loop where the agent reads `TODO.md`, generates Go code, runs a verification suite (build/test), and uses a `git reset --hard` rollback mechanism on failure. It also incorporates PR feedback for iterative refinement.

## 3. Production Hardening & Security
*   **Security Policies:** Strict `gosec` enforcement. HTTP servers use `ReadHeaderTimeout` to prevent Slowloris attacks. Cookies are hardened with `Secure`, `HttpOnly`, and `SameSiteStrictMode`.
*   **Compliance:** GDPR-ready with native soft-delete support (`deleted_at`) and dedicated data export/deletion endpoints.
*   **Resilience:** Global rate limiting (5 req/s), circuit breakers for external APIs, and top-level panic recovery in `main.go`.
*   **Observability:** Structured JSON logging via `slog`, Prometheus metrics for business KPIs (leads, deals, won/lost), and `pprof` for performance profiling.

## 4. Engineering Decisions & Constraints
*   **Language:** Standardized on **Go 1.25.0**.
*   **Database:** PostgreSQL with `golang-migrate` for versioned schema evolution.
*   **CI/CD:** Enforced testing of all internal packages. Integration tests require an ephemeral DB; otherwise, they are skipped to ensure local environment stability.
*   **Deployment:** Supports Docker and Docker Compose (staging/production), with sensitive configurations managed via AES-GCM encrypted secrets at rest.

---

I have verified the v0.9.0 hardening changes (GDPR soft-deletes, encryption utility, metrics, and security hardening) by running the build and test suites. I am now proceeding to finalize the commit and move to the upstream merge.