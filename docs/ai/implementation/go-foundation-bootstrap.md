# Go Foundation Bootstrap Implementation Notes

## What was added

### Code
- `foundation/pi/foundation.go`
  - Pi-derived foundation specification
  - thinking levels
  - transport and delivery modes
  - run-event vocabulary
  - built-in tool descriptors

- `foundation/pi/runtime_types.go`
  - tool input/output types
  - run event payloads
  - truncation/detail structures

- `foundation/pi/tools_native.go`
  - native `read`, `write`, `edit`, and `bash` implementations
  - truncation behavior
  - exact-name tool handlers

- `foundation/pi/runtime.go`
  - evented runtime for native tool execution
  - session-linked tool execution
  - session creation/list/load/fork helpers

- `foundation/pi/session.go`
  - JSONL-backed session metadata and tree entries
  - create/load/save/list/fork operations

- `foundation/compat/types.go`
  - exact tool contract types
  - parity maturity model

- `foundation/compat/catalog.go`
  - thread-safe contract registry

- `foundation/compat/default_catalog.go`
  - Pi-compatible default tool contract set (`read`, `write`, `edit`, `bash`)
  - updated to `native` maturity for the default tool set

- `foundation/assimilation/inventory.go`
  - upstream assimilation inventory covering imported toolchains and TormentNexus

- `foundation/assimilation/summary.go`
  - category summarization helpers

- `cmd/foundation.go`
  - CLI inspection for foundation spec, inventory, and tools
  - native execution surface for exact-name tools
  - native session create/list/show/fork commands
  - native repo-map generation command
  - native orchestration planning command
  - provider visibility, route-selection, and execution-preparation commands

- `foundation/repomap/repomap.go`
  - Aider-inspired native repo map baseline with graph-ranking groundwork
  - ranked file ordering using mentioned files/idents
  - lightweight definition/reference propagation across files
  - symbol extraction for common source forms

- `foundation/repomap/repomap_test.go`
  - repo map output validation
  - ranking validation for mention-based and graph-based prioritization

- `foundation/orchestration/planner.go`
  - orchestration planning primitives that combine task inference, provider execution preparation, and optional repo-map context

- `foundation/orchestration/planner_test.go`
  - validates orchestration planning results and repo-map inclusion

- `foundation/orchestration/daemon_plan.go`
  - foundation-backed daemon sweep planning for queue actions and telemetry summaries

- `foundation/orchestration/daemon_plan_test.go`
  - validates daemon sweep planning decisions

- `foundation/orchestration/webhook_plan.go`
  - foundation-backed webhook planning for queue actions and telemetry summaries

- `foundation/orchestration/webhook_plan_test.go`
  - validates webhook-to-action planning

- `tools/registry.go`
  - top-level tool registry now exposes exact-name Pi-compatible tools via the native foundation runtime
  - repomap is available from the legacy registry surface as a native foundation-backed tool
  - tool schemas are now forwarded instead of using one placeholder schema for every tool

- `tools/repomap.go`
  - legacy wrapper now delegates to `foundation/repomap`

- `mcp/client.go`
  - top-level MCP client now uses the adapter seam for connection status, tool hint listing, and mediated tool-call routing

- `mcp/manager.go`
  - top-level MCP manager now uses the adapter seam for configured server discovery, startup, and routed tool-call mediation

- `mcp/config.go`
  - top-level MCP config now wraps adapter-owned parsing instead of duplicating config logic

- `mcp/mcphost.go`
  - defensive guard added for empty MCP binary path to avoid nil-process panics in tests

- `cmd/foundation_http.go`
  - foundation-backed helper layer for execution, sessions, planning, repomap, adapters, foundation-backed file reads, foundation-backed MCP mediation, and provider-route selection

- `cmd/serve.go`
  - operator HTTP surface now exposes foundation-backed endpoints under `/api/v1/foundation/*`
  - `/api/v1/foundation/plan` now exposes foundation-backed orchestration planning
  - `/api/v1/foundation/providers*` now exposes provider visibility, route-selection, and execution-preparation behavior
  - `/api/v1/foundation/mcp/*` now exposes adapter-backed MCP tool listing and mediated call preparation
  - `/fs/read` now routes through the native foundation `read` tool instead of direct file reads
  - foundation-backed orchestration entrypoints continue to replace direct placeholder logic incrementally

- `tui/slash.go`
  - TUI slash-command handling now exposes foundation-backed `/plan`, `/repomap`, `/providers`, `/adapters`, and `/mcp`
  - `/clear` now resets the director cleanly instead of only clearing history text

- `tui/foundation_bridge.go`
  - non-slash TUI request helpers now wrap normal prompt handling and shell proposal generation in foundation-aware behavior

- `tui/slash_test.go`
  - validates foundation-backed slash-command planning, provider/adapter introspection, and repo-map behavior

- `tui/foundation_bridge_test.go`
  - validates foundation-backed prompt/shell helper behavior

- `agent/agent.go`
  - top-level agent now advertises the native exact-name tools preferentially
  - OpenAI tool registration now uses per-tool schemas instead of one fake generic schema
  - system prompt now incorporates TormentNexus/TormentNexus and provider adapter context

- `agent/pipe.go`
  - pipe processing now uses provider execution-preparation hints before invoking the agent

- `agents/provider_stub.go` and `agents/provider.go`
  - top-level provider stubs now consume provider execution-preparation hints instead of returning purely static placeholder text

- `agents/director.go`
  - top-level director now records orchestration plans and injects plan context into provider calls

- `agent/orchestrator.go`
  - top-level orchestrator now builds plans from `foundation/orchestration` instead of relying only on placeholder LLM planning

