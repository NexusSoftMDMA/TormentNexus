[PROJECT_MEMORY]

### 1. Core Identity: TormentNexus & TormentNexus
The project has stabilized under a dual-brand **AI TormentNexus** architecture:
- **TormentNexus (The Kernel):** The underlying coordination engine written in Go. It manages the "Operating System" layer for AI, including active memory, semantic tool routing, and multi-model orchestration.
- **TormentNexus (The Product):** The flagship developer-facing autonomous coding runtime and observation dashboard powered by the TormentNexus kernel.
- **The Philosophy:** Treats AI models as ephemeral **compute resources** and tools as **peripheral drivers**. TormentNexus serves as the deterministic management layer that optimizes context windows, minimizes cost, and ensures execution reliability.

### 2. Architectural Paradigm: The Modular Monolith
The system enforces a strict "Source of Truth" hierarchy:
- **TormentNexus Kernel (Go):** Absolute authority for orchestration, L1/L2 memory, high-performance BM25/Cosine ranking, and Model Context Protocol (MCP) synchronization. All logic resides in \`go/internal/\`.
- **Control Plane (TypeScript/Next.js):** The "Observation Deck" responsible for visual state representation, dashboard visualization, and high-level agent session management. It bridges requests to the Go sidecar via tRPC and REST APIs.
- **Database:** Standardized on **SQLite with sqlite-vec**. No external dependencies (Postgres, Redis) are permitted to maintain a "local-first," portable footprint.

### 3. Active Memory Substrate: Biological Tiering
TormentNexus implements a tiered memory system designed to mimic biological relevance:
- **Tiers:** L1 (Working Scratchpad/Ephemeral), L2 (Long-Term Vault/Persistent), and L3 (Cold Archive).
- **Heat-Based Tiering (0-100):** Every memory tracks its "temperature." Functional utility (access or success) increases Heat.
- **Exponential Decay:** Heat decays over time with a 24-hour half-life ($\approx$ 0.0288 per hour) to keep the working context lean.
- \*\*Promotion/Demotion:\*\* High-heat entries (>80) move from Working to Long-Term. Low-heat entries (<20) are demoted to the archive to maintain index performance.
- **Outcome Feedback:** The kernel records tool execution success/failure to reinforce the heat of relevant context, enabling the system to "learn" from its execution history.

### 4. Progressive Disclosure: Context Hygiene
To prevent "Context Blowout" and minimize token usage, TormentNexus employs semantic asset discovery:
- **Ranked Discovery:** Tools and "Skills" (runbooks) are ranked using BM25 and Cosine similarity against the \`activeGoal\`.
- **Pre-loading:** High-confidence assets are silently auto-loaded into the model's context before explicit requests.
- **Token Budgeting:** Strict soft caps are enforced for the L1 Scratchpad to ensure model stability and responsiveness.

### 5. Autonomous Operations: The Immune System
The system features a self-verifying "Immune System" known as **The Healer**:
- **Autonomous Healer Loop:** Implements a multi-turn \`Diagnose -> Fix -> Verify -> Retry\` cycle.
- **Self-Verification:** The kernel automatically executes tests (\`vitest\`) or type-checks (\`tsc\`) to verify its own fixes before committing them.
- **StopHooks & IdleHealer:** Supports \`agent:stop_healing\` signals to prevent interference and triggers background diagnostics during system inactivity.

### 6. Fleet Orchestration: Collective Intelligence
The kernel supports managing multiple concurrent TormentNexus sessions:
- **Fleet Manager:** Tracks active session PIDs and health in real-time.
- **Traffic Observer:** Passively harvests technical facts and lessons learned from A2A (Agent-to-Agent) signals.
- **Shared Memory:** Technical discoveries in one session are automatically indexed in the L2 Vault and shared globally across the local fleet.

### 7. Technical Decisions & Guardrails
- **Shell Hardening:** For security, \`child_process.exec\` is strictly prohibited. All commands must use \`spawn\` with tokenized argument arrays and \`shell: false\`.
- **Sync Authority:** All MCP configuration detection (Claude, Cursor, VS Code) resides in the Go kernel to maintain 100% environment authority and parity.
- **Stream Stability:** Exponential backoff for tRPC subscriptions and history-aware message buffering ensure signal continuity during network drops.
- **Consensus Loop:** Enforces a strict multi-model \`Planner -> Checker -> Implementer -> Critic\` turn cycle natively in Go.
## 1. Dynamic Tool Discovery & Registry (ToolRAG)
**The Problem:** The MCP ecosystem has over 25,000 tools. Static loading exhausts the LLM context window (imposing a 32% token overhead penalty).
**The Converging Solution:** "RAG but for tools." Embed tool names only and fetch full JSON schemas strictly on-demand.
*   **TormentNexus Implementation:** Build `tormentnexus-tool-registry` to index all discovered MCP servers, embed schemas using SQLite Vector Search, and inject only the 3-5 most relevant tools per query.

### 3. Core Architectural Patterns
- **Kernel/Control Plane Split:**
    - **Kernel:** Deterministic execution, memory, and routing (being migrated toward `go/` and `@tormentnexus/kernel`).
    - **Control Plane:** Dashboards, session management, and operator UI (`apps/web`, `packages/core`).
- **Active Memory Substrate:**
    - **Heat-Based Tiering:** Entries have a `heat_score` (0-100). Utility increases heat; time causes exponential decay (24h half-life).
    - **Feedback Loops:** Tool success/failure directly modifies the heat of the context used to achieve that outcome.
- **Provider Routing:**
    - Uses a waterfall fallback system. If one model/provider quota is exhausted, it automatically falls back to the next best available resource.
- **Progressive Disclosure:**
    - Tools and Skills are ranked and disclosure is limited to the most relevant entries based on the active goal.

### 4. Monorepo Structure & Module Roles
- **`packages/core`:** The central hub ("Brain") of the TypeScript control plane. It hosts tRPC routers, session logic, and bridges to the Go sidecar.
- **`packages/memory`:** The implementation layer for LanceDB and vector-based storage.
- **`go/`:** The Go Sidecar (Port 4300). Currently serves as a high-performance state authority and BM25 ranking engine, mirroring and bridging TypeScript services.
- **`apps/web`:** The primary operator dashboard for managing sessions and visualizing the knowledge graph.
- **`packages/tools`:** Contains functional tool implementations (Read, Write, Shell, etc.) shared across CLI and Web surfaces.

### 5. Technical Decisions & Constraints
- **Shell Hardening:** `child_process.exec` is strictly prohibited. All command execution must use `spawn` with tokenized argument arrays and `shell: false`.
- **Environment:** Standardized on Node 24 and Go 1.24.3. Port 443 is restricted; local caches/binaries must be used for dependency management.
- **Version Authority:** Versioning is synchronized globally. The current baseline is `1.0.0-alpha.57`.

### 6. Roadmap: The Autonomy Path
The next immediate milestones involve:
1.  **Autonomous Healer:** Multi-turn fix-verify-retry loop (Implemented).
2.  **Fleet Management:** Extending TormentNexus to manage multiple concurrent "TormentNexus" sessions with shared organizational memory.
3.  **Assimilation:** Systematically migrating high-performance logic (ranking, sync, memory) from TypeScript into the native Go kernel.

### 1. Identity & Vision: The AI TormentNexus
The project has evolved from its origin as "TormentNexus" into a dual-brand architectural vision:
- **TormentNexus:** The underlying coordination kernel or "AI TormentNexus." It manages active memory, tool routing, and orchestration.
- **TormentNexus:** The flagship, autonomous developer-facing coding product powered by the TormentNexus kernel.

The "AI TormentNexus" model treats AI models as compute resources and tools as peripheral drivers, with TormentNexus acting as the management layer that optimizes model selection, context management, and execution loops.

### 2. Current State (v1.0.0-alpha.56)
The project is currently in the transition between **Phase 1 (Active Memory)** and **Phase 2 (Autonomy Loop)**.
- **Phase 1 Status:** Complete. The foundational memory substrate is production-ready.
- **Phase 2 Status:** Initiated. Focus has shifted to self-healing reactors and the "execute-fix-verify-retry" autonomous loop.

### 3. Core Architectural Patterns
- **Kernel/Control Plane Split:**
    - **Kernel:** Deterministic execution, memory, and routing (being migrated toward `go/` and `@tormentnexus/kernel`).
    - **Control Plane:** Dashboards, session management, and operator UI (`apps/web`, `packages/core`).
- **Active Memory Substrate:**
    - **Heat-Based Tiering:** Entries have a `heat_score` (0-100). Utility increases heat; time causes exponential decay (24h half-life).
    - **Feedback Loops:** Tool success/failure directly modifies the heat of the context used to achieve that outcome.
- **Provider Routing:**
    - Uses a waterfall fallback system. If one model/provider quota is exhausted, it automatically falls back to the next best available resource.
- **Progressive Tool Disclosure:**
    - Instead of flooding context with all tools, TormentNexus uses semantic ranking to disclose only relevant tools based on the active goal.

### 4. Monorepo Structure & Module Roles
- **`packages/core`:** The central hub ("Brain") of the TypeScript control plane. It hosts tRPC routers, session logic, and bridges to the Go sidecar.
- **`packages/memory`:** The implementation layer for LanceDB and vector-based storage.
- **`go/`:** The Go Sidecar (Port 4300). Currently serves as a high-performance state authority and BM25 ranking engine, mirroring and bridging TypeScript services.
- **`apps/web`:** The primary operator dashboard for managing sessions and visualizing the knowledge graph.
- **`packages/tools`:** Contains functional tool implementations (Read, Write, Shell, etc.) shared across CLI and Web surfaces.

### 5. Technical Decisions & Constraints
- **Shell Hardening:** `child_process.exec` is strictly prohibited. All command execution must use `spawn` with tokenized argument arrays and `shell: false`.
- **Environment:** Standardized on Node 24 and Go 1.24.3. Port 443 is restricted; local caches/binaries must be used for dependency management.
- **Version Authority:** Versioning is synchronized globally. The current baseline is `1.0.0-alpha.56`.

### 6. Roadmap: The Autonomy Path
The next immediate milestones involve:
1.  **The Healer Loop:** Implementing the full `execute-fix-verify-retry` autonomous cycle within the `HealerReactor`.
2.  **Fleet Management:** Extending TormentNexus to manage multiple concurrent "TormentNexus" sessions with shared organizational memory.
3.  **Assimilation:** Systematically migrating high-performance logic (ranking, sync, memory) from TypeScript into the native Go kernel.

### 7. Governance & Intelligence
- **Supervisor/Council Pattern:** The system is designed to run under the supervision of an "Architect" or "Council" of models that verify plans before implementation.
- **Passive Harvesting:** The system automatically extracts facts and patterns from agent traffic to populate its L2 "Vault" memory without manual operator intervention.

---
*Last updated: Session v1.0.0-alpha.56*
# AI TormentNexus (TormentNexus) - Comprehensive Architectural Memory

This document summarizes the foundational architecture, established patterns, and strategic decisions of the project as of version **1.0.0-alpha.56**.

## 1. Strategic Identity: TormentNexus & TormentNexus
The project has successfully pivoted from "TormentNexus" to a dual-brand infrastructure model:
*   **TormentNexus (The Kernel/TormentNexus):** The underlying coordination runtime and "AI TormentNexus." It treats LLMs as "guest operating systems" and manages the low-level memory, routing, and execution buses.
*   **TormentNexus (The Product):** The user-facing, local-first autonomous coding environment powered by the TormentNexus kernel.

## 2. Active Tiered Memory Substrate (Implemented Phase 1)
*   **Heat Scoring (0-100):** Every memory entry tracks utility. Heat increases on access and decays exponentially (24-hour half-life).
*   **Tiered Hierarchy:** L1 (Working Memory, heat > 80) is promoted to context; L2 (Vault) is for semantic recall.
*   **Tool-Outcome Feedback:** `MemoryManager.recordToolOutcome()` boosts the heat of successful patterns and demotes failures.

## 3. "Kernel / Control Plane" Topology
*   **/kernel**: Deterministic brain (runtime, memory, router).
*   **/control-plane**: Observer layer (UI, Telemetry).

## 4. State Authority & The Sidecar Pattern
*   **Go Sidecar (Port 4300):** State authority and BM25 ranking.
*   **TS Bridge (Port 4100):** Primary control-plane bridge and tRPC host.

## 5. Intelligence Management: Progressive Disclosure
*   **Decision System:** Ranked discovery and LRU eviction ensures only 3-5 tools/skills are in the active working set.

## 6. Hardened Execution & Security
*   **Standard:** Tokenized argument arrays with `shell: false` for all command executions.
*   **Parity Principle:** 1:1 behavioral and schema parity for tools expected by proprietary models (e.g., Claude Code).

---
*Last updated: v1.0.0-alpha.56*
### Meta-Protocol for Future Sessions
1.  **Truth over Fiction:** Dashboards must reflect real state.
2.  **Autonomous Momentum:** Proceed through Phase 2 (Autonomy/Self-Healing).
3.  **Documentation Sync:** Every version bump syncs all meta files and manifests.
