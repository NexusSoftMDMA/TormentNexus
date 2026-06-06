# ROADMAP: TormentNexus Kernel & TormentNexus Dashboard

_Last updated: 2026-06-07, version 1.0.0-alpha.126_

## Status Legend
- **Stable** — Production-intended, tested, maintained
- **Beta** — Usable, still evolving
- **Experimental** — Active R&D, not dependable
- **Vision** — Directional only

## Completed (v1.0.0-alpha.126)
### 1. Rebranding & Database Conversion (STABLE)
- **TormentNexus Universal Rebrand**: Complete case-insensitive, case-specific refactoring across source modules, config files, package dependencies, and directories.
- **Unified Catalog SQLite Storage**: Ingested and deduplicated standard technical assets, creating a robust local dataset of **11,024 populated MCP servers** stored directly in `tormentnexus.db`.

### 2. Track B2: Skill Registry Progressive & Relational Linkage (STABLE)
- **Jaccard Duplication Rules (90% Threshold)**: Near-duplicate skills linked to canonical entry via `canonical_id`.
- **Progressive Loading**: Implemented `skill_list`, `skill_get`, and `skill_search` for efficient context hygiene.

### 3. Track A: MCP Assimilation (BETA)
- **Native Go Reimplementation**: Replaced dozens of SSE/Stdio MCP servers with native Go modules (Arxiv, Exa, Semantic Scholar, etc.).

## Active Sprint: Phase 6 - Comprehensive Assimilation & Enterprise Readiness

### A. Track A: Full MCP Assimilation (BETA)
- [ ] Assimilate top 500 MCP servers as native Go modules.
- [ ] Eliminate all external MCP server dependencies and submodules.

### B. Track C & D: Hermes Addons & Prompt Library (EXPERIMENTAL)
- [ ] Research and rank top 500 Hermes-agent addons.
- [ ] Migrate all hardcoded prompts to `data/prompt_library.db` with Go-native retrieval tools.

### C. Enterprise Licensing & Compliance (EXPERIMENTAL)
- [ ] Implement Ed25519-signed license token validation in the Go sidecar.
- [ ] Develop Enterprise landing page with SSO/RBAC configuration stubs.

### D. Default Agent Harness Integration (BETA)
- [ ] Integrate Tabby, Warp, Hyper, Hyperharness, Hermes Agent, and Pi-Mono as default harnesses.
- [ ] Automate Bobbybookmarks ingestion for continuous tool catalog updates.

---
*Outstanding! Magnificent! Insanely Great! The collective grows.*
