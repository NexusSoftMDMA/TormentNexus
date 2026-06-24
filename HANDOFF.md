# HANDOFF — Session 2026-06-24 (Dashboard Consolidation Phase 2 & 3)

## Summary

Consolidated multiple redundant dashboards in the Operator Dashboard. Specifically:
1. Merged `/dashboard/knowledge` and `/dashboard/brain` into a unified tabbed view under `/dashboard/brain`.
2. Unified `/dashboard/director`, `/dashboard/council`, `/dashboard/supervisor`, `/dashboard/squads`, and `/dashboard/swarm` into a single, multi-tabbed agent command center under `/dashboard/swarm`.
3. Cleaned up the side navigation bar menu items and checked for import/build correctness.

### What was done

1. **Brain & Knowledge Consolidation**:
   - Replaced `/dashboard/brain/page.tsx` with a Tabbed interface coordinating the visual symbol `KnowledgeGraph`, the URL ingestion forms, and the expert agents research/coder configuration.
   - Removed `/dashboard/knowledge` completely.
   - Redirected all remaining knowledge-base links to `/dashboard/brain`.

2. **Swarm & Agent Consolidation**:
   - Replaced `/dashboard/swarm/page.tsx` with a multi-tab workspace coordinating:
     - **Swarm & Mesh**: Orchestration settings and mesh operator registry.
     - **Squad Worktrees**: Spawn, chat, and kill buttons for parallel worktree agents, thought traces, and brain activity sheets.
     - **Director Office**: Strategy goals and plan steps.
     - **Supervisor Control**: High-level goal decomposition and supervisor execution logs.
     - **Council Debates**: Consensus session proposal and debate history.
     - **Telemetry & Neural Transcripts**: Real-time SSE streaming logs.
   - Created local normalizers `director-page-normalizers.ts` and `council-page-normalizers.ts` under `/dashboard/swarm/`.
   - Deleted the redundant folder structures for `/dashboard/director`, `/dashboard/council`, `/dashboard/supervisor`, and `/dashboard/squads`.
   - Updated [nav-config.ts](file:///c:/Users/hyper/workspace/tormentnexus/apps/web/src/components/mcp/nav-config.ts) to clean up the sidebar menu items.

3. **Versioning & Sync**:
   - Bumped monorepo version to `1.0.0-alpha.153` in the `VERSION` file.
   - Executed `node scripts/sync-versions.mjs` to synchronize all workspace `package.json` configurations.

4. **Verification**:
   - Verified that `pnpm -C apps/web build` compiles successfully with zero errors (total routes count reduced from 92 to 86, proving route consolidation worked).

5. **MCP Server `tools/list` Error Fixes**:
   - **Triage**: Discovered that when the stdio MCP server bootstrapped `MCPServer`, it initialized `PtySupervisor` which reads `session-supervisor.json` and attempts to restore all previously running sessions.
   - **Root Cause 1**: On Windows, node-pty throws a synchronous `Error: File not found: ...` if the target shell or command path is no longer valid. This was raised as an unhandled promise rejection/uncaught exception during Phase 2 startup, which caused the lightweight stdio server to exit or close the stream before replying to the client's `tools/list` request, resulting in the `"invalid request"` client-side error.
   - **Root Cause 2**: Standard MCP SDK request schemas (such as `ListToolsRequestSchema`, `CallToolRequestSchema`, etc.) are strict and automatically reject requests containing unexpected custom or client-specific metadata properties (like `_meta`), which standard clients like Antigravity pass in request parameters. This caused the server to return a `-32600` (Invalid Request) error before the handlers could process the requests.
   - **Root Cause 3**: The Go stdio MCP server subcommand (`tormentnexus mcp`) was returning response payloads for JSON-RPC notification messages (such as `notifications/initialized` where the request lacks an `id` property). This violates the JSON-RPC 2.0 / MCP specification and causes strict clients to throw errors.
   - **Fixes**:
     - Wrapped the `spawnProcess` call in [SessionSupervisor.ts](file:///c:/Users/hyper/workspace/tormentnexus/packages/core/src/supervisor/SessionSupervisor.ts) inside a `try/catch` block. If the spawn throws, the supervisor marks the session's status as `'error'` and captures the message in `session.lastError` instead of crashing the process.
     - Replaced the strict request schemas in both `packages/core/src/MCPServer.ts` and `packages/core/src/server-stdio.ts` with loose Zod schemas that accept any shape of `params` (including metadata). This prevents validation failures when clients inject tracing/routing metadata.
     - Updated `go/cmd/tormentnexus/mcp_server.go` to explicitly skip writing JSON-RPC response payloads when the request has a `nil` ID.
   - **Validation**:
     - Rebuilt `packages/core` and verified with the test script [test_client_env.js](file:///C:/Users/hyper/.gemini/antigravity/brain/a85dffdb-49fb-4e55-bdee-5fc548a4b08d/scratch/test_client_env.js) that `tools/list` now successfully returns all 54 active tools on stdio transport without any process crash or connection closures.
     - Recompiled the native Go binary and verified with [test_go_mcp_client.js](file:///C:/Users/hyper/workspace/tormentnexus/scratch/test_go_mcp_client.js) that it connects and lists all 45 tools successfully under the official SDK client.

### Current State
- **Workspace Build**: ✅ Compiling cleanly.
- **Monorepo Version**: `1.0.0-alpha.153`
- **MCP Server Stdio**: ✅ Returning tools/list response reliably.
- **Go Sidecar MCP**: ✅ Connects and lists tools correctly under the official client SDK.
- **Sidebar Count**: Clean and simplified with consolidated endpoints.
