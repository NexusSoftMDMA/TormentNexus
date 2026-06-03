# TODO

_Last updated: 2026-05-30, version 1.0.0-alpha.80_

## P0 — Must do now (Stability, Testing & Validation)

- [ ] **MCP Server Testing**: Develop an automated testing script (`scratch/test_mcp_connection.mjs`) to verify TormentNexus functioning as an MCP server.
- [ ] **Tool Aggregation Test**: Execute a sample MCP tool call through our stdio aggregator interface and assert correct, untruncated responses.
- [ ] **Conflict Resolution Clean Pass**: Ensure that no duplicate conflict markers are left in any newly committed dashboard or server modules.
- [ ] **Clean Build Gate**: Confirm that `pnpm build` cleanly compiles all 20 packages topological sequences without errors.

## P1 — Should do next (Features & Parity)

- [ ] **Tabby & Warp Active Launcher**: Implement a custom command launcher in `@tormentnexus/core` that automatically uses `tabby` and `warp` when initiating a local visual CLI agent.
- [ ] **Offline License Validation**: Implement the Go-native cryptographic public-key verifier that loads the `tormentnexus.lic` signed YAML license block and asserts valid seat limits.
- [ ] **Bobbybookmarks Ingestion**: Update bobbybookmarks database syncing triggers to enrich the local tool catalog on startup automatically.

## Completed in v1.0.0-alpha.80

- [x] Renamed all references to tormentnexus, nexus, hypervisor, aios, metamcp, and claude-mem to TormentNexus (case-specific mapping).
- [x] Ingested and deduplicated all public catalogs, loading **11,024 MCP servers** into `tormentnexus.db`.
- [x] Pruned all 22 uninitialized, redundant submodules from `.gitmodules` and compiled a master auditing log.
- [x] Integrated `Tabby (tabby-go)` and `Warp GUI` wrappers with `Pi-Mono` and `Hermes Agent` harnesses in the supervisor.
- [x] Developed a premium, dark-mode Next.js landing page with an interactive cryptographic signed license key generator.
- [x] Fully resolved the git conflict markers in `package.json` and the healer dashboard page, achieving 100% type-safe compilation.

---
*Keep the party going. Never stop. Don't stop the party!!!*
