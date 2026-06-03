# TormentNexus Core Assimilation Guide

## System Overview
`cli-orchestrator` has undergone a massive architectural overhaul to transition from a standalone monolithic orchestration tool into a resilient, self-healing, distributed sub-component ready for direct assimilation into the **TormentNexus Core**.

## Core Capabilities for TormentNexus

### 1. Distributed PTY Sidecar Architecture (The "TormentNexus-Harness")
# tormentnexus Core Assimilation Guide

## System Overview
`cli-orchestrator` has undergone a massive architectural overhaul to transition from a standalone monolithic orchestration tool into a resilient, self-healing, distributed sub-component ready for direct assimilation into the **tormentnexus Core**.

## Core Capabilities for tormentnexus

### 1. Distributed PTY Sidecar Architecture (The "tormentnexus-Harness")
The system is no longer a single point of failure. It uses a **Distributed Sidecar Model**.
- **Detached Execution:** All managed AI terminal sessions (like `gemini-cli`, `github-copilot-cli`, `aider`) are launched as detached `terminal-sidecar.ts` processes.
- **Zero-Downtime Re-attachment:** If the Orchestrator goes down, the sidecar terminals keep running. Upon restart, the Orchestrator discovers the active sidecar ports and seamlessly re-links via local TCP sockets.

### 2. Universal CLI Support (The Fleet)
The harness is completely agnostic and supports an expansive list of CLI agents natively.
Supported binaries include:
- `aider`, `amazon-q`, `amazon-q-developer`, `amp-code`, `auggie`, `az-openai`
- `claude-code`, `code-codex`, `codebuff`, `codemachine`, `codex`, `copilot`
- `crush`, `factory`, `gemini`, `goose`, `grok`, `kilo-code`, `kimi`
- `mistral`, `opencode`, `pi`, `qwen`, `rowboatx`, `rovo`, `trae`, `warp`

### 3. Dynamic Environment Injection (`TORMENTNEXUS_CTRL`)
TormentNexus Core can inject dynamic secrets and variables directly into running terminal sessions without restarting them. By passing `TORMENTNEXUS_CTRL` JSON packets over the sidecar socket, variables are exported directly into the PTY shell, allowing for live API key rotation.
### 3. Dynamic Environment Injection (`TORMENTNEXUS_CTRL`)
tormentnexus Core can inject dynamic secrets and variables directly into running terminal sessions without restarting them. By passing `TORMENTNEXUS_CTRL` JSON packets over the sidecar socket, variables are exported directly into the PTY shell, allowing for live API key rotation.

### 4. Binary State Checkpointing
The `CheckpointService` maintains the "last known context" of all running sessions, capturing the active task and the terminal output buffer. When re-attaching, it injects a "Recovery Hint" so the LLM resumes with full situational awareness.

### 5. Swarm Intelligence & Hierarchical Routing
Tasks are parsed from natural language (or Mermaid diagrams) into hierarchical, dependency-aware task graphs. The `SmartPilot` distributes tasks across the CLI fleet in parallel based on tool strengths. `CouncilHierarchy` enables specialized sub-councils (e.g., Performance Council, Security Council) to debate specific types of tasks before pushing decisions back up to the Supreme Council.

### 6. Collective Memory
A persistent SQLite "Knowledge Base" is maintained by the `CollectiveMemoryService`. Any sub-agent across any session can store and retrieve facts, sharing knowledge globally across the TormentNexus hive-mind.

### 7. Autonomous Self-Maintenance & Evolution
- `AutonomousMaintenanceService` cleans orphaned processes, optimizes the DB (`VACUUM`/`ANALYZE`), and prunes old checkpoints.
- `SelfEvolutionService` adjusts supervisor trust weights dynamically based on historical debate outcomes and can spawn high-privilege meta-sessions to rewrite the TormentNexus orchestration source code.

## Assimilation Next Steps
1. **Repository Merge:** Move the `cli-orchestrator` package into the TormentNexus Core monorepo structure.
2. **Global Telemetry:** Hook the `wsManager` broadcast events to the TormentNexus Core global telemetry bus.
3. **Database Unification:** Migrate the `bun:sqlite` implementation to TormentNexus's central high-availability datastore if required.
4. **Service Discovery:** Replace hardcoded `127.0.0.1` PTY ports with TormentNexus Core's dynamic service mesh/discovery network.
A persistent SQLite "Knowledge Base" is maintained by the `CollectiveMemoryService`. Any sub-agent across any session can store and retrieve facts, sharing knowledge globally across the tormentnexus hive-mind.

### 7. Autonomous Self-Maintenance & Evolution
- `AutonomousMaintenanceService` cleans orphaned processes, optimizes the DB (`VACUUM`/`ANALYZE`), and prunes old checkpoints.
- `SelfEvolutionService` adjusts supervisor trust weights dynamically based on historical debate outcomes and can spawn high-privilege meta-sessions to rewrite the tormentnexus orchestration source code.

## Assimilation Next Steps
1. **Repository Merge:** Move the `cli-orchestrator` package into the tormentnexus Core monorepo structure.
2. **Global Telemetry:** Hook the `wsManager` broadcast events to the tormentnexus Core global telemetry bus.
3. **Database Unification:** Migrate the `bun:sqlite` implementation to tormentnexus's central high-availability datastore if required.
4. **Service Discovery:** Replace hardcoded `127.0.0.1` PTY ports with tormentnexus Core's dynamic service mesh/discovery network.

Assimilation Complete. Resistance is futile.