- `orchestrator/webhooks.go`
  - webhook handling now uses foundation-backed webhook planning for queued actions and telemetry summaries

- `orchestrator/daemon_loop.go`
  - daemon sweep now uses foundation-backed daemon planning before queueing actions

- `orchestrator/orchestration_bridge.go`
  - daemon/autodrive bridge now converts foundation plans into execution objectives for sandboxed runs

- `foundation/adapters/tormentnexus.go`
  - first TormentNexus/TormentNexus adapter seam for the Go foundation
  - exposes assimilation status, memory context, provider status, MCP config visibility, and adjacent TormentNexus repo discovery

- `foundation/adapters/providers.go`
  - provider adapter seam for current provider/model visibility
  - detects available providers from config/env
  - probes Ollama models when relevant

- `foundation/adapters/provider_routing.go`
  - provider-route selection groundwork based on task type, cost preference, and local-execution preference
  - provides the first shared route-selection logic for CLI and HTTP surfaces

- `foundation/adapters/provider_execution.go`
  - provider-execution preparation seam that combines inferred task type, provider status, route selection, and execution hints

- `foundation/adapters/mcp_config.go`
  - adapter-owned MCP config parsing to avoid circular coupling
  - normalizes server command/env visibility for foundation consumers

- `foundation/adapters/mcp.go`
  - MCP adapter seam for configured server discovery, tool hints, route hints, mediated tool-call preparation, and configured-server startup

- `foundation/adapters/tormentnexus_test.go`
  - validates adapter status, routing, and system-context construction

- `foundation/adapters/providers_test.go`
  - validates provider status/context construction

- `foundation/adapters/provider_routing_test.go`
  - validates provider-route selection groundwork

- `foundation/adapters/provider_execution_test.go`
  - validates provider execution-preparation behavior

- `foundation/adapters/mcp_test.go`
  - validates MCP adapter status, tool hints, and route calls

- `tools/registry_test.go`
  - verifies exact Pi tools and repomap are present in the registry

- `foundation/pi/tool_snapshot_test.go`
  - snapshot-style verification for baseline tool results

- `agent/agent_test.go`
  - verifies top-level OpenAI tool registration exposes exact-schema tool definitions
  - verifies TormentNexus adapter presence on the top-level agent

- `mcp/client_test.go`
  - verifies MCP client tool hint listing and mediated call routing through the adapter seam

- `mcp/manager_test.go`
  - verifies MCP manager configured-tool listing, mediated call routing, and missing-server handling

- `cmd/foundation_http_test.go`
  - verifies foundation-backed execution/session/planning/repomap/adapter helper behavior used by HTTP surfaces, including MCP mediation and provider-route helpers

- `agents/director_test.go`
  - verifies orchestration plan state and planned response decoration

- `agent/orchestrator_test.go`
  - verifies foundation-backed orchestration plan building

- `orchestrator/orchestration_bridge_test.go`
  - verifies daemon/autodrive objective generation from foundation plans

### Documentation
- requirements, design, planning, implementation, and testing documents under `docs/ai/`

## Why this phase is useful
This phase still does not pretend to have completed the full port, but it now moves beyond pure scaffolding and establishes a truthful native baseline:
- one place for exact tool contracts,
- one place for the Pi-derived harness contract,
- one place for native default tool execution,
- one place for JSONL-backed native sessions,
- one place for the upstream assimilation inventory,
- one CLI surface to inspect and exercise those decisions,
- one documentation trail explaining the chosen architecture.

## Important baseline observations

### Existing codebase truthfulness gap
The current Go code advertises broad parity in some command descriptions, but several implementations are still placeholder-level. The new foundation work is intentionally separating:
- **declared compatibility** from
- **actual native implementation**.

### Existing test baseline issues
Before this phase, `go test ./...` already failed for unrelated reasons:
- `aider/tests/fixtures/languages/go/test.go` has an unused import.
- `mcp/mcphost_test.go` is out of sync with the host API.
- `orchestrator` panics because SQLite is registered twice.

These issues were observed and documented, not silently ignored or misrepresented as introduced by the new foundation work.

## Validation added in this phase
- native runtime tests for `read`, `write`, `edit`, and `bash`
- session persistence/list/fork tests
- ordered runtime event tests
- parity/truncation tests for `read` and `bash`
- snapshot-style tool result verification
- repo map generation and ranking tests
- orchestration planning tests
- daemon sweep planning tests
- webhook planning tests
- daemon/autodrive orchestration bridge tests
- top-level tool registry tests confirming native exact-name tool exposure
- top-level agent tool-schema registration tests
- TormentNexus/TormentNexus adapter seam tests
- provider adapter seam tests
- provider-route selection tests
- provider execution-preparation tests
- MCP adapter seam and top-level MCP package tests
- foundation-backed HTTP helper tests, including MCP mediation and provider-route helpers
- foundation-backed TUI slash-command tests
- foundation-backed TUI prompt/shell helper tests
- provider CLI smoke checks
- foundation plan CLI smoke checks
- TUI slash-command smoke coverage for provider/adapter introspection

## Recommended next implementation sequence
1. continue routing remaining top-level placeholder orchestration surfaces to `foundation/pi` runtime packages,
2. deepen repo-map ranking toward richer Aider-style graph semantics and add edit strategies,
3. expand `foundation/adapters` from visibility and route-selection seams into richer TormentNexus/TormentNexus provider routing and richer MCP execution adapters,
4. expand snapshot/result-shape coverage plus HTTP/CLI smoke coverage,
5. layer in delegation, verification, detached/background runs, and JSON/RPC transport.
