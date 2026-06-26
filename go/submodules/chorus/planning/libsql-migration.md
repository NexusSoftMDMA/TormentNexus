# libsql migration — `better-sqlite3` → `@libsql/client`

> Pure transport swap to make `npm install -g chorus-codes` robust on Windows / locked-down dev machines without node-gyp. Out of scope: voices schema, persona schema additions.

## Problem

`better-sqlite3@^11` is a native module. `npm install -g` runs node-gyp on the user's machine when prebuilt binaries don't match (Node ABI drift, glibc mismatch, fresh Windows without VS Build Tools, locked-down corp boxes without Python). Failure modes seen across similar Electron/Tauri projects:

- Windows: `gyp ERR! find Python` / `MSBuild not found`
- macOS Apple Silicon on Intel-built prebuilt: ABI mismatch
- Alpine / Bun runtimes: missing prebuild, falls back to source build
- Node 22 → Node 24 jump: prebuild lag, source build needed

This isn't theoretical — chorus's launch story is `npm install -g chorus-codes && chorus init`. A first-impression failure on `install` kills adoption.

## Approach

Replace `better-sqlite3` with `@libsql/client` (libsql, sqlite-compatible, pure-Rust, prebuilt for every platform via napi-rs). Same SQLite file format (`~/.chorus/chorus.db` is unchanged on disk), same SQL dialect, no schema changes.

Single seam: `src/lib/db/index.ts`. Every other file uses the exported `chats / phaseEvents / templates / settings / secrets / personas / getDb` namespaces — those become async, callsites await.

## Alternatives considered

| Option | Pros | Cons | Verdict |
|---|---|---|---|
| **`@libsql/client`** | napi-rs prebuilt matrix (no node-gyp invocation on install), sqlite-compatible, async API | Async ripple; BLOB type is `ArrayBuffer` (not `Buffer`); newer ecosystem; pre-1.0 | **Picked** — best install reliability |
| `node:sqlite` (Node 22.5+) | Zero deps, ships with Node | Forces engines `>=22.5.0`; experimental flag on some lines; users on Node 20 LTS blocked | Rejected — engines bump too aggressive for launch |
| Stay on `better-sqlite3`, fix UX | Cheapest | Doesn't fix the root cause; install flake remains | Rejected — defeats the migration's purpose |
| `sql.js` (WASM) | Pure JS, works everywhere | 30× slower; full-DB-in-memory model; no WAL; would force schema rework | Rejected — perf cliff |
| `libsql` Node addon (sync) | Sync API (better-sqlite3 drop-in), napi-rs prebuilt | Sync API means migrating BACK to it would un-do the entire async ripple — NOT a single-file rollback | **Not viable as rollback after migration** |

> **Native code clarification:** Both `@libsql/client` and the `libsql` Node addon ship native binaries — neither is "pure JS." The install-reliability advantage is **napi-rs prebuilds** (no node-gyp invocation, comprehensive prebuild matrix, install-time fallback to a clear error message instead of a build attempt) versus better-sqlite3's gyp fallback chain. On exotic platforms where napi-rs lacks a prebuild, `@libsql/client` falls back to its WASM build path (no native code) — that's the tertiary safety net.

## Pre-work: test scaffolding (DO FIRST, before any swap)

Today: 11 test files, **zero touch the DB seam**. Migration without coverage = "compiles + dogfood pass or broken." Unacceptable for a transport swap touching every persistence path.

### Test isolation requirements (per cdx-1 review)

Two module-level pieces of state break naive per-test isolation:
1. **`const dbPath = path.join(dbDir, 'chorus.db')`** — evaluated once at module load, so a test changing `CHORUS_DB_PATH` after import has no effect.
2. **`let dbInstance: Database | null = null`** — singleton; once initialized in test 1, test 2 reuses the same handle even if env is changed.

**Pre-work refactor (single commit, against `better-sqlite3`, before any libsql swap):**
- Move `dbPath` resolution INSIDE `getDb()` — read `process.env.CHORUS_DB_PATH ?? path.join(os.homedir(), '.chorus', 'chorus.db')` at init time.
- Add internal `_resetDbForTests()` helper that closes the singleton and nulls it. Exported but tagged `@internal`.
- Tests use `beforeEach` to set `CHORUS_DB_PATH` to a fresh `${os.tmpdir()}/chorus-test-${ulid}.db` and call `_resetDbForTests()`; `afterEach` closes + unlinks.

