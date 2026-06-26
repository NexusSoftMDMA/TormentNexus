# mvmt Architecture

mvmt is a local-first Multi-Volume Mount Transport. It exposes selected local folders through one permissioned MCP endpoint.

It is not sync, cloud storage, or a connector marketplace. Data stays where it lives. mvmt controls which mounted paths each client can search, list, read, write, or remove.

```text
                               LOCAL MACHINE
+--------------------------------------------------------------------------------+
|                                                                                |
|  ~/.mvmt/config.yaml                                                           |
|                                                                                |
|  mounts:                                                                       |
|    /notes      -> ~/Documents/Obsidian                                         |
|    /workspace  -> ~/code/mvmt                                                  |
|                                                                                |
|  clients:                                                                      |
|    codex can read/write /workspace                                             |
|    chatgpt can search/read /notes                                              |
|                                                                                |
|  +-----------------------+                                                     |
|  | mvmt serve            |                                                     |
|  |                       |                                                     |
|  | HTTP: 127.0.0.1:4141  |                                                     |
|  | stdio: optional       |                                                     |
|  +-----------+-----------+                                                     |
|              |                                                                 |
|              v                                                                 |
|  +-----------------------+                                                     |
|  | Security layer        |                                                     |
|  |                       |                                                     |
|  | bearer/OAuth auth     |                                                     |
|  | origin guard          |                                                     |
|  | client identity       |                                                     |
|  | path/action policy    |                                                     |
|  +-----------+-----------+                                                     |
|              |                                                                 |
|              v                                                                 |
|  +-----------------------+                                                     |
|  | Tool router           |                                                     |
|  |                       |                                                     |
|  | search/list/read      |                                                     |
|  | write/remove          |                                                     |
|  | audit log             |                                                     |
|  +-----------+-----------+                                                     |
|              |                                                                 |
|              v                                                                 |
|  +-----------------------+                                                     |
|  | Text context index    |                                                     |
|  | Mount registry        |                                                     |
|  | Local folder provider |                                                     |
|  +-----------+-----------+                                                     |
|              |                                                                 |
|              v                                                                 |
|  +-----------------------+                                                     |
|  | Selected local files  |                                                     |
|  +-----------------------+                                                     |
|                                                                                |
+--------------------------------------------------------------------------------+
```

## Runtime Paths

Local HTTP clients connect to:

```text
http://127.0.0.1:4141/mcp
```

They send:

```text
Authorization: Bearer <token>
```

Claude Desktop can launch mvmt over stdio instead:

```text
Claude Desktop -> mvmt serve --stdio -> MCP over stdio
```

Tunnel mode publishes the local HTTP endpoint through a public HTTPS URL for web clients such as ChatGPT or claude.ai. The local server still binds to `127.0.0.1`.

## Request Pipeline

```text
client
  -> transport auth
  -> origin guard
  -> client identity
  -> path/action policy
  -> tool router
  -> text index or local folder provider
  -> audit log
  -> client
```

Authentication answers who is calling.

Mount and client policy answer what that caller may access.

## Source Layout

```text
src/apps/mcp             MCP protocol router and tool modules
src/apps/dashboard       dashboard helper modules
src/apps/file-inspector  first-party reference app (HTML SPA + manifest)
src/apps/registry.ts     first-party app registry consumed by the server
src/core/leases          lease model, store, secrets, and file access
src/core/auth            ClientIdentity and auth resolution
```

`apps/*` is for code that belongs to one application surface; `core/*`
is the shared substrate any app may consume. Mount resolution, path
policy, storage, leases, auth, and audit should stay in core so apps
don't drift on access rules.

HTTP route composition still lives in `src/server/index.ts`. There is no
generic `Application` interface yet — `apps/registry.ts` is a static
first-party list, not a plugin loader. Third-party app install, dynamic
discovery, CORS, bearer-token FS access, and per-app permissions are
intentionally deferred until a real need surfaces.

What ships today is a first-party app registry plus a single launch route:
`GET /apps/<id>` returns the app's inline HTML behind the same dashboard
cookie auth. There is no `GET /apps/<id>/*` static asset surface yet —
apps that outgrow inline HTML will trigger that addition. Apps call back
into `/api/fs/*` for filesystem access; they never reach into the storage
provider directly.

Old paths under `src/dashboard/`, `src/lease/`,
`src/server/context-tools/`, and `src/server/client-identity.ts` remain as
one-line re-export shims so existing import sites keep working without
lockstep updates.

## Tool Surface

The mount runtime exposes five tools:

| Tool | Purpose | Required permission |
| --- | --- | --- |
| `search` | Search indexed text chunks across permitted mounts. | `search` |
| `list` | List visible mount roots or directories. | `read` |
| `read` | Read one text file by virtual path. | `read` |
| `write` | Create or overwrite one text file. | `write` |
| `remove` | Delete one text file. | `write` |

`write` and `remove` also require the target mount to have `writeAccess: true`, and protected paths remain blocked.

## Mount Registry

The mount registry maps virtual paths to configured mount roots.

Example:

```yaml
mounts:
  - name: workspace
    path: /workspace
    root: /Users/you/code/mvmt
```

`/workspace/docs/setup.md` resolves to a file under `/Users/you/code/mvmt/docs/setup.md`.

Path traversal and symlink escapes are rejected by the local folder provider. Clients never receive host filesystem paths in tool results.

## Text Index

The text index is used by `search`.

On startup, mvmt serves immediately and rebuilds the index in the background. `mvmt reindex` forces a full rebuild.

The current index is a JSON snapshot beside the config. SQLite and incremental file watching are planned.

## Shutdown

SIGINT or SIGTERM triggers cleanup:

- close stdio or HTTP MCP transports;
- stop the tunnel process if running;
- close the control socket;
- flush normal process cleanup.

If cleanup hangs, mvmt force-exits after the shutdown timeout.

## Future Federation

The long-term direction is mounted remote mvmt instances:

```text
/local    -> local folder
/desktop  -> remote mvmt instance
/server   -> remote mvmt instance
```

That is federation, not file sync. The first future abstraction is `resolve(path) -> AccessPlan`, followed by read-only remote mounts.
