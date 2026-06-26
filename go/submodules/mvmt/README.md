# mvmt

**Multi-Volume Mount Transport.**

mvmt exposes selected local folders through a small, permissioned tool API.

Mount folders into stable paths like `/notes`, `/workspace`, or `/research`,
then allow clients to search, list, read, write, or remove files within those
mounts. Clients never get full-computer access. They only see the mounts and
actions you allow.

```txt
client
  |
  v
mvmt
  |
  |-- /notes      -> ~/Documents/Obsidian
  |-- /workspace  -> ~/code/mvmt
  |-- /research   -> ~/papers
```

mvmt is **not sync**.
mvmt is **not cloud storage**.
mvmt is **not a general-purpose filesystem server**.

It is a local-first access layer for exposing specific parts of your machine
through explicit mounts, narrow tools, and per-client permissions.

![mvmt running in interactive mode](docs/assets/mvmt-start-interactive.png)

## Quick start

Install mvmt from source:

```bash
git clone https://github.com/philipnee/mvmt.git
cd mvmt
npm install
npm run build
npm link
mvmt serve -i
```

The `mvmt` package name is already taken on npm, so the first public beta uses
the source checkout path above.

`-i` starts mvmt in interactive mode.

## Desktop from source

`mvmt-desktop` is an Electron shell over the local mvmt engine. Until a signed
desktop build is published, keep the repos side by side and build this engine
first:

```bash
cd ~/code
git clone https://github.com/philipnee/mvmt.git
git clone https://github.com/philipnee/mvmt-desktop.git

cd mvmt
npm install
npm run build

cd ../mvmt-desktop
npm install
npm run dev
```

For a standalone pre-alpha app bundle, vendor the built engine into the desktop
repo before packaging:

```bash
cd ~/code/mvmt-desktop
npm run vendor
npm run dist:mac
```

On first run, mvmt creates:

```txt
~/.mvmt/config.yaml
```

It then walks through:

- adding local folder mounts;
- starting the MCP server.

For a one-off read-only folder without changing saved config:

```bash
mvmt serve --path ~/Documents -i
```

At least one enabled mount is required. If no mounts are configured, mvmt has no
data to serve.

## Why mvmt exists

Many tools need controlled access to local files.

Giving a tool broad filesystem access is risky. Uploading everything into a
cloud workspace is often unnecessary. Ad hoc local servers usually lack a clear
permission model.

mvmt takes a narrower approach:

- Mount only the folders you want clients to see.
- Expose those folders through stable virtual paths.
- Give each client its own path/action permissions.
- Keep data on your machine.
- Audit what clients search, read, write, and remove.

## Core ideas

### One local namespace

Clients operate on virtual paths:

```txt
/notes
/workspace
/research
```

They do not need to know where those folders live on disk.

### Explicit mounts

Each mount maps a virtual path to a real local folder.

```yaml
mounts:
  - name: notes
    path: /notes
    root: /Users/you/Documents/Obsidian
    writeAccess: false
```

### Per-client permissions

Different clients can see different parts of the namespace.

```yaml
clients:
  - id: codex
    permissions:
      - path: /workspace/**
        actions: [search, read, write]

  - id: readonly-client
    permissions:
      - path: /notes/**
        actions: [search, read]
```

### Read-only by default

A write requires both:

1. the client has `write` permission for the path; and
2. the mount has `writeAccess: true`.

Protected paths such as `.env`, `.claude/**`, or other configured patterns
cannot be written or removed. mvmt also has a global secret-path deny list
for paths such as `.mvmt/**`, `.ssh/**`, `.aws/**`, `.kube/**`, and common
cloud/dev credential files. Those paths are blocked even if an older config
does not list them.

## Example setup

Add a read-only notes mount:

```bash
mvmt mounts add notes ~/Documents/Obsidian \
  --mount-path /notes \
  --read-only \
  --description "Personal notes and project journals" \
  --guidance "Search first. Read specific files before answering."
```

