# Go Foundation Assimilation Testing Strategy

## Testing goals
The new foundation should be tested in layers.

## Layer 1: Contract tests
For every exact-name tool contract:
- validate parameter schema shape,
- validate name stability,
- validate result envelope invariants,
- compare native behavior against bridged/reference expectations.

Initial contract targets:
- `read`
- `write`
- `edit`
- `bash`

## Layer 2: Agent-loop tests
For the Pi-derived harness contract:
- event order (`agent_start` ... `agent_end`)
- tool execution sequencing
- parallel vs sequential tool modes
- steering/follow-up delivery modes
- session persistence and restore
- compaction hooks

## Layer 3: Integration tests
- TormentNexus provider routing integration
- TormentNexus MCP inventory integration
- memory retrieval integration
- imported session continuity
- TUI/JSON/RPC mode smoke tests

## Layer 4: Parity verification tests
For each assimilated upstream family:
- feature checklist coverage
- tool contract parity snapshots
- prompt/command UX parity snapshots where appropriate
- migration tests ensuring native replacement preserves behavior

## Current baseline run
Observed before any targeted fixes:

```text
go test ./...
- fixture compile error in aider/tests/fixtures/languages/go/test.go
- API mismatch in mcp/mcphost_test.go
- orchestrator panic from duplicate sqlite registration
```

## Tests added in this phase
- `foundation/compat/catalog_test.go`
- `foundation/assimilation/inventory_test.go`
- `foundation/pi/foundation_test.go`
- `foundation/pi/runtime_test.go`
- `foundation/pi/session_test.go`
- `foundation/pi/tool_parity_test.go`
- `foundation/pi/tool_snapshot_test.go`
- `foundation/repomap/repomap_test.go`
- `foundation/orchestration/planner_test.go`
- `foundation/orchestration/daemon_plan_test.go`
- `foundation/orchestration/webhook_plan_test.go`
- `foundation/adapters/tormentnexus_test.go`
- `foundation/adapters/providers_test.go`
- `foundation/adapters/provider_routing_test.go`
- `foundation/adapters/provider_execution_test.go`
- `foundation/adapters/mcp_test.go`
- `tools/registry_test.go`
- `agent/agent_test.go`
- `agent/orchestrator_test.go`
- `agents/director_test.go`
- `mcp/client_test.go`
- `mcp/manager_test.go`
- `cmd/foundation_http_test.go`
- `orchestrator/orchestration_bridge_test.go`
- `tui/slash_test.go` (plan/repomap/provider/adapter introspection)

## Tests that should be added next
1. `cmd/foundation` and HTTP route smoke tests
2. more tool contract schema/result snapshot tests
3. more truncation and image-path edge-case tests for `read` and `bash`
4. JSON/RPC transport tests
5. compatibility tests against richer TormentNexus-backed provider/MCP adapters
6. richer top-level `agent` integration tests around exact-schema tool registration and tool-call loops
7. end-to-end MCP execution tests once richer execution adapters exist
8. response-shape assertions for foundation-backed MCP HTTP endpoints
9. response-shape assertions for foundation-backed provider HTTP endpoints
10. response-shape assertions for provider execution-preparation endpoints
11. response-shape assertions for foundation-backed orchestration plan endpoints

## Exit criteria for the next milestone
- foundation packages compile cleanly,
- contract registry tests pass,
- CLI inspection and execution commands are covered,
- default tool contracts are backed by real implementations,
- session persistence is stable under create/list/load/fork flows,
- repo-map output is stable enough for deterministic tests,
- graph-ranking groundwork is stable enough for deterministic ranking tests,
- maturity labels are truthful and enforced by tests.
