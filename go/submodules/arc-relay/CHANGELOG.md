# Changelog

All notable changes to Arc Relay (formerly MCP Wrangler) are documented here.

## [Unreleased]

### Added
- **Stateful archive handoff** - "Set up the Comma Compliance Archive" flow now mints a server-side nonce before opening the compliance popup and validates it on the return trip
  - Without the nonce, any crafted `#mw-archive?...` fragment on an authenticated page could silently reconfigure archive credentials
  - New endpoints: `POST /api/archive/handoff/begin`, `POST /api/archive/handoff/complete`
  - In-memory nonce store with 10-minute TTL, bound to the initiating admin session
  - Fragment values are never applied directly from the browser; the client posts them to `/complete` where server-side validation is authoritative
  - See `docs/archive-handoff.md` for the protocol specification
- **Envelope schema v2** - NaCl Box archive envelopes now include `version` and `kid` (key fingerprint) fields
  - `version: "nacl-box-v1"` lets receivers dispatch on the version, not the presence of a ciphertext field
  - `kid = base64(blake2b-256(recipient_pub)[:8])` lets receivers route decryption through multiple keys during rotation
  - Shared `sealArchivePayload` helper - real traffic and synchronous test deliveries go through the same sealing code
  - Schema documented in `docs/archive-envelope.md`
- **Envelope encryption UI** - archive config section shows an "Envelope encrypted" indicator with fingerprint when a recipient key is configured, and a "Remove encryption" button for explicit plaintext downgrade
- `ValidateArchiveConfig` is extracted as a public function and called at save time
  - Rejects non-https URLs unless the host is localhost/loopback
  - Rejects unknown `auth_type` or `include` values
  - Rejects malformed `nacl_recipient_key` values before they reach the enqueue path
- **Tool Context Optimizer** - LLM-powered tool definition compression to reduce context token usage
  - Per-server opt-in: audit tool sizes, run LLM optimization, toggle serving optimized tools
  - Deterministic JSON Schema pruning plus LLM-based description compression via Anthropic API
  - Hash-based invalidation detects upstream tool changes, marks optimizations stale
  - Optimizer middleware intercepts tools/list responses when enabled
  - Before/after tool details table with per-tool savings and red/green coloring
  - Concurrent run guard, adaptive batch sizing for large schemas
  - Config: `ARC_RELAY_LLM_API_KEY`, `ARC_RELAY_LLM_MODEL` env vars
  - Migration 014: `tool_optimizations` table, `servers.optimize_enabled` column
- `scripts/lint.sh` - local lint script mirroring CI checks

## [1.0.0] - 2026-04-01

### Changed
- **Renamed from MCP Wrangler to Arc Relay** - new module path `github.com/comma-compliance/arc-relay`
- Binary names: `arc-relay` (server), `arc-sync` (CLI, formerly `mcp-sync`)
- Environment variables: `ARC_RELAY_*` (server), `ARC_SYNC_*` (CLI)
- Config directory: `~/.config/arc-sync/` (CLI)
- Docker image: `ghcr.io/comma-compliance/arc-relay`
- License changed from AGPL-3.0 to MIT

### Added
- **NaCl Box encryption** for archive webhook payloads (X25519 + XSalsa20-Poly1305)
- OSS documentation: AGENTS.md, CODE_OF_CONDUCT.md, SECURITY.md, GitHub issue/PR templates

## [0.3.0] - 2026-03-08

### Added
- **Proxy Middleware Pipeline** - bidirectional request/response processing for MCP traffic
- **Sanitizer middleware** — PII/secret redaction with configurable regex patterns (redact or block)
- **Content Sizer middleware** — response size limits with truncate/block/warn actions (default 500KB)
- **Alerter middleware** — pattern monitoring with log and webhook alert actions
- Middleware toggle UI on server detail page (per-server enable/disable)
- Middleware event log with event type badges (redacted, blocked, truncated, alert)
- `middleware_configs` and `middleware_events` database tables (migration 004)
- 10 unit tests covering all middleware and pipeline behavior

## [0.2.3] - 2026-03-08

### Fixed
- Docker API compatibility: probe daemon version via `/_ping` and pin client API version to match, bypassing the SDK's minimum version check (fixes Docker on Unraid 6.x / Docker 24.x with API 1.43)

## [0.2.2] - 2026-03-08

### Added
- OAuth auto-discovery + dynamic client registration for manual server entry (not just catalog)
- Client-side OAuth discovery triggers when switching auth type dropdown to "oauth"

### Changed
- Improved error message when OAuth auto-discovery fails

## [0.2.1] - 2026-03-03

### Fixed
- Docker startup no longer requires config.toml (env vars are sufficient)

## [0.2.0] - 2026-03-03

### Fixed
- Server edit now preserves status, timestamps, and OAuth tokens on update

## [0.1.1] - 2026-03-01

### Added
- Foundation: proxy, Docker lifecycle, web UI, API keys, health monitor

## [0.1.0] - 2026-02-28

### Added
- Initial open-source release with security hardening
