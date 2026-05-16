[PROJECT_MEMORY]

### 1. Identity & Vision: The AI Hypervisor
The project has evolved from its origin as "Borg" into a dual-brand architectural vision:
- **Nexus:** The underlying coordination kernel or "AI Hypervisor." It manages active memory, tool routing, and orchestration.
- **HyperCode:** The flagship, autonomous developer-facing coding product powered by the Nexus kernel.

The "AI Hypervisor" model treats AI models as compute resources and tools as peripheral drivers, with Nexus acting as the management layer that optimizes model selection, context management, and execution loops.

### 2. Current State (v1.0.0-alpha.56)
The project is currently in the transition between **Phase 1 (Active Memory)** and **Phase 2 (Autonomy Loop)**.
- **Phase 1 Status:** Complete. The foundational memory substrate is production-ready.
- **Phase 2 Status:** Initiated. Focus has shifted to self-healing reactors and the "execute-fix-verify-retry" autonomous loop.

### 3. Core Architectural Patterns
- **Kernel/Control Plane Split:**
    - **Kernel:** Deterministic execution, memory, and routing (being migrated toward `go/` and `@nexus/kernel`).
    - **Control Plane:** Dashboards, session management, and operator UI (`apps/web`, `packages/core`).
- **Active Memory Substrate:**
    - **Heat-Based Tiering:** Entries have a `heat_score` (0-100). Utility increases heat; time causes exponential decay (24h half-life).
    - **Feedback Loops:** Tool success/failure directly modifies the heat of the context used to achieve that outcome.
- **Provider Routing:**
    - Uses a waterfall fallback system. If one model/provider quota is exhausted, it automatically falls back to the next best available resource.
- **Progressive Tool Disclosure:**
    - Instead of flooding context with all tools, Nexus uses semantic ranking to disclose only relevant tools based on the active goal.

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
2.  **Fleet Management:** Extending Nexus to manage multiple concurrent "HyperCode" sessions with shared organizational memory.
3.  **Assimilation:** Systematically migrating high-performance logic (ranking, sync, memory) from TypeScript into the native Go kernel.

### 7. Governance & Intelligence
- **Supervisor/Council Pattern:** The system is designed to run under the supervision of an "Architect" or "Council" of models that verify plans before implementation.
- **Passive Harvesting:** The system automatically extracts facts and patterns from agent traffic to populate its L2 "Vault" memory without manual operator intervention.

---
*Last updated: Session v1.0.0-alpha.56*
