# Core Daemons & Extraction Seams

This folder is the staging ground for the future TormentNexus binary topology. Following the `modular monolith first` strategy, these daemon components currently run within the single Node.js runtime process but have explicitly defined boundaries, bounded scopes, and no circular dependencies.

As these components mature and require independent scaling, crash isolation, or separate deploy targets, they will be easily decoupled into standalone binaries using TRPC, gRPC, or direct WebSockets over `stdio`.

## Daemon Responsibilities

1. **`tormentnexusd`**: The primary control-plane daemon. Owns top-level orchestration, provider/routing policy, operator state, and API routing.
2. **`hypermcpd`**: The MCP router, aggregator, and connection pool manager. Handles single-instance tool execution logic across multiple clients.
3. **`hypermcp-indexer`**: Background worker for MCP schema scraping, probing, metadata cache updates, and catalog ingestion.
4. **`hypermemd`**: Long-running memory state manager, vector DB bridge, and semantic search handler.
5. **`hyperingest`**: The background asynchronous worker for parsing `bobbybookmarks`, `bookmarks.txt`, and automatically importing AI CLI sessions.
6. **`hyperharnessd`**: The execution loop, handling actual PTY isolation, model execution, and subagent state.

*Agents: Please organize future service abstractions into these logical daemon domains rather than growing the generic `services/` folder.*
