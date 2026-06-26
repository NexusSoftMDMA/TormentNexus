### Project Architecture & Design Philosophy
TormentNexus is built on a **Dual-Core Architecture** designed for maximum autonomy and performance:
1.  **Go Kernel (TormentNexus Sidecar)**: Acts as the high-performance Control Plane. It handles low-level system operations, tool execution, and long-term memory (L2 Vault) management. It is designed to be CGo-free using `modernc.org/sqlite` for portability.
2.  **TypeScript Dashboard (Next.js)**: Serves as the user-facing observation deck and orchestration layer, communicating with the Go kernel via tRPC and REST APIs.

### Core Patterns & Decisions
-   **Native Tool Assimilation**: The project systematically replaces external MCP servers (Node/Python/Docker) with native Go reimplementations. This reduces runtime overhead and dependency hell. 
-   **Three-Tier Skill Loading**: To preserve LLM context windows, skills are loaded progressively:
    1.  *Manifest*: Basic metadata for discovery.
    2.  *Summary*: Technical overview for planning.
    3.  *Full Content*: Detailed logic for execution.
-   **SQLite-First Persistence**: All state—from tool catalogs to the new Skill Registry—is being migrated to SQLite. We leverage **FTS5** for high-performance full-text search, enabling the "Cognitive Control Plane" to search through thousands of technical assets in milliseconds.
-   **Principle Directive: Continuous Autonomous Execution**: The system operates on a "Total Autopilot" protocol, requiring sequential implementation, atomic commits, and automatic documentation updates (`CHANGELOG`, `ROADMAP`, `MEMORY`) without pausing for human confirmation.

### Key Structural Decisions
-   **A2A (Agent-to-Agent) Skill Registry**: A global singleton that maps skill IDs to agent URLs, facilitating swarm-based task delegation.
-   **Zero-Conflict Policy**: All git conflict markers must be purged immediately during repository synchronization to maintain a "green" build baseline.
-   **Indented UI Components**: A strict 4-space indentation standard for dashboard components to ensure consistency across the Next.js workspace.

---

### Implementation Progress: Step 1 - Wiring the Indexer and Schema Initialization

I have already implemented the `InitSchema()` logic in `go/internal/skillregistry/db_registry.go`. Now, I will wire the `Indexer` and `DBRegistry` correctly into the `Server` lifecycle in `go/internal/httpapi/server.go`.