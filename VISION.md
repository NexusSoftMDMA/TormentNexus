# Vision

## North star

borg is the **ultimate local-first control plane** for multi-agent workflows, Model Context Protocol (MCP) tooling, provider routing, session continuity, and operator observability.

It is building toward a future where one local system seamlessly coordinates the most critical parts of AI-driven software development: tools, models, sessions, context, subagents, and full visibility across the entire stack.

That future only becomes credible if the present system is reliable, truthful, and meticulously engineered for extreme robustness.

## The problem

AI workflows are fragmenting faster than they are becoming usable. A serious operator may now rely on:
- Multiple, disparate MCP servers with overlapping tools.
- Multiple AI provider accounts, quotas, API keys, and fallback chains.
- Multiple coding or session harnesses (VS Code, Cursor, Windsurf, Aider, OpenCode, Gemini CLI, Claude Code).
- Multiple memory surfaces (RAG, vector DBs, file-based contexts).
- Weak observability, logging, and crash-recovery across all of the above.

borg exists to eradicate that fragmentation by providing a unified, practical, and highly capable local substrate.

## Core thesis

The strongest version of borg is not a chaotic wrapper, but a **decision system and universal bridge**. It is the version that makes a messy local AI stack feel:
- **Calm**: Automatic orchestration, recovery, and fallback models.
- **Visible**: Comprehensive dashboards with real-time status, traffic inspection, and memory insight.
- **Composable**: Dynamic tool grouping, code execution sandboxes, and progressive tool disclosure.
- **Recoverable**: Persistent session history, crash isolation, and auto-restarting daemons.

## Practical product direction

### 1. The Ultimate MCP Control Plane
borg must become the absolute authority for:
- Registering, instantiating, and managing the lifecycle of MCP servers.
- Inspecting traffic, enforcing timeout rules, and providing single-instance pooling for multiple clients.
- Normalizing, grouping, and semantically searching tools (Tool RAG).
- Enabling code-mode execution for complex tool chains and deferred binary loading.
- Managing environment variables, secrets, and auth across all servers.

**The Definitive Library**: Over time, borg must maintain a definitive internal library of MCP servers aggregated from public catalogs, deduplicating overlapping entries, tracking provenance, and benchmarking tool implementations. Models will be able to reach any relevant tool in the ecosystem through one unified, local router.

### 2. Universal Tool Reach & First-Class Parity
Large Language Models are fine-tuned on the exact tool signatures used by the most popular coding environments. borg's mandate is **Absolute 1:1 Parity**. If a model expects `bash`, `glob`, `file_read`, or `grep_search` (from Claude Code, Codex, or Gemini CLI), borg provides a tool that is byte-for-byte identical in schema and behavior.

### 3. Model and Provider Routing
borg makes provider behavior legible and automated:
- Configurable fallback chains across models and APIs.
- Intelligent routing based on quota exhaustion, free tiers, and budget limits.
- Clear dashboards for auth-state, historical cost, and routing rules.

### 4. Session Continuity and Omniscient Memory
borg ensures an operator can continue work without rebuilding context from scratch:
- **Auto-Detection**: Automatically detects, imports, and parses sessions from *all* AI harnesses (IDEs, CLIs, Web) into durable memories.
- **Memory Subsystems**: A pluggable memory ecosystem (file-based, vector DB, RAG) with automatic context harvesting, chunking, reranking, and semantic search.
- **Web Integrations**: Browser extensions that inject MCP tools into web chats and export histories into the universal memory bank.

### 5. Multi-Agent Orchestration & Council Debate
borg provides an ecosystem for autonomous subagents and team-based modeling:
- Orchestrating a session where multiple frontier models (e.g., GPT, Gemini, Claude) take turns implementing, testing, and planning in a shared chatroom.
- Debate protocols, consensus voting, and autoDev loops under the supervision of a dedicated council.

### 6. Architectural Convergence (The Daemon Family)
borg is transitioning toward a robust, small family of focused binaries rather than a monolithic IPC bottleneck. The long-term topology includes:
- `borgd` — the primary control-plane daemon.
- `borgmcpd` — the MCP router, aggregator, and pool manager.
- `borgmemd` — the long-running memory and context daemon.
- `borgingest` — the background worker for batch imports (like BobbyBookmarks) and deduplication.
- `borgharnessborgharnessd` — the execution loop and isolation boundary.
- GUI Clients: CLI (`borg`), Web UI (`borg-web`), Desktop (`borg-native`).

## What borg is not optimizing for in v1

- Mass-market simplicity at the cost of operator control.
- Hosted SaaS-first workflows (we are explicitly local-first).
- Speculative platform expansion that outruns core reliability and truthfulness.

## Design principles

1. **Local first, but Cloud Aware**: Total local sovereignty with seamless remote deployment and syncing.
2. **Truth over hype**: Dashboards must reflect actual database rows and runtime state, not mocked UI scaffolds.
3. **Visible systems**: Everything—from LLM reasoning to MCP network traffic—must be inspectable.
4. **Interoperability over reinvention**: Wrap, adapt, and assimilate existing excellent tools (like ripgrep, SQLite, LanceDB) instead of rewriting them poorly.
5. **Continuity over novelty**: Focus heavily on state recovery, crash resilience, and session persistence.

6. **A2A Interoperability:** Borg discovers and is discovered by other agents over standard protocols.
7. **Graph Intelligence:** Traverse relationships between tools and tasks, not just keyword matches.
8. **Zero-Trust Execution:** Total isolation and provenance tracking for arbitrary operations.
9. **Total Observability:** Anomaly detection and flight-recording for reasoning logic.
10. **Tiered Intelligence:** Dynamic cost/performance optimization across local SLMs and cloud models.
11. **Accessibility-Native:** Reading the web through intent, not raw HTML DOM strings.
12. **Verified Code-as-Action:** Moving from generation to validation as the primary metric of success.
