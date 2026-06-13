# TODO

_Last updated: 2026-06-13, version 1.0.0-alpha.128_

## P0 — Must do now (Stability, Testing & Validation)
- [x] **Track A: MCP Discovery**: Execute discovery script to rank top 500 MCP servers and seed state DB.
- [ ] **Track B: Skill Registry**: Verify 3-tier loading with comprehensive unit tests.
- [x] **Track D: Prompt Migration**: Migrate hardcoded prompts to SQLite. (Completed in alpha.127)
- [ ] **Data Integrity**: Clean up `assimilation_state.db` statuses for already assimilated tools.

## P1 — Should do next (Integrations)
- [x] **Harness Integration**: Integrate Tabby, Warp, Hyper, Hyperharness, Hermes Agent, and Pi-Mono. (Verified in alpha.127)
- [ ] **Bobbybookmarks Sync**: Configure automatic sync call triggers for catalog scraping.
- [ ] **New Native Tools**: Implement `browser-use` and `browsermcp` specialized logic if needed (currently aliased to playwright).

## P2 — Enterprise Readiness & Security
- [x] **License Validation**: Implement Ed25519 license token validation in Go sidecar. (Verified in alpha.127)
- [ ] **Compliance Boundary**: Separate SSO/RBAC/Audit logic into enterprise wrapper.

## P3 — Future Enhancements
- [ ] **Submodule Removal**: Systematic removal of redundant submodules after native reimplementation.
- [ ] **P2P Memory**: Implement gossip protocol for decentralized context sharing.

---
*Keep the party going. Never stop. Don't stop the party!!!*
