# Architectural Assessment: TormentNexus Cognitive Control Plane
Date: 2026-03-20

## 1. Architectural Analysis
TormentNexus is architected as a **Local-First Cognitive Control Plane**. It sits between high-level AI agents and low-level infrastructure (tools, providers, and data).

*   **Modular Kernel Design**: The system is divided into specialized "kernels" located in `packages/core`. This includes the **MCPAggregator** (tool routing), **SessionSupervisor** (process isolation), and the **ProviderTruth** service (quota and auth verification).
*   **Verification Layer (The "Evidence Lock")**: Unlike standard agent frameworks that trust tool definitions blindly, TormentNexus incorporates a formal verification layer. This is represented by `TORMENTNEXUS_MASTER_INDEX.jsonc`, which tracks "Truth" levels (L0-L3) for tool parity across different AI platforms (Copilot, Cursor, Claude Code, etc.).
# Architectural Assessment: tormentnexus Cognitive Control Plane
Date: 2026-03-20

## 1. Architectural Analysis
tormentnexus is architected as a **Local-First Cognitive Control Plane**. It sits between high-level AI agents and low-level infrastructure (tools, providers, and data).

*   **Modular Kernel Design**: The system is divided into specialized "kernels" located in `packages/core`. This includes the **MCPAggregator** (tool routing), **SessionSupervisor** (process isolation), and the **ProviderTruth** service (quota and auth verification).
*   **Verification Layer (The "Evidence Lock")**: Unlike standard agent frameworks that trust tool definitions blindly, tormentnexus incorporates a formal verification layer. This is represented by `TORMENTNEXUS_MASTER_INDEX.jsonc`, which tracks "Truth" levels (L0-L3) for tool parity across different AI platforms (Copilot, Cursor, Claude Code, etc.).
*   **Supervised Execution**: It uses a worktree-based isolation model. Every agent operation is supervised to ensure failures are contained and the "state of the world" remains truthful and observable.
*   **Tiered Memory Architecture**: It employs a three-tier memory strategy:
    *   **L1 (Session)**: Immediate context continuity.
    *   **L2 (Working)**: Active notes and extracted facts (SQLite/FTS5).
    *   **L3 (Long-Term)**: Semantic retrieval across historical data (LanceDB/Vector).

## 2. Frameworks & Foundations
*   **Protocol**: **Model Context Protocol (MCP)** is the foundational "language" of the project.
*   **Runtime**: **Node.js (>= 22.12.0)**. Chosen for maximum compatibility with the MCP ecosystem.
*   **Frontend**: **Next.js 16** for the Dashboard; **Vite 6** for the Browser Extension.
*   **Communication**: **tRPC** for end-to-end type safety between Dashboard and Core.
*   **Persistence**: **SQLite + Drizzle ORM** for operational state; **LanceDB** for vector storage.
*   **Monorepo**: **pnpm + Turborepo**.

## 3. Current State vs. Absolute Ideal

| Dimension | Current State (Phase B) | The Absolute Ideal |
| :--- | :--- | :--- |
| **Tool Trust** | Manual "Evidence Lock" tracking. | **Cryptographic Attestation**: Every tool output carries a signed provenance record. |
| **MCP Discovery** | Manual config or probe-based testing. | **Registry Intelligence**: Autonomous ingestion and auto-certification of every public MCP server. |
| **Autonomy** | Human-in-the-loop "Director" agent. | **Self-Healing Swarms**: Agents that detect tool drift and auto-repair their own configurations. |
| **Memory** | L1-L3 database-backed. | **Cognitive Graph**: Zero-latency recall where memory is as fluid as local CPU cache. |

## 4. Recommendations
*   **Transition to Fastify**: While Express is stable, moving the core control plane to Fastify would improve performance for high-frequency tool calls.
*   **Automate Registry Ingestion**: Prioritize the "Registry Intelligence" pipeline to move from manual `TORMENTNEXUS_MASTER_INDEX.jsonc` updates to a dynamic, DB-backed catalog.
*   **Automate Registry Ingestion**: Prioritize the "Registry Intelligence" pipeline to move from manual `TORMENTNEXUS_MASTER_INDEX.jsonc` updates to a dynamic, DB-backed catalog.