Add a writable project mount:

```bash
mvmt mounts add workspace ~/code/mvmt \
  --mount-path /workspace \
  --write \
  --protect ".env" \
  --protect ".env.*" \
  --protect ".claude/**" \
  --description "mvmt source code and design docs" \
  --guidance "Read README.md and docs before changing code."
```

Then rebuild the index and serve:

```bash
mvmt reindex
mvmt serve -i
```

## Status

| Area | Status |
| --- | --- |
| Local folder mounts | supported |
| Text index | supported as a JSON prototype index |
| MCP tools | `search`, `list`, `read`, `write`, `remove` |
| Mount management CLI | supported |
| API-token path permissions | supported via `mvmt token add` |
| Local Streamable HTTP | supported on `127.0.0.1` |
| Stdio mode | supported |
| OAuth/PKCE for web clients | supported, including Dynamic Client Registration |
| Tunnel mode | supported for personal remote access |
| Legacy proxy connector config | accepted by the schema for compatibility, ignored by the mount-only CLI runtime |
| Admin UI | not shipped |
| API-token issuance | supported for bearer-token clients |
| Remote mvmt mounts | not shipped |
| Binary/PDF/image indexing | not shipped |

## Client compatibility

mvmt can be used by MCP clients, local HTTP clients, and remote web clients
through tunnel mode.

| Client | Transport | Status | Auth method | Notes |
| --- | --- | --- | --- | --- |
| Claude Desktop | stdio | supported | process launch | Runs its own mvmt process from the client config |
| Claude Code | Streamable HTTP | supported | bearer token | Use a scoped API token from `mvmt token add` |
| Codex CLI | Streamable HTTP | supported | bearer token | Use a scoped API token from `mvmt token add` |
| Cursor | Streamable HTTP | expected | bearer token | Behavior depends on Cursor's MCP implementation |
| VS Code / Copilot | Streamable HTTP | expected | bearer token | Behavior depends on the MCP extension |
| claude.ai / ChatGPT web | public HTTPS tunnel | supported remote mode | OAuth/PKCE | Requires a reachable tunnel URL |
| Raw HTTP / curl | Streamable HTTP | debug only | bearer token | Must follow MCP session initialization rules |

Most local HTTP clients need:

```txt
URL: http://127.0.0.1:4141/mcp
Authorization: Bearer <token>
```

In local legacy mode, `<token>` can be the internal session token from
`mvmt token session`. For normal scoped access, create an API token with
`mvmt token add` and use that token instead.

Remote web clients use the tunnel URL, usually ending in `/mcp`, and authorize
through OAuth 2.1 + PKCE.

mvmt supports RFC 7591 Dynamic Client Registration. Issued OAuth access tokens
are audience-bound to the current mvmt resource, so a token minted for one mvmt
instance cannot be replayed against another. OAuth refresh tokens are rotated on
each use; replaying an old refresh token revokes that refresh-token family.

Claude Desktop is different: it launches mvmt over stdio, so there is no HTTP
listener and no bearer token header.

## Tool surface

When `mounts[]` is configured, mvmt exposes five MCP tools:

| Tool | Purpose | Required permission |
| --- | --- | --- |
| `search` | Search indexed text chunks across permitted mounts | `search` |
| `list` | List visible mount roots or directories | `read` |
| `read` | Read one text file by virtual path | `read` |
| `write` | Create or overwrite one text file | `write` |
| `remove` | Delete one text file | `write` |

Important behavior:

- `list("/")` returns visible mount roots with description, guidance, and
  write-access status.
- `search` uses simple keyword scoring over indexed text chunks.
- `read` and `write` are text-only.
- `write` supports `expected_hash` so clients can avoid overwriting stale reads.
- `remove` deletes files permanently. It does not remove directories.
- Non-text files are hidden from `list`, skipped by indexing, and rejected by
  `read`/`write`.

Supported text-like files include Markdown, plain text, JSON, YAML, TOML, CSV,
logs, shell scripts, HTML/CSS/XML, and common source-code extensions.

