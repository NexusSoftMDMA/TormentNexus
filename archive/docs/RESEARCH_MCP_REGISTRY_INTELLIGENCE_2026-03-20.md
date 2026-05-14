# MCP Registry Intelligence Research Notes (2026-03-20)

## Why this note exists
This captures implementation patterns from external MCP ecosystem projects and maps them to borg P0 catalog/validation work.

## External patterns worth adopting

### 1) Upstream server normalization contract
Observed in MCPProxy docs:
- explicit server-type split (`stdio`, `http`, OAuth-backed HTTP)
- consistent per-server shape: `name`, `command/url`, `args`, `env`, `enabled`, optional OAuth, optional isolation

Adopt in borg:
- canonical normalized shape per published server variant
- transport and auth modeled independently from install method

### 2) Isolation-first execution for untrusted servers
Observed in MCPProxy docker isolation:
- global isolation defaults + per-server overrides
- runtime->image mapping
- explicit lifecycle handling and cleanup labels
- security hardening knobs (`network_mode`, capability dropping)

Adopt in borg:
- validation harness should support isolation profiles
- classify results by profile (`native`, `isolated`, `network-none`)

### 3) Quarantine + approval workflow
Observed in MCPProxy quarantine model:
- newly added servers may start in quarantined state

Adopt in borg:
- published catalog state machine includes non-promoted state before installability claims
- optional manual review gate for risky auth/execution patterns

### 4) Sensitive-data detection + redacted audit logs
Observed in MCPProxy sensitive-data docs:
- category/severity model
- argument/response scanning
- redaction in activity logs
- compliance export flows

Adopt in borg:
- validation run logs should store detections as redacted findings, not raw secrets
- add severity-based badge/warnings in catalog UI

### 5) Namespace + endpoint composition model
Observed in MetaMCP and MCPHub:
- group servers into namespaces/groups
- expose unified endpoint and scoped endpoints
- selective tool curation per namespace

Adopt in borg:
- config recipes should support grouping/composition metadata
- published catalog should track if a server is best used standalone vs grouped

### 6) Agent-side tool narrowing controls
Observed in mcp-use:
- disallowed tools, server manager, bounded steps, memory toggles

Adopt in borg:
- operator flow should include optional safe defaults for tool exposure when first enabling a server

## Proposed borg data model additions (P0)

1. `published_mcp_servers`
- canonical_id, display_name, normalized_transport, install_method, auth_model
- status (`discovered|normalized|probeable|validated|certified|broken`)
- confidence, last_seen_at, last_verified_at

2. `published_mcp_server_sources`
- server_id, source_name, source_url, raw_payload_ref, first_seen_at, last_seen_at

3. `published_mcp_config_recipes`
- server_id, recipe_version, template, required_secrets, required_env, confidence, explanation

4. `published_mcp_validation_runs`
- server_id, run_mode, started_at, finished_at, outcome, failure_class, findings_summary, trace_ref

## Immediate implementation order
1. Create schema + repository layer
2. Add one ingestion adapter with provenance capture
3. Add one validation runner path storing results
4. Expose read-only catalog API endpoint
5. Wire minimal dashboard table from DB

## Open questions
- Canonical ID strategy across registries when names diverge heavily
- Minimum safe default for low-confidence recipes
- Retention policy for validation traces and redacted findings
