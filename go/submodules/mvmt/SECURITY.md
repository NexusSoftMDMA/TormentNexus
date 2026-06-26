# Security Policy

mvmt exposes personal local data. Security bugs should be handled privately and conservatively.

## Supported Versions

mvmt is pre-1.0. Security fixes target the latest code on `main` until a stable release process exists.

| Version | Supported |
| --- | --- |
| main | yes |
| < 0.1.0 | no |

## Reporting A Vulnerability

Please do not open a public issue with exploit details.

Use GitHub private vulnerability reporting for this repository if it is enabled. If private reporting is not available, open a minimal public issue asking for a private security contact and do not include sensitive details.

Include:

- Affected version or commit.
- Operating system.
- Whether mvmt was running local-only or through a tunnel.
- Minimal reproduction steps.
- Impact: what data or capability became accessible.
- Whether the issue requires local user access, local network access, or public internet access.

## Security Model Summary

Current protections:

- HTTP mode binds to `127.0.0.1` only.
- HTTP `/mcp` and `/health` requests require a bearer token.
- Origin checks reject non-localhost browser origins unless allowlisted.
- Local folder access is scoped by explicit mounts.
- Scoped API tokens can restrict each client by virtual path and action.
- Write/remove access is opt-in at both the client policy and mount level.
- Protected paths block write/remove even when write access is otherwise allowed.
- Tunnel mode is closed to legacy all-mount access unless explicitly overridden.
- Auth, MCP, and health routes have rate limits.
- Stdio child processes receive a scrubbed environment.
- Tool calls are written to a local audit log.

Known limits:

- Tunnel mode exposes the configured HTTP surface to the internet. Use narrow
  mounts and scoped API tokens.
- Localhost traffic is plaintext.
- `protect` blocks write/remove, not reads. Use `exclude` to hide readable
  secrets from listing, indexing, and reads.
- The audit log can include truncated argument values.
- OAuth/tunnel auth is pre-1.0 and should be treated conservatively.

See `docs/threat-model.md` for the current threat model and
`docs/security-memo.md` for implementation notes.
