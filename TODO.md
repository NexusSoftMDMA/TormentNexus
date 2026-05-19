# TODO

_Last updated: 2026-05-17, version 1.0.0-alpha.62_

## P0 — Must do now (Stability & Truth)

- [x] **Protocol Scaffolding**: Implement the basic `hypercode://` handler in the Go kernel to support session attachment.
- [ ] **Dashboard Truth Pass**: Verify that the "Immune System" status card in the dashboard shows real-time data from the Go `HealerService`.
- [ ] **L2 Vault Visualization**: Wire the `vaultRecords` query to the Next.js frontend to show persistent heal history.
- [ ] **Mobile Style Audit**: Fix overlapping elements in the Sidebar and KernelStats on small viewports.

## P1 — Should do next (Features & Parity)

- [ ] **Browser Extension Attach**: Implement the DOM injection to add a "Nexus Kernel" button to Claude.ai and ChatGPT.
- [ ] **Wails Migration**: Scaffold the `apps/native-ui` directory using Wails for the Go-native dashboard.
- [ ] **A2A Mesh Protocol**: Implement the discovery layer for agents running on different local network hosts.
- [ ] **TOON Format**: Implement the native Go encoder/decoder for the TOON (Thread-Oriented Object Notation) context format.

## P2 — Helpful but not urgent

- [ ] **Intelligence Heatmap**: Create a 3D visualization of the L2 Vault using the embedding vectors.
- [ ] **Skill Marketplace**: Implement the REST API for downloading community-contributed skills.
- [ ] **Decentralized Memory**: Scoping Phase for P2P memory sync using gossip protocols.

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
