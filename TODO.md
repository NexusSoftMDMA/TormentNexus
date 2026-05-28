# TODO

_Last updated: 2026-05-25, version 1.0.0-alpha.63_

## P0 — Must do now (Stability & Truth)

- [x] **Protocol Scaffolding**: Implement the basic `hypercode://` handler in the Go kernel to support session attachment.
- [x] **Dashboard Truth Pass**: Verify that the "Immune System" status card in the dashboard shows real-time data from the Go `HealerService`.
- [x] **L2 Vault Visualization**: Wire the `vaultRecords` query to the Next.js frontend to show persistent heal history.
- [x] **Mobile Style Audit**: Fix overlapping elements in the Sidebar and KernelStats on small viewports.

## P1 — Should do next (Features & Parity)

- [x] **Browser Extension Attach**: Implement the DOM injection to add a "Hypercode Kernel" button to Claude.ai and ChatGPT.
- [x] **Wails Migration**: Scaffold the `apps/native-ui` directory using Wails for the Go-native dashboard.
- [x] **A2A Mesh Protocol**: Implement the discovery layer for agents running on different local network hosts.
- [x] **TOON Format**: Implement the native Go encoder/decoder for the TOON (Thread-Oriented Object Notation) context format.

## P2 — Helpful but not urgent

- [x] **Intelligence Heatmap**: Create a 3D visualization of the L2 Vault using the embedding vectors.
- [x] **Skill Marketplace**: Implement the REST API for downloading community-contributed skills.
- [x] **Decentralized Memory**: Scoping Phase for P2P memory sync using gossip protocols.

## Completed (v1.0.0-alpha.63)
- [x] Bridged TypeScript `healerRouter` to Go-native `HealerService` and `VectorStore`.
- [x] Implemented total count metrics for L2 Vault records in Go.
- [x] Standardized field mapping (snake_case -> PascalCase) for native record UI compatibility.

## Completed (v1.0.0-alpha.62)
- [x] Autonomous Healer Loop (diagnose -> fix -> verify -> retry) in Go.
- [x] L2 Vault persistence for Healer events.
- [x] Integrated CodeExecutor with Healer for native verification.
- [x] Consensus Engine with L2 Vault logging.
- [x] Quota Manager for budget-aware model switching.
- [x] Biological tiered memory (L1/L2/L3) logic and decay.
- [x] Go-native MCP Sync for Claude/Cursor/VSCode.
- [x] Standardized ports: Go (4300), Bridge (4100), Dashboard (3000), Socket.io (3001).

---
*Keep the party going. Never stop. Don't stop the party!!!*
