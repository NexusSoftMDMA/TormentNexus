# Setup Guide

This guide covers install, mount setup, server startup, and client connection.

## Requirements

- Node.js 20 or newer.
- npm.
- Optional: `cloudflared` or another tunnel command for remote web clients.

## Install

Install the CLI from source:

```bash
git clone https://github.com/philipnee/mvmt.git
cd mvmt
npm install
npm run build
npm link
mvmt --version
```

The `mvmt` package name is already taken on npm, so the first public beta uses
the source checkout path above.

## Desktop App From Source

`mvmt-desktop` is an Electron shell over the mvmt engine. For local testing,
keep both repos beside each other and build the engine first:

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

To make a standalone pre-alpha desktop bundle, vendor the built engine into the
desktop app before packaging:

```bash
cd ~/code/mvmt-desktop
npm run vendor
npm run dist:mac
```

## First 10 Minutes

Use this path for a local smoke test:

```bash
mvmt mounts add documents ~/Documents --mount-path /documents --read-only
mvmt token add codex --scope documents:read --expires 7d
mvmt reindex
mvmt serve -i
codex mcp add mvmt --url http://127.0.0.1:4141/mcp
```

`-i` starts mvmt in interactive mode.

When the client asks for authentication, paste the `mvmt_t_...` API token printed
by `mvmt token add`.

Then ask the client to:

```text
Search /documents for a project note and read the most relevant result.
```

The happy path is mount -> token -> index -> serve -> connect -> test.

## Configure Mounts

Run guided setup:

```bash
mvmt config setup
```

Or add mounts directly:

```bash
mvmt mounts add notes ~/Documents/Obsidian \
  --mount-path /notes \
  --read-only \
  --description "Personal notes and project journals" \
  --guidance "Search first. Read specific files before answering."
```

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

The config is written to:

```text
~/.mvmt/config.yaml
```

At least one enabled mount is required. If there are no mounts, mvmt has no data to serve.

## Start mvmt

Interactive mode is easiest while testing:

```bash
mvmt serve -i
```

HTTP mode listens on:

```text
http://127.0.0.1:4141/mcp
```

For a one-off read-only folder without changing saved config:

```bash
mvmt serve --path ~/Documents -i
```

When adding a mount:

- **Folder on this computer** is the real local folder. It must already exist.
- **Virtual path clients will use** is the mvmt path, such as `/notes` or
  `/workspace`.

## API Tokens

List scoped API tokens:

```bash
mvmt token
```

Create a token:

```bash
mvmt token add codex --scope notes:read --expires 7d
```

The plaintext token is printed once. mvmt stores only a scrypt verifier.

The internal legacy session token is still stored at
`~/.mvmt/.session-token` with file mode `600`. HTTP `mvmt serve` creates it
automatically during startup. Use `mvmt token session` only for legacy local
testing when no scoped API tokens are configured.

```bash
mvmt token rotate codex --expires 30d --yes
```

Running mvmt processes load API-token changes on the next auth request.

## Rebuild the Index

Search uses the text index. mvmt rebuilds it in the background on startup.

Force a rebuild:

```bash
mvmt reindex
```

Use this after adding a mount or changing files outside mvmt.

## Connect Codex CLI

Create a scoped token first:

```bash
mvmt token add codex --scope notes:read --expires 7d
```

Use the printed `mvmt_t_...` token when Codex asks for authentication.

```bash
codex mcp add mvmt --url http://127.0.0.1:4141/mcp
```

When prompted, paste the `mvmt_t_...` token printed by `mvmt token add`.

## Connect Other Clients

Most local HTTP clients need:

```text
URL: http://127.0.0.1:4141/mcp
Authorization: Bearer <token>
```

Claude Desktop is different because it can launch mvmt over stdio:

```json
{
  "mcpServers": {
    "mvmt": {
      "command": "mvmt",
      "args": ["serve", "--stdio"]
    }
  }
}
```

See [Client Setup](client-setup.md) for more examples.

## Verify

Check server health:

```bash
TOKEN="<paste mvmt_t_... token from mvmt token add>"
curl -i http://127.0.0.1:4141/health \
  -H "Authorization: Bearer $TOKEN"
```

Expected:

```json
{"status":"ok","tools":5}
```

In an MCP client, the normal mount tools are:

```text
search
list
read
write
remove
```

The exact list depends on client permissions.

## Common Problems

### `No mounts loaded. Nothing to serve.`

In an interactive terminal, `mvmt serve` will offer to add a mount before
starting. In scripts or CI, add at least one mount explicitly:

```bash
mvmt mounts add workspace ~/code/mvmt --mount-path /workspace --read-only
mvmt serve -i
```

### `Tools: (none)`

Usually one of these:

- no enabled mounts exist;
- the client has a stale token;
- the client is mapped to a policy with no matching permissions;
- the client cached a failed tool-list response.

Run:

```bash
mvmt doctor
mvmt mounts list
mvmt token
```

Then restart the MCP client.

### Port `4141` is already in use

Start on another port:

```bash
mvmt serve -i --port 4142
```

### Token works in curl but not Codex

Codex only sees environment variables from the shell that launched it.

```bash
export MVMT_TOKEN="<paste mvmt_t_... token here>"
codex
```