### `tests/db.test.ts` surfaces

| Surface | What's tested |
|---|---|
| `getDb()` | first-call schema init on missing DB; idempotent ALTER TABLE; personas CREATE IF NOT EXISTS |
| `chats.create / getById / update / list / cancel / delete` | round-trip a chat; status enum coercion; partial update semantics; cascade on delete |
| `phaseEvents.create / list / getById / update` | append + list-by-chat ordering; `lastInsertRowid` returned correctly |
| `templates.create (INSERT OR REPLACE) / list / getById` | **`INSERT OR REPLACE` overwrites `created_at` to `now`** — test asserts current behavior; transport swap must preserve it. YAML BLOB→string coercion. |
| `settings.get / set / getAll` | JSON-string round-trip; non-JSON string fallback |
| `secrets.set / get / list` | meta nullable; list omits `value` |
| `personas.upsert / list / getById / delete` | **preserves `created_at` on re-upsert** (uses explicit existing-row read, unlike templates) |

> **Note on templates created_at semantics:** the current code at [src/lib/db/index.ts:352](src/lib/db/index.ts#L352) does `INSERT OR REPLACE INTO templates (..., created_at, updated_at) VALUES (..., now, now)` — so re-creating a template wipes `created_at`. This is existing behavior; the test asserts it as-is. Personas use a different upsert pattern that DOES preserve created_at via explicit read-before-write — the test asserts that distinction too. Out of scope to change either.

**Runner:** vitest (chorus already uses it). Each test creates a fresh temp DB at `${os.tmpdir()}/chorus-test-${ulid}.db`, sets `CHORUS_DB_PATH`, calls `_resetDbForTests()`, runs assertions, deletes the file in `afterEach`.

This suite must be GREEN against `better-sqlite3` first. Then the migration is a refactor — same tests, new transport. **Tests are the regression net.**

## Migration plan

### 1. Callsite inventory

**Hot-path** (per-chat, per-phase-event):
- `src/daemon/index.ts` — `phaseEvents.create()` per state transition; `chats.update()` on terminal events; `phaseEvents.list(chatId)` reads on every SSE poll/decision
- `src/daemon/agents/*.ts` — none directly; agents return data, daemon writes

**Warm-path** (per-page-load / per-API-request):
- `src/daemon/routes/templates-personas.ts` — list/get on cockpit nav
- `src/daemon/routes/settings.ts` — getAll on settings page; set on every toggle
- `src/daemon/routes/system.ts` — `chats.list({ status: 'blocked' })`

**Cold-path** (init / once-per-session):
- `src/cli/index.ts` — `getDb()` + `templates.create()` during `chorus init` builtin seeding
- `src/lib/personas.ts` — `seedBuiltinPersonas()` on daemon boot
- `src/lib/settings/{transport,permissions,billing}.ts` — module-load-time reads
- `src/lib/cli-health.ts` — health snapshot persistence

All HTTP/MCP entrypoints are already async (Fastify + MCP SDK). The CLI subcommands are async. The settings helper modules are sync today — they need `await` at call sites or to wrap reads in async helpers.

### 2. API mapping (`better-sqlite3` → `@libsql/client`)

| `better-sqlite3` (sync) | `@libsql/client` (async) |
|---|---|
| `db.prepare(sql).run(...args)` | `await db.execute({ sql, args })` returns `{ rowsAffected, lastInsertRowid }` |
| `db.prepare(sql).get(...args)` | `(await db.execute({ sql, args })).rows[0] ?? null` |
| `db.prepare(sql).all(...args)` | `(await db.execute({ sql, args })).rows` |
| `db.exec(multi-statement SQL)` | `await db.executeMultiple(sql)` |
| `db.pragma('journal_mode = WAL')` | `await db.execute('PRAGMA journal_mode = WAL')` (libsql defaults to WAL anyway) |
| `db.transaction(fn)(args)` | `await db.batch([{sql, args}, ...], 'write')` for atomic multi-statement |

### 3. Prepared-statement caching

better-sqlite3 caches by `db.prepare(sql)` — same SQL string returns same compiled statement. libsql doesn't expose a prepare API; `execute()` is the unit. Two implications:

- **No memoization layer to maintain.** Drop `stmt = db.prepare(...)` patterns; inline `execute({ sql, args })`.
- **Perf:** libsql parses SQL per call. Benchmarked perf delta on local SQLite is single-digit µs per query — irrelevant for chorus's read/write rates (max ~10 QPS during a chat run; UI is request/response).

If perf regresses on hot path (`phaseEvents.list(chatId)` on a 200-event chat), fallback is the `libsql` package (Node addon, prebuilt, also async). Same client surface, same library family. Documented as the rollback lever in `src/lib/db/index.ts` comment header.

### 4. Transactions

Two places need multi-statement atomicity:

- **`chats.delete(id)`** — DELETE phase_events + DELETE chats. Currently two sequential `stmt.run()` calls. Migration: use `db.transaction()` callback (libsql ≥0.14) for explicit BEGIN/COMMIT, NOT `db.batch()`:
  ```ts
  const tx = await db.transaction('write');
  try {
    await tx.execute({ sql: 'DELETE FROM phase_events WHERE chat_id=?', args: [id] });
    await tx.execute({ sql: 'DELETE FROM chats WHERE id=?', args: [id] });
    await tx.commit();
  } catch (e) { await tx.rollback(); throw e; }
  ```
  Reason for not using `batch()`: while `db.batch([...], 'write')` is documented to wrap in BEGIN/COMMIT for the local-file case, the explicit `transaction()` API is unambiguous and works identically across local + remote. Tests must include a "kill mid-transaction" scenario verifying no orphaned phase_events.
- **`getDb()` first-time init** — schema.exec() + ALTER TABLEs + personas CREATE. Currently sync sequence. Migration: wrap in a single `db.transaction()` callback. Already idempotent so partial failure is recoverable.

### 5. BLOB / Buffer handling

`coerceTemplateYaml()` checks `Buffer.isBuffer(r.yaml)` because admin tools insert via `INSERT ... readfile(...)` producing a SQLite BLOB column. better-sqlite3 surfaces it as Node `Buffer`. **`@libsql/client` surfaces it as `ArrayBuffer`** (empirically verified during plan review by running `instanceof` checks against an actual libsql client — `ArrayBuffer` is NOT an instance of `Uint8Array`, so the naive Uint8Array check would silently pass through and fail Zod's string parse downstream).

```ts
function coerceTemplateYaml(row: unknown): unknown {
  if (!row || typeof row !== 'object') return row;
  const r = row as Record<string, unknown>;
  // libsql surfaces BLOB as ArrayBuffer; check it FIRST.
  if (r.yaml instanceof ArrayBuffer) {
    return { ...r, yaml: new TextDecoder().decode(new Uint8Array(r.yaml)) };
  }
  // Node Buffer (legacy better-sqlite3 reads or Buffer-typed args) — extends Uint8Array.
  if (r.yaml instanceof Uint8Array) {
    return { ...r, yaml: new TextDecoder().decode(r.yaml) };
  }
  return r;
}
```

Test fixture: insert a real BLOB via `INSERT INTO templates (id, source, yaml, ...) VALUES (?, ?, X'...', ...)` and assert read returns a string after coercion.

### 6. `getDb()` becomes async with init guard + reject-recovery

```ts
import { createClient, Client } from '@libsql/client';

let dbInstance: Client | null = null;
let dbInitPromise: Promise<Client> | null = null;

export async function getDb(): Promise<Client> {
  if (dbInstance) return dbInstance;
  if (dbInitPromise) return dbInitPromise;
  dbInitPromise = initDb()
    .then((db) => { dbInstance = db; return db; })
    .catch((err) => {
      // CRITICAL: clear the cached promise on failure. Without this, a
      // single transient init error (corrupted DB, FS hiccup, permission
      // glitch) would lock the daemon forever — every subsequent getDb()
      // call returns the same rejected promise until restart. With the
      // catch, the next caller retries from scratch.
      dbInitPromise = null;
      throw err;
    });
  return dbInitPromise;
}

// @internal — for tests only. Closes the singleton + clears the init guard.
export async function _resetDbForTests(): Promise<void> {
  if (dbInstance) await dbInstance.close();
  dbInstance = null;
  dbInitPromise = null;
}
```

The `dbInitPromise` guard prevents two concurrent first-callers from racing past the `isNew` check. The `.catch` clears the cache on failure so transient errors are retryable. `_resetDbForTests` enables proper per-test isolation.

### 7. Async ripple — concrete callsite changes

All DB exports become async. Example:

```ts
// before
const chat = chats.create({ work, template_id });
// after
const chat = await chats.create({ work, template_id });
```

> **Critical pattern for settings helpers (per cdx-1 review):** `settings.get(key)` returns `unknown`. After the swap it returns `Promise<unknown>`. Code like:
> ```ts
> const raw = settings.get(KEY);                     // becomes Promise<unknown>
> const parsed = TransportSchema.safeParse(raw);     // safeParse(promise) → success: false
> resolved = parsed.success ? parsed.data : DEFAULT; // ← silently uses DEFAULT
> ```
> would compile cleanly under tsc strict (`unknown` allows assignment), but would silently DROP the user's stored setting because `safeParse` rejects the Promise object. **This is the highest-risk callsite class.** Mitigation:
> 1. Make every settings helper explicitly async with explicit return types (`async function getTransport(): Promise<Transport>`).
> 2. Add `eslint-plugin-promise` with `no-floating-promises` enabled — catches `settings.get(key)` not awaited.
> 3. Test coverage: each settings helper (`getTransport`, `getPermissions`, `getBilling`, `recordHealth`, `readHealth`) gets a focused test that round-trips a non-default value (set → restart → get → assert).

Caller updates:
- **`src/daemon/index.ts`** — route handlers are already `async (request, reply) => {...}`. Just `await`.
- **`src/cli/index.ts`** — every command body is already async (Commander async actions). Add `await`.
- **`src/lib/settings/{transport,permissions,billing}.ts`** — convert each export to `async function` with explicit `Promise<T>` return type. Lazy-promise memoization NOT applicable here because each call must read live DB state (settings can change at runtime via cockpit toggles).
- **`src/lib/cli-health.ts`** — `recordHealth` / `readHealth` become async. Callers in onboarding flow + status panel API — already async.
- **`src/lib/personas.ts`** — `seedBuiltinPersonas()` becomes async. Called once from daemon boot — already in async context.

No callsite is in a sync hot loop. The ripple is mechanical. tsc strict + no-floating-promises catches missed `await`s.

### 8. Node engines

No bump. `@libsql/client` supports Node 18+. Current `package.json` has `engines.node: ">=18.18.0"` — stays.

### 9. Postinstall + first-run UX

No changes. `scripts/postinstall.mjs` is purely informational ("chorus init" hint). Removing the node-gyp dependency means it just works — no new failure UX needed.

### 10. Rollback plan

If a hot-path perf regression ships in dogfood (e.g., 200-event `phaseEvents.list` doubles), the rollback is a **clean revert of the PR** — full branch was a transport swap, so reverting the PR restores `better-sqlite3` semantics in one operation. We do NOT downgrade to the sync `libsql` Node addon as an intermediate step: that package's API is sync (better-sqlite3 drop-in), so switching to it would require unwinding every `await` we added — strictly more work than reverting.

The decision tree on perf regression:
1. **Mild (<2× regression on hot path):** ship as-is, file follow-up to optimize.
2. **Severe (>2×):** revert the PR; cut a v0.7.x patch with `better-sqlite3` restored; re-evaluate transport choice.

## Edge cases

1. **Existing `~/.chorus/chorus.db`** written by `better-sqlite3` — file format is identical SQLite3. libsql opens it cleanly. PoC: open dev DB through libsql before merging. **WAL recovery:** if the existing DB has uncheckpointed WAL frames from an unclean better-sqlite3 shutdown, SQLite (and libsql, being SQLite-compatible) runs WAL recovery automatically on next open — no data loss. Belt-and-braces: pre-merge PoC includes a test against a DB with active `-wal` file present (we kill -9 the dev daemon, swap, re-open, verify all rows readable).
2. **Concurrent daemon startups** — `dbInitPromise` guard handles intra-process race. Cross-process is impossible (single daemon, single port).
3. **Init failure recovery** — `dbInitPromise` `.catch` clears the cache (see §6); transient FS / permission errors are retryable on next call instead of locking the daemon until restart.
4. **Schema-init failure mid-statement** — schema.sql is `IF NOT EXISTS` throughout. Re-run safe.
5. **WAL mode** — libsql defaults to WAL on local file URLs. Drop the explicit `PRAGMA` (or keep for clarity, no-op).
6. **`lastInsertRowid` type** — better-sqlite3 returns `number | bigint`. libsql returns `bigint | undefined`. `phaseEvents.create` casts to `number` — needs `Number(result.lastInsertRowid)` with a guard.
7. **Boolean coercion** — `yolo` and `builtin` columns are stored as `0/1` integers. better-sqlite3 returns `number`; libsql returns `number | bigint`. Existing Zod `z.coerce.boolean()` handles both.

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Async ripple breaks a sync callsite missed in audit | Medium | Pre-work tests catch most; tsc strict catches the rest (Promise<X> assigned to X = type error) |
| libsql perf regression on hot path | Low | Pre-flag rollback to `libsql` package; benchmark `phaseEvents.list` against 500-row dataset before merge |
| Buffer→Uint8Array coercion misses a callsite | Low | One known site (templates BLOB); grep for `Buffer.isBuffer` confirms; new test fixture exercises the path |
| Schema-init race under load | Low | `dbInitPromise` guard; daemon is single-process |
| Existing user DBs incompatible | Very low | Same SQLite file format; PoC verified |
| `lastInsertRowid` bigint→number truncation | Very low | chorus row counts <<2^53; explicit `Number()` cast with sanity check |

## Test strategy

1. **Pre-migration commit A (against `better-sqlite3`, no transport change):**
   - Refactor `getDb()` to read `CHORUS_DB_PATH` env at init time (not module load).
   - Add `_resetDbForTests()` helper.
   - Write `tests/db.test.ts` covering all 7 surfaces (chats, phaseEvents, templates with INSERT-OR-REPLACE wipe-on-replace assertion, settings round-trip, secrets, personas with preserve-created-at assertion, BLOB coercion fixture).
   - Add settings-round-trip tests (`tests/settings.test.ts`): set non-default value → reset DB → get → assert. Covers transport, permissions, billing, cli-health.
   - Add `eslint-plugin-promise` with `no-floating-promises` enabled.
   - All tests GREEN against `better-sqlite3`. **This is the regression net.**
2. **Migration commit B:** Add `@libsql/client@^0.14.0` (pinned minor, pre-1.0 risk noted). Refactor `src/lib/db/index.ts` only — async exports, ArrayBuffer BLOB coercion, `db.transaction()` for atomic deletes, init-promise reject recovery. Run `tests/db.test.ts` — must stay GREEN.
3. **Migration commit C:** Update all callsites to `await`. Convert settings helpers to explicit async. Run full `pnpm test` — must stay GREEN. `pnpm typecheck` clean. `pnpm lint` clean (no-floating-promises catches anything missed).
4. **Migration commit D:** Remove `better-sqlite3` and `@types/better-sqlite3`. Re-run tests + `pnpm build:server` + `pnpm typecheck`.
5. **Pre-merge daemon boot test:** `node dist/cli/index.js init && node dist/cli/index.js start` against a temp dir, verify daemon boots + serves `/health` + seedBuiltinPersonas runs without error. Catches import-time async bugs unit tests on the DB module alone could miss.
6. **Pre-merge dogfood:** `npm pack` + `npm install -g <tarball>` + `chorus init` + `chorus start` + run a real 3-reviewer chat. Verify `_meta.json` sidecars + run page renders + cancel/resume work.
7. **Pre-merge perf check:** Seed a chat with 200 phase events, measure `phaseEvents.list(chatId)` latency before/after. Acceptable: <2× regression. Severe regression triggers full PR revert (see §10).
8. **Pre-merge atomicity test:** in `tests/db.test.ts`, simulate `chats.delete` failure mid-transaction (kill the libsql client between the two DELETEs) and verify no orphan phase_events rows survive.

## Out of scope

- Voices schema additions (ROADMAP.md §2)
- Persona schema additions
- Schema.sql edits beyond what's already there
- Any logic change in `chats / phaseEvents / templates / settings / secrets / personas` — pure transport swap, semantics unchanged
- Web UI / cockpit changes

## Acceptance criteria

- [ ] `better-sqlite3` removed from dependencies + devDependencies
- [ ] `@libsql/client` added at `^0.14.0`
- [ ] `tests/db.test.ts` exists and passes (all 7 surfaces incl. INSERT-OR-REPLACE created_at wipe + persona created_at preserve + BLOB-from-ArrayBuffer coercion + atomicity-on-delete-failure)
- [ ] `tests/settings.test.ts` round-trips persisted values (transport, permissions, billing, cli-health)
- [ ] `eslint-plugin-promise` installed; `no-floating-promises` rule active; `pnpm lint` clean
- [ ] Full `pnpm test` GREEN
- [ ] `pnpm typecheck` clean
- [ ] `pnpm build:server` succeeds
- [ ] Daemon boot smoke: `chorus init && chorus start` against temp dir reaches `/health` GREEN
- [ ] `npm pack` produces a tarball; `npm install -g <tarball>` succeeds on a clean dir without node-gyp invocation
- [ ] `chorus init` + `chorus start` + 3-reviewer real chat completes end-to-end on the published install path
- [ ] No perf regression >2× on `phaseEvents.list` for a 200-event chat
- [ ] Revert-as-rollback documented in `src/lib/db/index.ts` comment header
- [ ] PoC verified: dev DB with active `-wal` file from a kill -9'd better-sqlite3 daemon opens cleanly through `@libsql/client` with all rows readable

## Reviewer agreement (round 1)

Multi-LLM plan review fan-out: cdx-1 (gpt-5.5) + gem-1 (gemini-3.1-pro-preview) + deepseek (opencode-go/deepseek-v4-pro). Decision: `disagree` round 1 — 11 findings raised, 9 valid (3 HIGH, 5 MED, 1 LOW), 1 partial, 1 duplicate.

| # | Sev | From | Finding | Status |
|---|---|---|---|---|
| 1 | HIGH | cdx-1 | Settings reads → `Promise<unknown>` fail Zod safeParse silently → user's stored values dropped | **Addressed** §7 critical-pattern callout + lint rule + settings round-trip tests |
| 2 | HIGH | gem-1 | `dbInitPromise` caches rejected promise forever → daemon hangs on transient init failure | **Addressed** §6 `.catch` clears cache |
| 3 | HIGH | gem-1 | `@libsql/client` returns `ArrayBuffer` not `Uint8Array` for BLOB (empirically verified during review) | **Addressed** §5 ArrayBuffer-first coercion |
| 4 | MED | cdx-1 | Plan claimed templates upsert preserves created_at — actual code wipes it via `INSERT OR REPLACE ... VALUES (..., now, now)` | **Addressed** test now asserts current wipe-on-replace behavior |
| 5 | MED | cdx-1 | `CHORUS_DB_PATH` won't isolate — `dbPath` const + singleton at module load | **Addressed** pre-work commit moves path into `getDb()` + adds `_resetDbForTests()` |
| 6 | MED | gem-1 | Sync `libsql` package can't serve as "single-file rollback" — its API is sync, would un-do the entire async ripple | **Addressed** §10 reframed: rollback = clean PR revert, not partial swap |
| 7 | MED | deepseek | Settings module sync→async pattern underspecified | **Duplicate** of #1 (same concern, lower severity wording) |
| 8 | MED | deepseek | `db.batch()` atomicity for `file:` URLs not verified | **Addressed** §4 switched to explicit `db.transaction()` + atomicity test |
| 9 | MED | deepseek | Alternatives table mislabels `@libsql/client` as "no gyp" — both packages have native code; real distinction is napi-rs | **Addressed** alternatives table clarification + native-code note |
| 10 | LOW | deepseek | WAL checkpoint before swap | **Addressed** edge cases #1 — SQLite WAL recovery handles this; PoC includes `-wal` fixture as belt-and-braces |
| 11 | LOW | deepseek | Missing daemon boot-path integration test | **Addressed** test strategy step 5 |

Round 2 was skipped (user decision) — fixes are direct addresses of validated findings, no architectural changes required. Implementation proceeds against this revised plan.
