# AI Hypercode (Hypercode) - Comprehensive Architectural Memory

This document serves as the authoritative summary of the project's architecture, established patterns, and strategic decisions as of version **1.0.0-alpha.56**.

## 1. Core Brand Strategy: Hypercode & TormentNexus
The project has transitioned from "Hypercode" to a sophisticated dual-brand strategy:
*   **Hypercode (The Kernel/Substrate):** The underlying "AI Hypercode" and coordination runtime. It treats Large Language Models as "guest operating systems" and manages the low-level memory, routing, and execution buses.
*   **TormentNexus (The Flagship Product):** The user-facing, local-first autonomous coding environment powered by the Hypercode kernel.
*   **Rationale:** This separation prevents "identity collapse" by allowing the infrastructure (Hypercode) to remain a general-purpose agent coordination layer while the product (TormentNexus) focuses on delivering a world-class developer experience.

## 2. Active Tiered Memory Substrate (Implemented Phase 1)
Memory has been transitioned from a passive storage model into an active, biological-inspired system:
*   **Heat Scoring (0-100):** Every memory entry tracks real-world utility. Heat increases on access and decays over time (24-hour half-life).
*   **Tiered Hierarchy:**
    *   **L1 (Working Memory):** High-heat entries (>80) are automatically promoted into the model's immediate context.
    *   **L2 (Vault):** Long-term storage (LanceDB/SQLite) providing semantic recall and historical heuristics.
*   **Tool-Outcome Feedback:** A critical closed-loop system where `MemoryManager.recordToolOutcome()` boosts the heat of successful execution patterns and demotes failures, allowing the system to "learn" which tool sequences are reliable.

## 3. "Kernel / Control Plane" Architecture
The repository is organized into two primary functional layers:
*   **/kernel**: The deterministic "brain" of the system.
    *   `runtime`: Schedulers and execution lifecycles.
    *   `memory`: Active L1/L2 substrate.
    *   `router`: Semantic tool selection and waterfall provider fallback.
*   **/control-plane**: The "observer" layer.
    *   `UI`: Dashboards and TUI/Web interfaces.
    *   `Telemetry`: Real-time traffic inspection and execution graphs.
*   **Rationale:** Ensures the system can run headless (as `TormentNexusd`) while multiple clients (CLI, Web, Browser Extension) interact with it via standardized tRPC/Go APIs.

## 4. State Authority & The Go Sidecar
*   **Go Sidecar (Port 4300):** Acts as the high-performance state authority and BM25 ranking engine.
*   **TS Bridge (Port 4100):** Serves as the primary control-plane bridge, managing tRPC routers and web-compatible integrations.
*   **Assimilation Pattern:** Logic is systematically migrated from reference submodules into native Go/TypeScript implementations to ensure 100% wire-protocol parity and eliminate dependency on fragile third-party wrappers.

## 5. Intelligence Management: Progressive Disclosure
*   **The Problem:** Exposing hundreds of MCP tools or skills to an LLM degrades reasoning ("context poisoning") and inflates token costs.
*   **The Solution:** The **Decision System**. It uses ranked discovery and LRU (Least Recently Used) eviction to ensure only the 3-5 most relevant tools or skills are present in the model's active working set at any given moment.

## 6. Enterprise Blueprint Alignment
The system is being architected to support "AI Hypercode Enterprise" features:
*   **Fleet Management:** Private vLLM pod orchestration for GPU clusters using the `pi-pods` model.
*   **Governance:** SSO/SAML integration and RBAC.
*   **Security:** Immutable audit ledgers for every tool call and PII scrubbing (DLP) for outgoing prompt streams.

## 7. Hardened Execution Standards
*   **Security Pattern:** All command executions must use **tokenized argument arrays** with `shell: false`. Blind `child_process.exec` and `sed` scripts are strictly prohibited to prevent shell injection and codebase corruption.
*   **Parity Principle:** Hypercode must provide 1:1 behavioral and schema parity for tools expected by proprietary models (e.g., Claude Code's `bash`, `read`, `write`).

## 8. Meta-Protocol for Future Development
*   **Truth over Fiction:** UI dashboards must always reflect actual database rows and runtime state.
*   **Autonomous Momentum:** The agent is authorized to proceed through the established `ROADMAP.md` (currently entering **Phase 2: Autonomy/Self-Healing**) without constant approval.
*   **Version Synchronization:** Every major functional increment must sync `VERSION`, `CHANGELOG.md`, `VISION.md`, and all monorepo `package.json` manifests.

---
*This memory is the definitive record of the Hypercode/TormentNexus architecture and directs all subsequent implementation turns.*