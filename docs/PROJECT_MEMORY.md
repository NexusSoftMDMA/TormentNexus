[PROJECT_MEMORY]

# AI Hypervisor (Nexus) - Comprehensive Architectural Memory

This document summarizes the foundational architecture, established patterns, and strategic decisions of the project as of version **1.0.0-alpha.56**.

## 1. Strategic Identity: Nexus & HyperCode
The project has successfully pivoted from "Borg" to a dual-brand infrastructure model to resolve identity collapse:
*   **Nexus (The Kernel/Hypervisor):** The underlying coordination runtime and "AI Hypervisor." It treats LLMs as "guest operating systems" and manages the low-level memory, routing, and execution buses.
*   **HyperCode (The Product):** The user-facing, local-first autonomous coding environment powered by the Nexus kernel.
*   **Rationale:** To provide a general-purpose coordination layer (Nexus) while delivering a focused, world-class developer tool (HyperCode).

## 2. Active Tiered Memory Substrate (Implemented Phase 1)
Memory has been transitioned from passive storage to an active, biological-inspired substrate:
*   **Heat Scoring (0-100):** Every memory entry tracks utility. Heat increases on access and decays exponentially (24-hour half-life).
*   **Tiered Hierarchy:**
    *   **L1 (Working Memory):** High-heat entries (>80) promoted into the immediate context.
    *   **L2 (Vault):** Long-term storage (LanceDB/SQLite) for semantic recall.
*   **Tool-Outcome Feedback:** A closed-loop system where `MemoryManager.recordToolOutcome()` boosts the heat of successful patterns and demotes failures.

## 3. "Kernel / Control Plane" Topology
*   **/kernel**: The deterministic "brain" (runtime, memory, router).
*   **/control-plane**: The "observer" layer (UI, Telemetry).
*   **Rationale:** Ensures the system can run headless while multiple clients observe and direct it via standardized APIs.

## 4. State Authority & The Sidecar Pattern
*   **Go Sidecar (Port 4300):** High-performance state authority and ranking engine.
*   **TS Bridge (Port 4100):** Primary control-plane bridge and tRPC host.
*   **Modular Monolith:** Preference for a unified Go engine over micro-binaries to eliminate IPC lag.

## 5. Intelligence Management: Progressive Disclosure
*   **Decision System:** Uses ranked discovery and LRU eviction to ensure only the 3-5 most relevant tools/skills are present in the model's active working set at any time.

## 6. Hardened Execution & Security
*   **Standard:** Tokenized argument arrays with `shell: false` for all command executions.
*   **Parity Principle:** 1:1 behavioral and schema parity for tools expected by proprietary models (e.g., Claude Code).

---
### Meta-Protocol for Future Sessions
1.  **Truth over Fiction:** UI and CLI must reflect real database rows and runtime state.
2.  **Autonomous Momentum:** Authorization to proceed through ROADMAP.md phases (Phase 2: Autonomy/Self-Healing).
3.  **Documentation Sync:** Every version bump must sync VERSION, CHANGELOG, and manifests.
