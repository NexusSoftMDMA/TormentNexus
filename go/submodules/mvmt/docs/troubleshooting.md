# Troubleshooting

## `mvmt serve` says port is in use

Another process is using the port.

```bash
mvmt serve --port 4142
```

## `No mounts loaded. Nothing to serve.`

mvmt needs at least one enabled mount.

In an interactive terminal, `mvmt serve` offers to add a mount and then
continues startup. In non-interactive runs, add one first:

```bash
mvmt mounts list
mvmt mounts add workspace ~/code/mvmt --mount-path /workspace --read-only
mvmt serve -i
```

## Tools are missing

Check:

- at least one mount is enabled;
- the client token is current;
- the API token policy grants the needed virtual path and action;
- the mount itself allows writes when using `write` or `remove`;
- protected paths are not being targeted.

Useful commands:

```bash
mvmt doctor
mvmt mounts list
mvmt token
```

## Search returns no results

The text index may not be built yet.

```bash
mvmt reindex
```

Also check that the file type is text-like and smaller than the current text size limit.

## Token rejected by client

Scoped API tokens are stored as scrypt verifiers, so mvmt cannot print their
plaintext value again. If you lost a token, create a replacement with
`mvmt token add` and update the client.

```bash
mvmt token
```

If the token works in curl but not an MCP client, restart the client after
updating its stored token.

## Codex says login is required

For local bearer-token setup, do not use `codex mcp login mvmt`.

Codex usually says this when its stored token is missing or stale.

```bash
mvmt token add codex --scope notes:read --expires 7d
codex mcp add mvmt --url http://127.0.0.1:4141/mcp
```

## OAuth client gets `redirect_uri is not registered for this client`

mvmt only redirects OAuth authorization codes to registered callback URLs.

The remote client must either:

- support RFC 7591 Dynamic Client Registration and call `/register`; or
- use a pre-registered exact `redirect_uri`.

## OAuth provider returned `invalid_target`

The OAuth flow supplied a resource that does not match this mvmt instance's MCP resource.

The canonical resource is:

```text
https://your-public-mvmt-host/mcp
```

Recent mvmt versions default a missing authorize `resource` to the canonical `/mcp` resource. Explicitly wrong resources are rejected.

## Remote client cannot connect and no tool calls appear

OAuth and MCP handshake failures can happen before a tool call exists, so they do not appear in the tool-call audit log.

Use interactive request logs:

```bash
mvmt serve -i
> logs on
```

You should see sanitized events such as:

```text
oauth.discovery GET /.well-known/oauth-authorization-server 200
oauth.register POST /register 201
oauth.authorize GET /authorize 200
oauth.token POST /token 400 invalid_grant
mcp.auth GET /mcp 401 missing_or_invalid_bearer
```

These live events do not include bearer tokens, session tokens, authorization codes, code challenges, redirect URIs, or full query strings.
