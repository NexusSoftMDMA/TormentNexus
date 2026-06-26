# Text Index Prototype

This branch prototypes the simplified mvmt product shape:

```text
permissioned mounts -> local text index -> five MCP tools
```

It intentionally supports text-like files only. Markdown and plain text are
the primary targets. Binary files such as PDFs and images are skipped.

## Config

Add local folder mounts to `~/.mvmt/config.yaml`:

```yaml
mounts:
  - name: workspace
    type: local_folder
    path: /workspace
    root: /Users/you/code/mvmt
    description: "Project source for mvmt development."
    guidance: |
      This is the mvmt repo.
      Check /docs for design notes and /tests for behavior coverage.
      Do not edit generated files under /dist.
    exclude:
      - .git/**
      - "**/.git/**"
      - node_modules/**
      - "**/node_modules/**"
      - .claude/**
      - "**/.claude/**"
    protect:
      - .claude/**
      - .env
      - .mvmt/**
      - .ssh/**
      - .aws/**
    writeAccess: true
    enabled: true
```

Or use the mount commands:

```bash
mvmt mounts add workspace /Users/you/code/mvmt --write \
  --description 'Project source for mvmt development.' \
  --guidance 'This is the mvmt repo. Check /docs for design notes and /tests for behavior coverage.' \
  --exclude '.git/**' \
  --exclude 'node_modules/**' \
  --exclude '.claude/**' \
  --protect '.env' \
  --protect '.env.*' \
  --protect '.claude/**'

mvmt mounts list
mvmt mounts edit workspace --read-only
mvmt mounts edit workspace --description 'Read-only notes archive' \
  --guidance 'Daily notes live under /daily. New AI-written notes belong in /inbox.'
mvmt mounts edit workspace --root /Users/you/code/other --write
mvmt mounts remove workspace
```

mvmt also enforces a global secret-path deny list for `.mvmt/**`, `.ssh/**`,
`.aws/**`, `.kube/**`, and common credential files even if a mount omits them.

In interactive mode (`mvmt serve -i`), use:

```text
mounts
mounts add
mounts edit
mounts remove
```

```yaml
clients:
  - id: codex
    name: Codex CLI
    auth:
      type: token
      tokenHash: "..."
    rawToolsEnabled: false
    permissions:
      - path: /workspace/**
        actions: [search, read, write]
```

## Tool Surface

When `mounts[]` is configured, mvmt exposes:

| Tool | Purpose | Permission |
| --- | --- | --- |
| `search` | Search indexed text chunks | `search` |
| `list` | List permitted mount roots or directories | `read` |
| `read` | Read one text file | `read` |
| `write` | Create or overwrite one text file | `write` |
| `remove` | Remove one text file | `write` |

`write` accepts `expected_hash` to reject stale writes after a previous `read`.

`list("/")` returns each permitted mount root with its `description`,
`guidance`, and `write_access` flag. Agents should use that mount context to
choose where to search, read, and save files.

## Index Lifecycle

On `mvmt serve`, mvmt starts serving immediately and rebuilds the text index in
the background.

Force a rebuild:

```bash
mvmt reindex
```

The current prototype stores the index as JSON beside the config file. The
module boundary is isolated so this can be replaced with SQLite after choosing
a dependency.

The indexer skips common generated/cache directories such as `node_modules`,
`.git`, `dist`, `build`, `.next`, `.turbo`, `coverage`, and virtualenv/cache
folders. It also caps index size so overly broad mounts truncate instead of
exhausting the Node heap.

Default caps are intentionally conservative for the JSON prototype: 5,000 files,
20,000 chunks total, and 24 chunks per file. Use narrower mounts or explicit
`exclude` rules when a broad mount truncates before the files you care about.

Run a synthetic benchmark without crawling real projects:

```bash
npm run bench:text-index -- --docs 10000
```

## Current Limits

- No SQLite dependency yet.
- No file watcher yet.
- Text-like extensions only.
- Very broad mounts may produce a truncated index; narrow the mount or add
  `exclude` rules if important files are missing.
- Removes are permanent file deletes in this prototype.
