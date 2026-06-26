# Proxy-learnings batch — pre-spawn precheck + output cap + compression substrate

## Problem

While investigating three multi-LLM HTTP proxies (CLIProxyAPI, OmniRoute, 9router) for
ideas Chorus could borrow, we surfaced four concrete pain-points worth fixing before
they bite us in production:

1. **No pre-spawn validation.** When a CLI's OAuth bearer goes stale or the user is
   already over quota, Chorus spawns the subprocess anyway, waits 3–5 s for cold-start,
   then the pane scraper eventually catches `token_refresh_lost` or `quota_exhausted`.
   Confusing UX, wasted spawn tax, particularly painful on multi-reviewer phases.
2. **`phase_events.output` has no size cap.** SQLite handles big TEXT fine in
   isolation, but every SSE re-attach calls `phaseEvents.list(chatId)` which re-fetches
   every row. A long peer-review transcript could turn each re-attach into a 4 MB read.
   OmniRoute v3.4.0 hit the same shape and crashed with `SQLITE_FULL`.
3. **No documented invariant on SSE encoding.** `Content-Encoding: gzip` would batch
   bytes and break the `data: …\n\n` line parser on the client side. The CLIProxyAPI
   commit history flags this as a known proxy gotcha. We currently rely on default
   Fastify behavior — fine, but undocumented; a future "performance tweak" PR could
   regress it silently.
4. **No tool-result compression substrate.** 9router's RTK Saver compresses
   `tool_result` blocks mid-stream (base64 images → thumbnails, repeated JSON →
   deltas), reportedly saving 20–40 % tokens on long review chains. We don't need the
   compressor today, but a settings substrate now keeps the future PR small.

The fan-out research also surfaced two audits and one N/A:

5. **Body/frame mutation audit.** OmniRoute v3.7.8 had a Codex-stream-detection bug
   from in-place body mutation. `grep` on `src/daemon/` + `src/lib/` found only spread
   patterns — clean.
6. **Cache-key fingerprint stability audit.** OmniRoute v3.5.0 had a billing-cache
   thrash from content-derived keys. The only `_meta.json` writer (`agents/kimi.ts`)
   uses `chatId + binary + model + ts` — clean.
7. **Strip `x-` schema prefixes for Gemini.** N/A. We don't translate tool schemas
   between providers; Chorus drives the CLI directly and the CLI handles its own
   schemas. Only relevant if we ever expose an HTTP-compatible endpoint.

## Approach

Five surgical changes, no new abstractions. Each is independently revertable.

### 1 — Pre-spawn precheck (`src/lib/cli-precheck.ts`)

Two-layer cheap check, called from `runDoer` and `runReviewer` before any spawn:

| Layer | What | When it fires |
|---|---|---|
| Quota | Read `getHealth(lineage)`. If `status='quota_exhausted'` AND `resetAt > now`. | Skip spawn, emit `cli_warning` with `resetAt` so UI can show countdown |
| Auth | Stat the per-lineage credential file (`.claude/.credentials.json`, `.codex/auth.json`, etc). If missing or zero-byte. | Skip spawn, emit `cli_warning` with login command in `cta` |

**Deliberately cheap** — no network calls. The third layer (probe `/v1/models` with the
stored bearer to confirm token validity) is the natural next step but adds 200–500 ms
per voice and isn't worth the latency until we have evidence the file-existence check
misses real cases.

**Failure path** — both runDoer and runReviewer return `null` (existing convention for
"never produced an answer"). The phase loop already handles `null` gracefully by
counting it toward the all-reviewers-failed threshold and continuing.

**Stale markers self-clear** — if `resetAt` is in the past or absent, the precheck
falls through and lets the spawn happen. A successful run will overwrite the health
record via the existing `recordHealth` path.

### 2 — Output cap on `phase_events.output`

256 KB hard cap, head 192 KB + tail 32 KB, with a truncation marker pointing at the
chat dir. Long-form artifacts already live on disk under `~/.chorus/chats/<id>/` —
the DB row is meant to be a summary handle. Cap is enforced in both `create` and
`update` paths via a private `capOutput` helper.

