# Scarf v2.9.2

Patch release fixing two reported bugs in remote chat and session resume.

## Fixes

### Remote Chat no longer blocks on a bare `hermes` binary ([#100](https://github.com/awizemann/scarf/issues/100))

Remote servers where `hermes` resolves on the login `PATH` (e.g. a pipx install at `~/.local/bin/hermes`) but isn't pinned to an explicit binary path showed **"Hermes Not Found — Expected at hermes"** in Chat, even though Remote Diagnostics passed 14/14 and the iOS app worked against the same server.

The pre-flight gate ran `fileExists("hermes")`, which on a remote context executes `test -e hermes` — a filesystem check against the remote working directory, not a `PATH` lookup. A bare command name always failed that check, blocking the chat UI, even though the actual ACP launch (`bash -lc`, a login shell) resolves `hermes` correctly.

`ServerContext.hermesBinaryProbablyResolvable()` now distinguishes the two cases: a **path-shaped** binary (absolute/relative, contains `/`) still gets the accurate `fileExists` check, while a **bare command name** is presumed resolvable and the authoritative check is deferred to the login-shell ACP launch — whose failure path already surfaces a "command not found" hint if `hermes` is genuinely absent. Remote chat now opens without needing a `~/hermes` symlink workaround.

### Session resume no longer silently drops context ([#99](https://github.com/awizemann/scarf/issues/99))

Reopening a chat could spawn a brand-new Hermes session — losing conversation history, task state, and loaded skills — with no error shown.

`ACPClient.loadSession` assumed any non-throwing response meant the session loaded. But Hermes's `load_session` returns a JSON-RPC **`result: null`** (not an error) when a session can't be restored into the ACP runtime. Scarf treated that null as success and proceeded against a phantom session id, so prompts landed in fresh context.

`loadSession` now throws on a null/non-dict result, routing into the existing fallback that creates a fresh session **and replays the conversation transcript from `state.db`** so the thread reads continuously. Restorable sessions resume correctly as before, keeping full context (history, task state, skills) — the fix is specifically the previously-silent not-restorable path.

## Tests

3 new ScarfCore tests: `loadSession` throws on a null result + succeeds on a dict result (mock round-trip), and `hermesBinaryProbablyResolvable` returns true for a bare remote name without a transport round-trip.

## Hermes compatibility

Targets Hermes v0.14.0 (v2026.5.16); works against v0.13+ with capability gating. No `~/.hermes/` schema changes.
