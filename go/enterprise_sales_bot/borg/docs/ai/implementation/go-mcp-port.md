# Go Sidecar MCP Port — Implementation Plan

## Objective
Port all remaining TS MCP engine features into the Go sidecar (`go/internal/mcp/`), achieving full MCP autonomy without TS dependency.

## Layer 1: MCP Features (22 files → Go)
**Goal**: Go sidecar fully owns MCP tool routing, discovery, caching, and inspection.

| # | TS File | Go Target | Priority |
|---|---------|-----------|----------|
| 1 | `cachedToolInventory.ts` | `go/internal/mcp/cached_inventory.go` | HIGH |
| 2 | `trafficInspector.ts` | `go/internal/mcp/traffic_inspector.go` | HIGH |
| 3 | `downstreamDiscovery.ts` | `go/internal/mcp/downstream_discovery.go` | HIGH |
| 4 | `discoveryPreflight.ts` | `go/internal/mcp/discovery_preflight.go` | HIGH |
| 5 | `catalogMetadata.ts` | `go/internal/mcp/catalog_metadata.go` | HIGH |
| 6 | `serverMetadataCache.ts` | `go/internal/mcp/server_metadata_cache.go` | MED |
| 7 | `namespaces.ts` | `go/internal/mcp/namespaces.go` | MED |
| 8 | `configStore.ts` | `go/internal/mcp/config_store.go` | MED |
| 9 | `clientConfigSync.ts` | `go/internal/mcp/client_config_sync.go` | MED |
| 10 | `compatibilityToolDefinitions.ts` | `go/internal/mcp/compat_tool_defs.go` | MED |
| 11 | `compatibilityToolRuntime.ts` | `go/internal/mcp/compat_tool_runtime.go` | MED |
| 12 | `directModeCompatibility.ts` | `go/internal/mcp/direct_mode_compat.go` | LOW |
| 13 | `toolLoadingDefinitions.ts` | `go/internal/mcp/tool_loading_defs.go` | LOW |
| 14 | `toolLoadingCompatibility.ts` | `go/internal/mcp/tool_loading_compat.go` | LOW |
| 15 | `toolAccessGuards.ts` | `go/internal/mcp/tool_access_guards.go` | LOW |
| 16 | `toolSetCompatibility.ts` | `go/internal/mcp/tool_set_compat.go` | LOW |
| 17 | `toolSelectionTelemetry.ts` | `go/internal/mcp/tool_selection_telemetry.go` | LOW |
| 18 | `legacyProxyMode.ts` | `go/internal/mcp/legacy_proxy_mode.go` | LOW |
| 19 | `mcpJsonConfig.ts` | `go/internal/mcp/mcp_json_config.go` | LOW |
| 20 | `savedScriptExecution.ts` | `go/internal/mcp/saved_script_exec.go` | LOW |
| 21 | `SessionToolWorkingSet.ts` | `go/internal/mcp/session_working_set.go` | MED |
| 22 | `SubmoduleManager.ts` | `go/internal/mcp/submodule_manager.go` | LOW |

## Layer 2: API Routers (34 files → Go httpapi)
**Goal**: All tRPC/REST endpoints served natively from Go.

| # | TS Router | Go Handler | Priority |
|---|-----------|-----------|----------|
| 1 | `graphRouter` | `go/internal/httpapi/graph_handlers.go` | HIGH |
| 2 | `knowledgeRouter` | `go/internal/httpapi/knowledge_handlers.go` | HIGH |
| 3 | `toolsRouter` | `go/internal/httpapi/tools_handlers.go` | HIGH |
| 4 | `researchRouter` | `go/internal/httpapi/research_handlers.go` | HIGH |
| 5 | `ragRouter` | `go/internal/httpapi/rag_handlers.go` | HIGH |
| 6 | `metricsRouter` | `go/internal/httpapi/metrics_handlers.go` | MED |
| 7+ | remaining 28 routers | various | LOW |

## Layer 3: Core Services (50 files → Go)
**Goal**: Business logic running natively.

| # | TS Service | Go Target | Priority |
|---|-----------|-----------|----------|
| 1 | `DeepResearchService` | `go/internal/ai/deep_research.go` | HIGH |
| 2 | `KnowledgeService` | `go/internal/memory/knowledge.go` | HIGH |
| 3 | `ToolRegistry` | `go/internal/toolregistry/registry.go` (expand) | HIGH |
| 4 | `MetricsService` | `go/internal/metrics/` | MED |
| 5+ | remaining 46 services | various | LOW |

## Execution Order
1. **Batch 1**: cachedToolInventory + trafficInspector + discoveryPreflight + downstreamDiscovery
2. **Batch 2**: catalogMetadata + serverMetadataCache + namespaces + configStore
3. **Batch 3**: compat layer + remaining MCP features
4. **Batch 4**: High-priority API routers (graph, knowledge, tools, research, rag)
5. **Batch 5**: Core services (DeepResearch, Knowledge, ToolRegistry, Metrics)
6. **Batch 6**: Remaining routers and services
