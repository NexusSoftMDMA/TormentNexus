# Threat Model

mvmt exposes selected local folders to MCP clients. The main boundary is the
mount table plus scoped API-token permissions, not the client or model.

## Assets

- Local files under configured mounts.
- Scoped API tokens and OAuth access tokens.
- The internal session token at `~/.mvmt/.session-token`.
- The saved config at `~/.mvmt/config.yaml`.
- The local audit log.

## Trust Boundaries

- **Local OS user:** mvmt assumes the user account running mvmt can read the
  configured folders. It does not protect against a compromised local account.
- **Local HTTP listener:** HTTP mode binds to `127.0.0.1`; non-local machines
  should not reach it directly.
- **Public tunnel:** tunnel mode moves the HTTP surface onto the internet.
  Treat the tunnel URL as public.
- **MCP client:** clients can be buggy or compromised. mvmt still enforces its
  own path/action policy on every request.
- **Mounted folders:** each mount root is resolved before access. Traversal and
  symlink escapes are rejected.

## What mvmt Protects Against

- Accidental full-computer exposure: no mount means no data, and clients see
  virtual paths such as `/notes`, not host paths such as `/Users/you`.
- Unauthorized data-plane access in tunnel mode: public `/mcp` access requires
  a scoped API token or OAuth access token mapped to a scoped client.
- Unknown OAuth clients: once `clients[]` exists, unknown OAuth `client_id`
  values are quarantined and receive no tools.
- Path traversal and symlink escapes outside a mount root.
- Writes to read-only mounts.
- Writes/removes of protected paths such as `.env`, `.claude/**`, and configured
  credential patterns.
- Drive-by browser requests from untrusted origins.
- Basic brute force and request floods through rate limits on auth, MCP, and
  health routes.

## What mvmt Does Not Protect Against

- A malicious or compromised local OS user.
- A client that has been intentionally granted broad mount access.
- A model or client persuading another authorized client to disclose data. Use
  narrow tokens per client and audit logs to reduce blast radius.
- Secrets that are readable inside a mount. `protect` blocks write/remove; use
  `exclude` to hide a path from listing, indexing, and reads.
- Plaintext localhost traffic. Use tunnel HTTPS only when remote access is
  needed.
- A compromised tunnel provider, DNS account, or machine running the tunnel.
- Secrets that are intentionally left inside readable scope. mvmt does not
  inspect or rewrite file contents; use mount scope, `exclude`, `protect`, and
  narrow client permissions to keep sensitive data out of reach.

## Token Model

Normal clients should use scoped API tokens:

```bash
mvmt token add codex --scope notes:read --expires 7d
mvmt token rotate codex --expires 30d --yes
mvmt token remove codex
```

Plaintext API tokens are printed once. mvmt stores only a scrypt verifier in
config. Token TTLs are enforced during auth, and running servers reload client
policy on the next auth request.

The internal session token exists for legacy local compatibility. Once
`clients[]` exists, it no longer grants data-plane access to `/mcp`.

## Tunnel Mode

Before enabling a public tunnel:

1. Use narrow mounts.
2. Prefer read-only mounts.
3. Create scoped API tokens.
4. Add `exclude` rules for secrets that must not be read.
5. Add `protect` rules for files that must not be written or removed.
6. Watch the audit log during first use.

Tunnel mode without scoped clients is closed by default. The
`MVMT_ALLOW_LEGACY_TUNNEL=1` escape hatch exists only for temporary debugging.

## Audit And Detection

Tool calls are appended to the local audit log. Audit entries can contain
truncated argument previews. Do not treat the audit log as secret-free.

Audit logging helps answer:

- which client called a tool;
- which path or query was requested;
- whether the call failed or was denied.

Audit log rotation is manual today.

## Operational Guidance

- Grant one scoped token per client.
- Keep token TTLs short for tunnel use.
- Do not mount home directories wholesale.
- Use `exclude` for readable secrets and `protect` for write/remove guardrails.
- Run `mvmt doctor`, `mvmt mounts list`, and `mvmt token` when debugging access.
- Revoke or rotate a token when a client is lost, compromised, or no longer
  needed.
- Treat `--client` as an intended-client hint, not a security boundary. It can
  catch accidental use by the wrong client, but the plaintext token remains the
  credential to protect.
- Use `mvmt token edit <id> --no-permissions` to keep a token record while
  removing all data access.
- Use `mvmt token rotate <id>` to invalidate one API token. Use `mvmt token
  session-rotate` when you need to rotate the internal session token and OAuth
  signing key across clients.
