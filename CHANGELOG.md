# Changelog

## [1.0.0-alpha.126] - 2026-06-07
### Added
- **Assimilation State Database**: Created `data/assimilation_state.db` to track the status of MCP servers, Hermes addons, and skill ingestion.
- **Project Roadmap & TODO Update**: Re-aligned project goals with the comprehensive assimilation pipeline (Tracks A, B, C, D).
- **Harness Integration Preparation**: Identified external repos (Tabby, Warp, etc.) for integration as default harnesses.

## [1.0.0-alpha.126] - 2026-06-07
### Added
- **Assimilation State Database**: Created `data/assimilation_state.db` to track the status of MCP servers, Hermes addons, and skill ingestion.
- **Harness Integrations**: Integrated Tabby, Warp, Hyper, Hyperharness, Hermes-Agent, and Pi-Mono as submodules and added Go handlers.
- **Bobbybookmarks Integration**: Added native Go handler for `bobbybookmarks_sync`.
- **Enterprise Licensing**: Implemented Ed25519-signed license validation and updated landing page with license generator.

## [1.0.0-alpha.125] - 2026-06-06
### Added
- **Track B2 — SQLite Skill Registry relational duplicate linkage**:
  - Implemented 90% Jaccard word-similarity threshold inside `skill_registry.go` HandleSkillStore.
  - Linked near-duplicate skills (similarity 70-89%) to their canonical entry using `canonical_id`.
- **Fixed test suite issues**:
  - Fixed variable redeclaration error in `cmd/foundation_http_test.go`.
  - Resolved `htormentnelloxus` test snapshot difference.
