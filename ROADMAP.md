# Roadmap: AI Hypervisor (Nexus & HyperCode)

_Last updated: 2026-05-17, version 1.0.0-alpha.62_

## Status legend

- **Stable** — Production-intended, tested, maintained
- **Beta** — Usable, still evolving
- **Experimental** — Active R&D, not dependable
- **Vision** — Directional only

## Completed (v1.0.0-alpha.62)

### 1. Nexus Kernel: Active Memory Substrate (STABLE)
- **Biological Tiered Memory**: L1 (Scratchpad), L2 (Vault), L3 (Archive) implementation in Go.
- **Heat Mechanics**: Implemented utility-based scoring and temporal decay for all memories.
- **Semantic Search**: Integrated `sqlite-vec` for hyper-fast, local-first context matching.

### 2. The Immune System: Autonomous Healing (BETA)
- **Multi-turn Healer Loop**: Go-native `diagnose -> fix -> verify -> retry` pipeline.
- **Native Verification**: Integrated `CodeExecutor` to run `tsc`, `vitest`, and `go test` autonomously.
- **Intelligence Harvesting**: All healing attempts are now persisted to the L2 Vault for fleet-wide learning.

### 3. Deep Orchestration & Budget Routing (STABLE)
- **PairOrchestrator**: Hardened state machine for `Planner -> Implementer -> Tester -> Critic` turn cycles.
- **Consensus Engine**: Weighted multi-model voting with persistent disagreement logging.
- **Quota Manager**: Real-time token/credit tracking for budget-aware waterfall routing.

### 4. Infrastructure & Parity (STABLE)
- **Go-Native MCP Sync**: Authority over Claude, Cursor, and VS Code configurations moved to Go Kernel.
- **Progressive Skill Disclosure**: BM25/Cosine ranking engine for JIT skill loading.
- **Stream Stabilization**: Exponential backoff and history replay for tRPC subscriptions.

## Active Sprint: Phase 5 - Native Integration

### A. Universal Protocol & Attachment (BETA)
- [x] Implement `hypercode://` protocol handlers for deep-linking.
- [ ] Browser Extension: Implement `hypercode-attach` to link web-based AI chats directly to the local Nexus Kernel.
- [ ] Implement Global Command Hub (Cmd+K) for system-wide HyperCode access.

### B. UI Consolidation (EXPERIMENTAL)
- [ ] Port Maestro logic to Wails/Tauri native runtime (replacing Electron).
- [ ] Implement Intelligence Heatmap visualization for the L2 Vault.
- [ ] Audit 86 dashboard pages for mobile-responsive styling.

### C. Agent Connectivity (VISION)
- [ ] Implement A2A Mesh for cross-host agent collaboration.
- [ ] Decentralized / P2P Memory Swarm for fleet-wide intelligence sync without centralized servers.
- [ ] Context Compression: Native TOON format implementation for context-saving snapshots.

---
*Outstanding! Magnificent! Insanely Great! The collective grows.*
