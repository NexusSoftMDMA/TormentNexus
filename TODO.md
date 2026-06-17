# TODO

_Last updated: 2026-06-14, version 1.0.0-alpha.129_

## P0 — Must do now (Stability, Testing & Validation)
- [x] **Track A: MCP Discovery**: Execute discovery script to rank top 500 MCP servers and seed state DB.
- [ ] **Track B: Skill Registry**: Verify 3-tier loading with comprehensive unit tests.
- [x] **Track D: Prompt Migration**: Migrate hardcoded prompts to SQLite. (Completed in alpha.127)
- [ ] **Data Integrity**: Clean up `assimilation_state.db` statuses for already assimilated tools.

## P1 — Should do next (Integrations)
- [x] **Harness Integration**: Integrate Tabby, Warp, Hyper, Hyperharness, Hermes Agent, and Pi-Mono. (Verified in alpha.127)
_Last updated: 2026-06-14, version 1.0.0-alpha.130_

## P0 — Must do now (Stability, Testing & Validation)
- [x] **Track A: MCP Discovery**: Execute discovery script to rank top 500 MCP servers and seed state DB.
- [x] **Track B: Skill Registry**: Verify 3-tier loading with comprehensive unit tests.
- [x] **Track B: Bulk Skill Assimilation**: Assimilated 3,229 unique skills from `~/.a5c`, `~/.agent/skills`, `~/.ccs`, `~/.hermes/skills`, `~/.pi`, `~/.agents/skills` into `~/.tormentnexus/skills/`.
- [x] **Track C: Hermes Research**: Research and rank top 500 Hermes-agent addons.
- [x] **Track D: Prompt Migration**: Migrate hardcoded prompts to SQLite.

## P1 — Should do next (Integrations)
- [x] **Harness Integration**: Integrate Tabby, Warp, Hyper, Hyperharness, Hermes Agent, and Pi-Mono.
- [x] **A2A Skill Registry**: Map assimilated skills into the FreeLLM A2A registry as `AgentSkill` structs so swarm agents can discover and use them via `findAgentForSkill(skillID)`.
- [ ] **Skill HTTP API**: Wire the skill store into the Go sidecar's HTTP API endpoints (`/api/skills/list`, `/api/skills/get`, `/api/skills/search`). ✅ Implemented (v1.0.0-alpha.130) using `orchestration.GlobalSkillRegistry`.
- [x] **Browser Automation MCP**: Finalize tests and add optional args (`fullPage`, `timeout`) for browser handlers.
- [ ] **ChunkHound / Probe Integration**: Implement remaining assimilated MCP search tools as native handlers.
- [ ] **Bobbybookmarks Sync**: Configure automatic sync call triggers for catalog scraping.
- [ ] **New Native Tools**: Implement `browser-use` and `browsermcp` specialized logic if needed (currently aliased to playwright).

## P2 — Enterprise Readiness & Security
- [x] **License Validation**: Implement Ed25519 license token validation in Go sidecar. (Verified in alpha.127)
- [ ] **Compliance Boundary**: Separate SSO/RBAC/Audit logic into enterprise wrapper.

## P3 — Future Enhancements
- [ ] **Skill Evolution**: With ~3,000+ skills loaded, implement win-rate tracking, auto-retirement of low-performing skills, and `/evolve` command.
- [ ] **Catalog DB Sync**: Index new skills into `catalog.db` (`published_mcp_servers`, `published_mcp_config_recipes`) for unified search.
- [ ] **Submodule Removal**: Systematic removal of redundant submodules after native reimplementation.
- [ ] **P2P Memory**: Implement gossip protocol for decentralized context sharing.

---
*Keep the party going. Never stop. Don't stop the party!!!*
