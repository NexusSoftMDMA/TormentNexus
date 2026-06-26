# Go Sidecar API Reference

> **Context**: The `go/` workspace is an experimental sidecar and coexistence lane designed for operator visibility and read-parity. It currently runs alongside the primary Node.js control plane (`tormentnexusd`).

This document maps the HTTP API endpoints exposed by the experimental Go orchestrator port, explicitly classifying which routes are **Truthful Local Fallbacks** (reading local SQLite/Config files natively) versus **Bridge-Only Passthroughs** (forwarding requests to the Node.js control plane). 

## 1. Native Truthful Surfaces
These endpoints read actual configuration, database rows, or lock files natively using Go standard libraries or `database/sql`. They do not depend on the Node control plane being active.

* **`/api/runtime/locks`**: Scans the `.tormentnexus` directory for active `tormentnexus-go` and Node.js lock files.
* **`/api/config/status`**: Natively reads and parses `mcp.jsonc` and reports parsing success or failure.
* **`/api/runtime/status`**: Aggregates native SQLite counts (`mcp_servers`, `tools`, `bookmarks`) and merges them with config health.
* **`/api/mesh/peers`**: Queries the local peer table (if implemented) or broadcasts standard UDP discovery packets.

## 2. Bridge-Only Passthroughs (tRPC Wrappers)
These endpoints exist to maintain API parity for clients connecting to the Go sidecar on port `4000`, but they strictly proxy to the upstream Node.js TRPC control plane (default `3000` or defined in `.tormentnexus_startup_marker`). If the Node process is offline, these routes will return `503 Service Unavailable`.

### Sessions & Harnesses
* **`/api/sessions/supervisor/*`**: Proxies to `trpc.supervisor.*`
* **`/api/sessions/imported/*`**: Proxies to `trpc.sessionExport.*`

### MCP & Tools
* **`/api/mcp/tools/auto-call`**: Wraps the `auto_call_tool` meta-tool for one-shot execution via `mcpRouter`.
* **`/api/mcp/registry/*`**: Proxies `mcpServers.registrySnapshot` and dynamic installation endpoints.
* **`/api/mcp/traffic`**: Forwards historical traffic logs directly from the Node memory buffer.

### Billing & Providers
* **`/api/billing/status`**: Proxies `trpc.billing.getStatus`.
* **`/api/billing/provider-quotas`**: Proxies `trpc.billing.getProviderQuotas`.
* **`/api/billing/fallback-history`**: Proxies `trpc.billing.getFallbackHistory`.

### Memory & Context
* **`/api/memory/context/save`**: Proxies `trpc.memory.saveContext`.
* **`/api/agent-memory/search`**: Proxies `trpc.memory.searchAgentMemory`.
* **`/api/context/prune`**: Proxies to the Node-side ContextManager.

### Multi-Agent Council
* **`/api/council/rotation/*`**: Proxies room management, plan-mode execution, and round-robin turn advancing.
* **`/api/swarm/*`**: Proxies debate, consensus voting, and subagent spawning controls.
* **`/api/autodev/start-loop`**: Proxies autonomous coding execution.

## 3. Implementation Rules for New Go Routes
If you are extending the Go sidecar, you must adhere to the **Stabilization-First** architectural rule:
1. **Reads**: Try to implement natively (Truthful Local Fallback) if the data is safely readable via SQLite or static files without complex serialization logic.
2. **Writes/Execution**: ALWAYS implement as a Bridge-Only Passthrough to the Node control plane unless specifically migrating ownership of a daemon layer (e.g., `tormentnexusingest`). Do not duplicate execution logic in Go while the Node daemon still owns it.