Files over 2 MiB are skipped by the index and rejected by direct text reads.

## Configuration

The saved config lives at:

```txt
~/.mvmt/config.yaml
```

Inspect it:

```bash
mvmt config
```

Validate it:

```bash
mvmt doctor
```

Manage mounts:

```bash
mvmt mounts list
mvmt mounts add
mvmt mounts edit
mvmt mounts remove
```

Minimal config:

```yaml
version: 1

server:
  port: 4141
  allowedOrigins: []
  access: local

mounts:
  - name: notes
    type: local_folder
    path: /notes
    root: /Users/you/Documents/Obsidian
    description: Personal notes and project journals.
    guidance: Search first. Read specific files before answering.
    exclude:
      - .git/**
      - node_modules/**
      - .claude/**
      - .mvmt/**
      - .ssh/**
      - .aws/**
    protect:
      - .env
      - .env.*
      - .claude/**
      - .mvmt/**
      - .ssh/**
      - .aws/**
    writeAccess: false
    enabled: true

  - name: workspace
    type: local_folder
    path: /workspace
    root: /Users/you/code/mvmt
    description: mvmt project source and design docs.
    guidance: Read README.md and docs before changing code.
    exclude:
      - .git/**
      - node_modules/**
      - dist/**
    protect:
      - .env
      - .env.*
      - .claude/**
      - .mvmt/**
      - .ssh/**
      - .aws/**
    writeAccess: true
    enabled: true
```

Mount fields:

| Field | Meaning |
| --- | --- |
| `name` | Stable lowercase mount id, such as `notes` |
| `path` | Virtual path visible to clients, such as `/notes` |
| `root` | Local folder on disk |
| `description` | Short summary returned from `list("/")` |
| `guidance` | Optional client-facing guidance returned from `list("/")` |
| `exclude` | Paths hidden from listing, reads, writes, removal, and indexing |
| `protect` | Paths that cannot be written or removed |
| `writeAccess` | Mount-level write gate; defaults to `false` |
| `enabled` | Whether the mount is active |

An Obsidian vault is just a local folder mount. There is no special Obsidian
runtime connector in the current mount-only shape.

## API tokens and policy

For local testing, mvmt keeps legacy behavior when no API tokens are configured:
the internal session bearer token can access all configured mounts. The session
token is stored at `~/.mvmt/.session-token` with file mode `600`. HTTP
`mvmt serve` creates it if it is missing.

For repeatable client access, create scoped API tokens:

```bash
mvmt token add codex --scope notes:read --expires 7d
mvmt token add laptop-claude --scope all:read --client claude-desktop --expires never
```

`mvmt token add` prints the plaintext token once. mvmt stores only a scrypt
verifier in config. Tokens do not expire by default; use values like
`--expires 30m`, `--expires 7d`, or `--expires never` when creating or editing a token.
If a client asks for an authorization token, paste the printed `mvmt_t_...`
value. If a web client sends you through OAuth, paste the same token into the
mvmt approval page.

`--client` is an intended-client hint checked against the request client label
or User-Agent. It helps catch misconfiguration, but it is not a cryptographic
binding if the plaintext token leaks.

Once API tokens are present, `/mcp` becomes strict:

- bearer tokens must match a configured client `tokenHash`;
- OAuth access tokens inherit the API-token identity selected on the approval page;
- the session token no longer grants data-plane access;
- unknown OAuth clients are quarantined with zero permissions.

