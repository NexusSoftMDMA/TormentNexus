# TODO

_Last updated: 2026-06-17, version 1.0.0-alpha.132_

## P0 — Must do now (Stability, Testing & Validation)
- [x] **Track A: MCP Discovery**: Execute discovery script to rank top 500 MCP servers and seed state DB. (14,250 rows in assimilation_state.db)
- [x] **Track B: Skill Registry**: Verify 3-tier loading with comprehensive unit tests. (Completed alpha.128)
- [x] **Track B: Bulk Skill Assimilation**: Assimilated 3,229 unique skills from 7 harness ecosystems. (Completed alpha.128)
- [x] **Track D: Prompt Migration**: Migrate hardcoded prompts to SQLite. (Completed alpha.127)
- [x] **Branch Merge**: Intelligently merged `jules/baseline-128-hardened` into `main`, fast-forwarded `assimilation-pipeline` and `assimilation-final`. (Completed alpha.132)
- [x] **README Rewrite**: Comprehensive 657-line README with full architecture, capabilities, and roadmap. (Completed alpha.132)
- [x] **Data Integrity**: 14,250 total / 10,796 done / 10 pending / 9 processing (swarm actively finishing). (alpha.134)
- [x] **Swarm Output**: Swarm running persistently with 7-model pool. Generated 34 new Go tool stubs. (alpha.134)
- [x] **Go Build Verification**: Root build passes clean (4,042 tool files). (alpha.134)

## P1 — Should do next (Integrations)
- [x] **Harness Integration**: Integrate Tabby, Warp, Hyper, Hyperharness, Hermes Agent, and Pi-Mono. (Verified alpha.127)
- [x] **A2A Skill Registry**: Map assimilated skills into FreeLLM A2A registry. (Completed alpha.128)
- [x] **Skill HTTP API**: Wire skill store into Go sidecar HTTP endpoints. (Completed alpha.130)
- [x] **Browser Automation MCP**: Finalize tests and add optional args. (Completed alpha.129)
- [ ] **ChunkHound / Probe Integration**: Implement remaining assimilated MCP search tools as native handlers.
- [ ] **Bobbybookmarks Sync**: Configure automatic sync call triggers for catalog scraping. (Blocked by DNS failure — use Smithery.ai or Glama.ai)
- [ ] **New Native Tools**: Implement `browser-use` and `browsermcp` specialized logic if needed (currently aliased to playwright).
- [x] **Session Import**: Format resolved — wraps JSONL in ExportPackage format (228 sessions detected). Orchestrator POST endpoint missing for actual restoration.
- [ ] **Git LFS**: Consider tracking large `.db` files (provider_metrics.db 145MB, tormentnexus.db 34MB) with Git LFS to avoid repo bloat.
- [x] **.out Cleanup**: `swarm_*.out` and `*.pid` added to `.gitignore`. (alpha.133)

## P2 — Enterprise Readiness & Security
- [x] **License Validation**: Implement Ed25519 license token validation in Go sidecar. (Verified alpha.127)
- [ ] **Compliance Boundary**: Separate SSO/RBAC/Audit logic into enterprise wrapper.
- [x] **Enterprise Security**: SSO/RBAC middleware and JSONL auditing added from jules merge. (alpha.132)
- [x] **Autonomous CI/CD**: `deployment_manager`, `health_monitor`, `repository_healer` added from jules merge. (alpha.132)

## P3 — Future Enhancements
- [ ] **Skill Evolution**: With ~3,000+ skills loaded, implement win-rate tracking, auto-retirement of low-performing skills, and `/evolve` command.
- [ ] **Catalog DB Sync**: Index new skills into `catalog.db` for unified search.
- [ ] **Submodule Removal**: Systematic removal of redundant submodules after native reimplementation.
- [ ] **P2P Memory**: Implement gossip protocol for decentralized context sharing.
- [ ] **L3 Cold Archive**: Implement long-term compressed memory tier for infinite context.
- [ ] **Fleet-Wide Intelligence**: Cross-machine memory sharing via encrypted mesh.
- [ ] **Wails Native Runtime**: Replace Electron with Go-native desktop shell.
- [ ] **Deep Link Protocol**: Expand `tormentnexus://` protocol for browser-to-kernel attachment.

---
*Keep the party going. Never stop. Don't stop the party!!!*
