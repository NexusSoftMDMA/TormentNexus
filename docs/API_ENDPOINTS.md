# TormentNexus API Endpoints

This document provides a comprehensive list of the HTTP API endpoints available in the TormentNexus Go sidecar (Port 4300).

## Metadata & Health
- `GET /api/service/connectivity`: Service health and discovery overview.
- `GET /health`: Basic service health check.
- `GET /version`: Build version for the Go sidecar.
- `GET /api/index`: Self-describing index of the Go sidecar API surface.
- `GET /api/health/server`: Health check alias for API consumers.

## Configuration
- `GET /api/config/status`: Path and config visibility snapshot for the sidecar and main workspace.
- `GET /api/config/list`: List config key/value entries.
- `GET /api/config/get`: Read one config key.
- `POST /api/config/upsert`: Upsert a config key.
- `POST /api/config/delete`: Delete a config key.
- `POST /api/config/update`: Update a config key.
- `GET /api/config/mcp-timeout`: Read MCP timeout.
- `POST /api/config/mcp-timeout/set`: Update MCP timeout.
- `GET /api/config/auth-providers`: Read auth providers.
- `GET /api/config/always-visible-tools`: Read always-visible tools.

## MCP (Model Context Protocol)
- `POST /api/mcp/client-sync`: IDE configuration synchronization (Claude/Cursor/VSCode).
- `GET /api/mcp/status`: MCP runtime status and pool state.
- `GET /api/mcp/servers`: Aggregated list of all MCP servers.
- `GET /api/mcp/tools`: Aggregated list of all MCP tools.
- `POST /api/mcp/tools/search`: Search tools with optional profile hinting.
- `POST /api/mcp/tools/call`: Execute an MCP tool.
- `POST /api/mcp/tools/predict-conversational`: Predict relevant tools based on a conversational prompt context.
- `POST /api/mcp/sync`: Trigger MCP server synchronization.
- `POST /api/mcp/decision/search`: Search tools using the MCP Decision System.
- `POST /api/mcp/decision/call`: Call a tool via the Decision System.

## Skills
- `GET /api/skills/list`: List all registered skills from the A2A skill registry.
- `GET /api/skills/get?id=<skillID>`: Get details for a specific skill by ID (agent URLs).
- `GET /api/skills/search?q=<query>`: Search skills by ID substring match.
- `GET /api/skills/load`: Load a skill into the active working set (stub).
- `GET /api/skills/unload`: Remove a skill from the active working set (stub).
- `GET /api/skills/list-loaded`: List currently loaded skills (stub).
- `GET /api/skills`: Legacy list of all skill entries.
- `GET /api/skills/read`: Read full skill content.
- `POST /api/skills/create`: Create a new skill.
- `POST /api/skills/save`: Save/update a skill.
- `GET /api/skills/summary`: Get a summary of all skills.
- `POST /api/skills/assimilate`: Assimilate content into a skill.

## Memory & Context
- `GET /api/memory/list`: List session memory entries.
- `POST /api/memory/add`: Add a memory entry.
- `POST /api/memory/search`: Contextual memory search.
- `GET /api/memory/contexts`: List saved memory contexts.
- `POST /api/memory/hydrate`: Hydrate memory from long-term storage.
- `GET /api/agent-memory/stats`: Agent-memory counts by tier.
- `POST /api/agent-memory/search`: Search agent-memory across namespaces and tiers.

## Agents & Swarm
- `GET /api/squad`: List squad members.
- `POST /api/squad/spawn`: Spawn a squad member.
- `POST /api/squad/chat`: Send a message to a squad member.
- `POST /api/swarm/start`: Initiate a multi-agent swarm mission.
- `GET /api/swarm/missions`: Swarm mission history.
- `POST /api/supervisor/decompose`: Decompose a high-level goal into tasks.
- `GET /api/supervisor/status`: Read supervisor status.

## Governance & Security
- `GET /api/api-keys`: List API keys.
- `POST /api/api-keys/create`: Create a new API key.
- `POST /api/api-keys/validate`: Validate an API key.
- `GET /api/audit`: List system audit logs.
- `POST /api/autonomy/set-level`: Set system autonomy level.

## DevOps & Operator
- `GET /api/git/status`: Read git repository status.
- `GET /api/git/log`: Read git log.
- `GET /api/submodules`: List git submodules and their state.
- `POST /api/scripts/execute`: Execute a saved operator script.
- `GET /api/infrastructure`: Read infrastructure daemon status.

## Code & Symbols
- `POST /api/code/exec`: Execute code in a secure sandbox.
- `GET /api/graph`: Get the repository dependency graph.
- `POST /api/lsp/find-symbol`: Locate symbols via Language Server Protocol.
- `GET /api/symbols/find`: Search for pinned symbols.

> **Note**: This is an abbreviated list. For a full, auto-generated list of all 600+ endpoints, use the `/api/index` endpoint on a running TormentNexus instance.
## Verified API Response Envelope
Standard response format: `{"success": true, "data": { ... }}`

## Integration Test Coverage
- Health and System Status: VERIFIED
- Native Tool Execution (/api/agent/tool): VERIFIED
- Skill Discovery: VERIFIED
- Script/Prompt Catalog: VERIFIED

## Detailed Native Tool Specifications
### ripgrep_search
- **Description**: High-speed recursive regex search.
- **Arguments**:
  - `pattern`: The regex to search for.
  - `path`: The directory to search in (default: ".").

### anyquery
- **Description**: SQL interface to file system and other data sources.
- **Arguments**:
  - `query`: The SQL query to execute.

### codemod
- **Description**: Execute large-scale codebase refactoring.
- **Arguments**:
  - `command`: The codemod command to run.

### service_connectivity
- **Description**: Probe and report health of upstream/downstream services.
- **Status**: VERIFIED

### client_sync
- **Description**: Generate MCP configuration for IDE clients.
- **Status**: VERIFIED