Policy changes take effect on the next auth request. Scope edits, including
adding access, removing paths, or `--no-permissions`, apply to existing API
tokens and OAuth grants without reauthorization. Editing `clientBinding` bumps
that token's credential version, so OAuth access and refresh tokens selected
through the old binding must reauthorize.
`mvmt token rotate <id>` replaces only that API token's secret; `mvmt token
session-rotate` rotates the internal session token and OAuth signing key,
revoking OAuth access tokens across clients.

Tunnel mode can start with no API tokens, but MCP data access rejects the
legacy session token. Add API tokens to grant access. For temporary debugging
only, set `MVMT_ALLOW_LEGACY_TUNNEL=1` to allow the legacy session token over a
tunnel.

API-token permissions are written against virtual paths, not local disk paths.
The underlying config field is currently named `clients[]`.

```yaml
clients:
  - id: codex
    name: Codex CLI
    description: Local Codex token
    expiresAt: "2026-05-06T12:00:00.000Z"
    auth:
      type: token
      # Verifier for the client API key.
      # Do not store the plaintext key.
      tokenHash: "scrypt:v1:..."
    rawToolsEnabled: false
    permissions:
      - path: /workspace/**
        actions: [search, read, write]
      - path: /notes/**
        actions: [search, read]

  - id: readonly-client
    name: Read-only client
    auth:
      type: token
      tokenHash: "scrypt:v1:..."
    rawToolsEnabled: false
    permissions:
      - path: /notes/**
        actions: [search, read]
```

Policy is additive. A call succeeds only when the resolved API token has the
required action for the target virtual path.

For writes, the mount itself must also have `writeAccess: true`, and the target
must not match `protect`.

## Commands

| Command | Description |
| --- | --- |
| `mvmt serve` | Configure mvmt if needed, then start the MCP server |
| `mvmt serve -i` | Start with an interactive control prompt |
| `mvmt serve --path <dir>` | Temporarily serve one read-only folder |
| `mvmt reindex` | Rebuild the text index |
| `mvmt mounts` | List configured mounts |
| `mvmt mounts --json` | List configured mounts as JSON |
| `mvmt mounts add [name] [root]` | Add a local folder mount |
| `mvmt mounts edit [name]` | Edit a mount |
| `mvmt mounts remove [name]` | Remove a mount |
| `mvmt mounts remove [name] --yes` | Remove a mount without an interactive confirmation |
| `mvmt config` | Show the saved config summary |
| `mvmt config setup` | Rerun guided setup |
| `mvmt doctor` | Validate config and startup prerequisites |
| `mvmt token` | List scoped API tokens |
| `mvmt token add [name]` | Create a scoped API token and print it once |
| `mvmt token edit [id]` | Edit a scoped API token |
| `mvmt token edit [id] --no-permissions` | Keep a token but remove all data access |
| `mvmt token rotate [id] --yes` | Replace a scoped API token secret |
| `mvmt token remove [id]` | Remove a scoped API token |
| `mvmt token session` | Hidden compatibility command for the internal session token |
| `mvmt tunnel` | Show tunnel status |
| `mvmt tunnel config` | Choose and save a tunnel command |
| `mvmt tunnel start` | Start the configured tunnel |
| `mvmt tunnel refresh` | Restart the tunnel and print the new URL |
| `mvmt tunnel stop` | Stop public tunnel exposure |
| `mvmt tunnel disable` | Switch config back to local-only access |
| `mvmt tunnel logs` | Show recent tunnel output |
| `mvmt tunnel logs stream` | Stream live tunnel output |
| `mvmt --version` | Print version and check for updates |

Interactive mode accepts the same command groups:

```txt
> mounts
> mounts add
> mounts edit
> mounts remove
> config
> config setup
> token
> token add
> token edit
> token rotate
> token remove
> tunnel
> tunnel refresh
```

## Index lifecycle

On `mvmt serve`, mvmt starts serving immediately and rebuilds the text index in
the background.

Search may return fewer results until the rebuild finishes.

Use:

```bash
mvmt reindex
```

after changing files outside mvmt or after editing mount configuration.

Writes and removes performed through mvmt rebuild the index automatically.

The current index is a JSON file beside the config file. SQLite and incremental
file watching are planned but not shipped.

## Security model

Every data operation is gated by path and action.

- HTTP mode binds to `127.0.0.1`, not `0.0.0.0`.
- HTTP `/mcp` and `/health` require authentication.
- The session token is stored at `~/.mvmt/.session-token` with mode `600`.
- Mount roots are resolved before access.
- Path traversal outside the mount root is rejected.
- `exclude` hides paths from listing, reading, writing, removal, and indexing.
- `protect` blocks write/remove for sensitive paths such as `.env` and
  `.claude/**`.
- Global secret paths such as `.mvmt/**`, `.ssh/**`, `.aws/**`, `.kube/**`,
  and common credential files are denied regardless of per-mount config.
- Write operations require both client `write` permission and mount
  `writeAccess`.
- Unknown OAuth clients are quarantined once API-token/OAuth policy exists.
- Browser-origin checks block drive-by browser requests from non-local origins.
- Tool calls are appended to `~/.mvmt/audit.log`.

Authentication controls who connects. Mounts and client policy control what they
can access.

## Remote access

mvmt is local-first.

Remote clients cannot reach `127.0.0.1` directly, so tunnel mode can publish a
public HTTPS URL.

Remote access checklist:

1. Mount only the folders the remote client needs.
2. Prefer read-only mounts.
3. Start the tunnel when ready. Without API tokens, the public endpoint is
   reachable but cannot read data.
4. Create scoped API tokens with `mvmt token add`.
5. Use `exclude` for secrets that must not be read.
6. Use `protect` for files that must not be written or removed.
7. Watch `~/.mvmt/audit.log` when testing a new remote client.
8. Use a stable tunnel URL for repeatable OAuth flows.

Quick tunnels are convenient but temporary.

## Current limits

- Local folders are the only shipped mount provider.
- mvmt currently runs as a single instance.
- Remote mvmt mounts are not shipped.
- The active CLI runtime is mount-only.
- Legacy proxy connector config still parses for compatibility, but proxy
  connectors are not loaded as runtime tools.
- Search is prototype keyword scoring, not semantic embedding search.
- PDFs, images, archives, and other binary files are skipped.
- There is no file watcher yet.
- API-token rotation is shipped. An admin UI is not shipped.
- mvmt does not sync, replicate, or resolve conflicts across machines.

## Coming next

Near-term:

- SQLite-backed text index.
- Incremental index updates.
- Better mount and API-token management commands.
- Admin UI for client keys, path permissions, audit, and mount visibility.

Later:

- Remote mvmt mounts.
- Multiple mount providers behind the same `search`/`list`/`read`/`write`/`remove`
  surface.
- `resolve(path) -> AccessPlan` for routing virtual paths to local or remote
  backends.

## Testing

`npm test` runs the unit and integration suite (no external dependencies).
`npm run verify` is the PR gate — build, full tests, whitespace, package
dry-run, and a runtime smoke test. See [CONTRIBUTING.md](CONTRIBUTING.md) for
the full matrix.

`npm run e2e` runs a Docker-based end-to-end suite that builds the package,
installs it fresh into a `node:20-alpine` container, and exercises real CLI
flows against the running server. Requires Docker. The first run pays the
image-build cost (~30s); subsequent runs reuse the cached image. The current
suite covers the token + auth surface — scope enforcement, path validity,
and bearer-token failure modes — and is excluded from `npm test` so the
default loop stays fast.

## Project docs

- [Setup guide](docs/setup.md)
- [Client setup](docs/client-setup.md)
- [Text index prototype](docs/text-index-prototype.md)
- [Remote access](docs/remote-access.md)
- [Audit log](docs/audit-log.md)
- [Threat model](docs/threat-model.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Architecture](docs/architecture.md)
- [Security policy](SECURITY.md)
- [Contributing](CONTRIBUTING.md)
- [Changelog](CHANGELOG.md)

## Contributing

Contributions are welcome while the project is early.

Keep changes focused and security-conscious. Read [CONTRIBUTING.md](CONTRIBUTING.md)
and report vulnerabilities through [SECURITY.md](SECURITY.md).

## License

MIT