Byte-length based (not char-count) because SQLite stores UTF-8 and a multi-byte run
could blow past the cap if we used `String.length`.

### 3 — SSE no-gzip comment

Single comment block at the SSE-headers writeHead site explaining why
`Content-Encoding: gzip` would break the line parser. References the CLIProxyAPI
`Accept-Encoding: identity` precedent. Anti-regression doc; no behavioral change.

### 4 — Tool-result compression settings stub

`src/lib/settings/output-compression.ts` ships three keys (`enabled` default off,
`imageThumbnailMaxKb`, `textTruncateKb`) following the same pattern as
`settings/permissions.ts`. The compressor itself is deliberately deferred — when
implemented, it reads these settings before forwarding `tool_result` between agents.
v0.7 dogfood doesn't change behavior.

### 5 — Per-CLI env-var matrix memory

`chorus_cli_env_var_matrix.md` documents the cross-CLI flag/env-var surface. Reference
for shim authors and for the future case where we route via `voices.model_id` and need
to know which CLI honors which override.

## Alternatives considered

- **Network probe in precheck** — rejected for now. Adds 200–500 ms × N voices = real
  latency. File-existence catches the highest-frequency failure mode (logged-out user)
  without the tax. Network probe earns its keep only after we see file-check false
  negatives in real use.
- **Move large outputs to filesystem with path reference** — rejected for the cap. We
  already write per-round `answer.md` artifacts to the chat dir. Duplicating the full
  output to a sidecar would double disk usage. Truncation marker + chat-dir pointer is
  cheaper and gives the user the recovery path.
- **Implement the compressor now** — rejected. The lift is non-trivial (image resize,
  JSON delta encoding, streaming integration) and the win is theoretical until we see
  review chains hitting actual cost pain. Substrate now, compressor when there's
  evidence.
- **Make precheck a runner-internal helper** — rejected. Lives in `src/lib/` as a
  pure function so tests can hit it without a runner harness, and so future MCP tools
  (e.g. "is voice X usable?") can call it without re-importing daemon code.

## Risks

| Risk | Mitigation |
|---|---|
| False-positive auth_missing on machines that store creds in non-standard paths | Multiple candidate paths per lineage; user can override via env if needed in a follow-up. Existing precheck failure is non-terminal — emits `cli_warning` and skips that voice; remaining voices proceed. |
| Output cap truncates legitimate large reviewer output | Truncation marker + on-disk artifact path means the full content is still recoverable. Cap is generous at 256 KB. |
| Precheck test relies on `process.env.HOME` swap | Test isolates via `beforeEach` save+restore + explicit `delete` in afterEach. Verified clean run alongside other suite. |

## Test strategy

- **9 new precheck tests** (`tests/cli-precheck.test.ts`) — quota gate (future, past,
  missing), cred gate (missing, zero-byte, present, fallback path), per-lineage CTAs.
  Isolates `process.env.HOME` to a tempdir per test.
- **2 new output-cap tests** (`tests/db.test.ts`) — oversized input is capped with
  head/tail + marker, under-cap input passes through unchanged.
- **Existing 175 tests** must remain green — verify no async ripple from the precheck
  hooks in runner.ts.

## Out of scope

- Network-layer auth probe (next iteration if file check misses real cases)
- The actual tool-result compressor (v0.8+ when there's evidence)
- Strip `x-` schema prefixes (only relevant if we ship an HTTP-compatible endpoint)
- Subprocess pool / session reuse (state-contamination risk, not worth it)
- Internal MITM proxy for observability (complexity for nice-to-have telemetry)

## Acceptance criteria

- [x] `pnpm typecheck` clean
- [x] `pnpm test` green: 175 → 186 (+11 new tests)
- [x] No behavioral change for healthy spawns (precheck falls through when ok)
- [x] Failed precheck emits `cli_warning` event with reason + cta + optional resetAt
- [x] Truncated output preserves head + tail + marker; under-cap output passes through
- [x] Settings stub for compression is reachable but defaults to off
