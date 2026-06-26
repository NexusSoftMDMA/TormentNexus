# ROADMAP: TormentNexus Kernel & TormentNexus Dashboard

_Last updated: 2026-06-17, version 1.0.0-alpha.132_

## Status Legend
- **Stable** — Production-intended, tested, maintained
- **Beta** — Usable, still evolving
- **Experimental** — Active R&D, not dependable
- **Vision** — Directional only

## Completed (v1.0.0-alpha.132)
### 1. Comprehensive Documentation & Merge (STABLE)
- **README.md Rewrite**: Expanded from 82 lines to 657 lines covering full architecture, capabilities, monorepo structure, Go sidecar, dashboard, MCP ecosystem, memory model, swarm, and API surface.
- **Branch Reconciliation**: Intelligently merged `jules/baseline-128-hardened` into `main`, fast-forwarded `assimilation-pipeline` and `assimilation-final` to merged tip.
- **All Branches Synchronized**: `main`, `jules`, `feat/assimilation-pipeline`, `feature/assimilation-final` all point to `988ec114a`.

### 2. Autonomous Engineering & Orchestration (STABLE)
- **CI/CD Pipeline**: Integrated multi-stage `deployment_manager` (lint, build, test, containerize) via `.github/workflows/autonomous-deploy.yml`.
- **Repository Sync**: Automated dependency management and version alignment via `go/cmd/repo_sync`.
- **Self-Healing**: Native Go `health_monitor` and `repository_healer` for autonomous kernel maintenance.
- **Enterprise Security**: SSO/RBAC middleware and structured JSONL auditing in `go/internal/enterprise/`.

### 3. Dashboard Widgets (BETA)
- **BrowserToolWidget**: Real-time browser automation control panel.
- **VibeCheckWidget**: Code quality and pattern analysis widget.

### 4. Assimilation Scale (STABLE)
- **14,250+ MCP Servers Tracked**: In `assimilation_state.db` with 3,270 pending, 10,796 implemented.
- **3,900+ Native Go Tools**: Replacing external MCP server dependencies.
- **11,024+ Populated Catalog**: In `catalog.db` with verified metadata.

## Completed (v1.0.0-alpha.131)
- **Swarm v7 Recovery**: Generated ~130 new MCP server Go tool wrappers, removed 2,268 lines of obsolete files.
- **Session Import Pipeline**: Validated 49 candidates from `~/.claude` and `~/.aider` artifacts, 586 imported sessions tracked.
- **Version Sync**: All 35 workspace packages synchronized to `1.0.0-alpha.131`.

## Completed (v1.0.0-alpha.130)
- **Skill HTTP API**: Implemented `/api/skills/list`, `/api/skills/get`, `/api/skills/search` with 10 passing unit tests.
- **API Documentation**: Updated `docs/API_ENDPOINTS.md` with skill endpoints.

## Completed (v1.0.0-alpha.129)
- **Browser Automation**: Native `chromedp` handlers (navigate, screenshot, evaluate, click, fill) replacing 5+ separate MCP entries.
- **A2A Skill Registry**: Global singleton with `FindAgentForSkill` helper.

## Completed (v1.0.0-alpha.128)
- **Bulk Skill Assimilation**: 3,229 unique skills from 7 harness ecosystems with Jaccard deduplication.
- **Hardened Kernel**: Restored ~60 swarm tool registrations and verified compilation.

## Completed (v1.0.0-alpha.127)
- **Native Go Tools**: High-performance handlers for `ripgrep`, `anyquery`, `codemod`.
- **E2E Integration Testing**: Formal test suite in `go/internal/tools/e2e_test.go`.
- **API Documentation**: 600+ endpoint reference in `docs/API_ENDPOINTS.md`.

## Completed (v1.0.0-alpha.126)
- **Universal Rebrand**: Case-insensitive refactoring across all source modules.
- **Catalog SQLite Storage**: 11,024 populated MCP servers in `tormentnexus.db`.

## Active Sprint: Phase 8 - Predictive Intelligence & Enterprise Readiness

### A. Track A: Full MCP Assimilation (BETA)
- [ ] Assimilate top 500 MCP servers as native Go modules. (3,900+ done, 3,270 pending)
- [ ] Eliminate all external MCP server dependencies and submodules.

### B. Track B: Skill Registry Progressive & Relational Linkage (STABLE)
- [x] Jaccard Duplication Rules (90% Threshold): Near-duplicate skills linked to canonical entry.
- [x] Progressive Loading: Implemented `skill_list`, `skill_get`, `skill_search`.
- [ ] Win-rate tracking and auto-retirement of low-performing skills.

### C. Track C: Enterprise Licensing & Compliance (EXPERIMENTAL)
- [x] Ed25519-signed license token validation in Go sidecar.
- [ ] Enterprise landing page with SSO/RBAC configuration.
- [ ] Structured audit logs for native tool execution.
- [ ] RBAC permission schema for multi-user environments.

### D. Track D: Default Agent Harness Integration (BETA)
- [x] Integrate Tabby, Warp, Hyper, Hyperharness, Hermes Agent, and Pi-Mono as default harnesses.
- [ ] Automate Bobbybookmarks ingestion (use Smithery.ai or Glama.ai as alternative).

### E. Phase 8: Predictive Intelligence (VISION)
- [ ] Predictive Conversational Tool Injection: Local model-based prediction of relevant tools.
- [ ] L3 Cold Archive: Long-term compressed memory tier for infinite context.
- [ ] Fleet-Wide Intelligence: Cross-machine memory sharing via encrypted mesh.

### F. Phase 9: Native Runtime (VISION)
- [ ] Wails Native Runtime: Replace Electron with Go-native desktop shell.
- [ ] Deep Link Protocol: Expand `tormentnexus://` for browser-to-kernel attachment.

---
*Outstanding! Magnificent! Insanely Great! The collective grows.*
