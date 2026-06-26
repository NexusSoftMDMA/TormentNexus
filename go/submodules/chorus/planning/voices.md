# Voices abstraction — table + auto-populate + UI surfaces

> First of three PRs in the v0.7 Voices × Personas × Phases trio. Scope: replace the per-lineage `<lineage>.enabled_models` settings substrate with a proper `voices` table backed by the libsql DB. Out of scope: OpenRouter inline (next PR), Phase composition (PR after that).

## Problem

The v0.7 hardening cycle introduced `<lineage>.enabled_models: string[]` settings keys (one per CLI lineage — `claude.enabled_models`, `codex.enabled_models`, `gemini.enabled_models`, `kimi.enabled_models`, `opencode.enabled_models`) as a substrate for picking which models per CLI count as "voices" in templates. Five UI surfaces read those settings: home `cli-status-panel`, home `lineage-fleet-card`, home `opencode-fleet-card`, connect `orchestrator-card`, onboarding picker, template editor.

That substrate has hit its ceiling:

1. **No structured per-voice metadata** — settings stores model IDs as strings; no place to attach lineage tagging, cost per million tokens, label, or human-readable provider distinct from CLI binary.
2. **API-routed models can't ride the same surface** — OpenRouter (next PR), Anthropic API direct, etc. don't fit into "CLI binary X has enabled_models[]" because they aren't CLIs at all. They're API providers with hundreds of models.
3. **Diversity scoring needs lineage on the row** — the template designer's "you have anthropic + openai — consider adding moonshot for spread" pitch reads from per-voice lineage. Settings strings don't carry it; we recover lineage from the parent key (`claude.enabled_models[i]` → lineage=anthropic) which couples the data shape to the storage shape.
4. **Phase composition (next-next PR) needs voice IDs** — `template_phases.voice_id` references will need a stable PRIMARY KEY surface. Settings strings have no IDs — we'd be storing `"opencode-go/kimi-k2.6"` as a foreign-key everywhere.

## Approach

Add a `voices` table to the libsql schema:

```sql
CREATE TABLE IF NOT EXISTS voices (
  id TEXT PRIMARY KEY,
  label TEXT NOT NULL,
  source TEXT NOT NULL,                  -- 'cli' | 'api'
  provider TEXT NOT NULL,                -- 'claude-code' | 'codex-cli' | 'gemini-cli'
                                         -- | 'kimi-cli' | 'opencode-cli' | 'openrouter'
                                         -- | 'anthropic-api' | ...
  model_id TEXT NOT NULL,                -- canonical qualified model ID
                                         -- single-model CLIs: latest CLI's bundled model
                                         -- multi-model CLIs: gateway-prefixed (opencode-go/kimi-k2.6)
                                         -- API: provider's canonical ID (moonshotai/kimi-k2)
  lineage TEXT NOT NULL,                 -- 'anthropic'|'openai'|'google'|'opencode'|'moonshot'
                                         -- (matches existing daemon-side Lineage enum
                                         -- in src/daemon/agents/types.ts; NOT widened here)
  vendor_family TEXT,                    -- finer taxonomy below the lineage axis:
                                         -- 'deepseek'|'meta'|'mistral'|'xai' etc.
                                         -- NULL when lineage already names the vendor
                                         -- (e.g. lineage='anthropic', vendor_family=NULL).
                                         -- Used by OpenRouter/multi-vendor gateway voices
                                         -- where lineage='opencode' but the underlying
                                         -- vendor is meaningful for cost/UX.
  input_cost_per_mtok REAL,              -- API: $/Mtok in;  CLI: NULL (subscription)
  output_cost_per_mtok REAL,             -- API: $/Mtok out; CLI: NULL
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_voices_lineage ON voices(lineage);
CREATE INDEX IF NOT EXISTS idx_voices_provider ON voices(provider);
CREATE INDEX IF NOT EXISTS idx_voices_source ON voices(source);
```

> **Lineage stability (per round 1 cdx-1+gem-2 review):** `voices.lineage` reuses the existing `Lineage` enum from `src/daemon/agents/types.ts` — `'anthropic' | 'openai' | 'google' | 'opencode' | 'moonshot'`. NOT widened to add `deepseek/meta/mistral/xai`. The `vendor_family` column carries that finer taxonomy. This avoids rippling type changes through daemon/template/UI/runner/quorum-diversity logic in this PR; lineage widening (if ever needed) is a separate scope.

### ID conventions (revised after round 1)

**Single-model CLIs use IMMUTABLE IDs that don't carry the model version**:

| Voice | id | provider | model_id | lineage | label |
|---|---|---|---|---|---|
| Claude Code | `claude-code` | `claude-code` | `claude-opus-4-7` (bumps to latest on each seed) | `anthropic` | "Claude (claude-opus-4-7)" |
| Codex CLI | `codex-cli` | `codex-cli` | `gpt-5.5` (bumps on seed) | `openai` | "Codex (gpt-5.5)" |
| Gemini CLI | `gemini-cli` | `gemini-cli` | `gemini-3.1-pro-preview` (bumps on seed) | `google` | "Gemini (gemini-3.1-pro-preview)" |
| Kimi CLI | `kimi-cli` | `kimi-cli` | `kimi-k2.6` (bumps on seed) | `moonshot` | "Kimi (kimi-k2.6)" |

This avoids the "ghost row" failure mode (gem-2 BLOCKER 3): if we used `claude-code:claude-opus-4-6` and then the CLI ships `claude-opus-4-7`, we'd get two voice rows for the same binary. Stable ID + mutable model_id+label keeps things tidy. Templates referencing `claude-code` voice automatically pick up whatever model is current.

**Multi-model CLIs (opencode) use composite IDs**: `opencode-cli:opencode-go/kimi-k2.6`. Each model is independently selectable per invocation, so each is a distinct voice with its own ID.

**API voices**: `<provider>:<model_id>` (e.g. `openrouter:moonshotai/kimi-k2`, `anthropic-api:claude-opus-4-7`).

## Auto-populate seed (daemon boot)

`src/lib/voices.ts` exposes `seedCliVoices(): Promise<{added: number; updated: number; disabled: number}>` — called from daemon main, idempotent.

### Algorithm (revised after round 1)

**Phase 1 — synchronous, pre-listen (single-model CLIs):**
1. Run CLI detect (already exists at `src/lib/cli-detect.ts`).
2. For each single-model CLI present, derive `(provider, lineage, ui_lineage)`:

   | provider | daemon-side lineage | UI-side lineage (key for `UI_LINEAGE_AVAILABLE_MODELS`) |
   |---|---|---|
   | `claude-code` | `anthropic` | `claude` |
   | `codex-cli` | `openai` | `codex` |
   | `gemini-cli` | `google` | `gemini` |
   | `kimi-cli` | `moonshot` | `kimi` |

   The `UI_LINEAGE_AVAILABLE_MODELS` map in `src/lib/lineage-maps.ts` is keyed by UI-side names (per round 2 cdx-1 BLOCKER 1). The seed must translate daemon→UI before the lookup. This translation lives in `src/lib/voices.ts` as `LINEAGE_TO_UI` constant (the inverse of the existing `mapLineage()` helper). Voice row gets the daemon-side `lineage` value; the lookup uses the UI-side key.

3. Upsert the immutable voice row for each detected CLI using `UI_LINEAGE_AVAILABLE_MODELS[ui_lineage][0]` (the recommended/default per existing convention) for `model_id` and `label`. Update `model_id` + `label` + `updated_at` if the voice row already exists. Preserve `enabled` across boots.
4. For each single-model CLI that was present last boot but is now NOT detected: set `enabled=0` (auto-disable, NOT delete). Per round 1 deepseek MED — uninstall→reboot would otherwise leave stale voices that phase composition might dispatch to.

**Phase 2 — background warmup, post-listen (opencode multi-model):**
4. After `fastify.listen()` returns, fire `seedOpencodeVoicesAsync()` as non-blocking warmup. Per deepseek LOW — `opencode models` can take up to 10s; blocking boot on it would add up-to-10s daemon-start latency on every boot. The cockpit will see partial state until the warmup lands; the home page already polls /voices on a debounce so the UI catches up.
5. The background warmup:
   - Shells out `opencode models` (10s timeout, EXISTS-graceful).
   - Iterates the FULL discovered list (not just user's `opencode.enabled_models` — per cdx-1 MED. Disabled-but-discoverable models still need rows so the user can re-enable them via the fleet card without restarting.).
   - For each model, derives lineage from gateway prefix (`opencode-go/kimi-*` → `moonshot`, `opencode-go/deepseek-*` → `opencode` with `vendor_family='deepseek'`, etc. — see edge case 7 below for the full mapping).
   - Upserts the voice with stable ID `opencode-cli:<gateway-prefix>/<model>`.
   - On first install (no migrated settings), uses `defaultPicks` = `[opencode-go/kimi-k2.6, opencode-go/deepseek-v4-pro]` for `enabled=1`; rest get `enabled=0` so they're discoverable but opt-in.
   - On `opencode-cli` not detected this boot: auto-disable existing opencode voices.

**Migration from `<lineage>.enabled_models` settings (one-shot, gated on `voices.list().length === 0`):**

For each lineage in {claude, codex, gemini, kimi, opencode}:
- **Setting key absent** → treat as "user has only the default model enabled" — that's the v0.7 fallback at `src/app/connect/page.tsx:56` (`return def ? [def] : []`). Voice for default model = `enabled=1`; other curated models = `enabled=0`.
- **Setting key empty array** → no models enabled; ALL curated models for that lineage = `enabled=0`. (User explicitly emptied the list.)
- **Setting key populated** → enabled set = exactly the listed models; remaining curated models = `enabled=0`.

The two semantics (absent vs empty) MUST be distinguished — round 1 cdx-1 MED. Tests must cover both.

**Leave the `<lineage>.enabled_models` settings keys in place** after migration — the substrate stays read-compatible for one minor (v0.8) so a downgrade doesn't strand users. The voices table is additive; future writes go to voices, but the settings keys remain inert and untouched.

## Read API

`GET /voices?lineage=<l>&source=<s>&provider=<p>&enabled=<bool>` → returns array of voices, optionally filtered. **Defaults to ALL voices (enabled and disabled)** — per round 1 cdx-1 BLOCKER. Fleet/management surfaces need the disabled rows so users can re-enable. Template selection dropdowns explicitly pass `?enabled=true` to get only the enabled subset.

Sort: provider ASC, label ASC. (Per round 1 gem-2 BLOCKER 5: fleet cards group by provider, not lineage.)

Response shape:
```ts
{ ok: true, data: VoiceRow[] }
```

Single-row variant: `GET /voices/:id` → `{ ok: true, data: VoiceRow }` or 404.

## Write API

- `PUT /voices/:id` — partial update of `{label?, enabled?, input_cost_per_mtok?, output_cost_per_mtok?}`. Returns the updated row. Writes touch `updated_at`. `source`/`provider`/`lineage`/`vendor_family`/`model_id` are immutable post-create. (`model_id` for single-model CLIs IS rewritten on each seed by the daemon, but not via PUT — the seed loop owns that column for cli-sourced rows.)
- `DELETE /voices/:id` — allowed for both `source='api'` AND `source='cli'`. Per round 1 gem-2 MED: blocking DELETE for cli rows prevents users from cleaning up deprecated OpenCode models that are no longer returned by the gateway. If the model is still active in the gateway, DELETE is reversible — the next seed re-creates the row (with the user's previous enabled state lost; that's acceptable for the rare gateway-deprecation case).
- `POST /voices` — used by the next PR's OpenRouter inline flow. Body: `{provider, model_id, label, lineage, vendor_family?, input_cost_per_mtok?, output_cost_per_mtok?}`. Generates `id = <provider>:<model_id>` (or rejects if it already exists).

`PUT` and `POST` validate via Zod. `DELETE` is FK-safe at the moment (no incoming refs). Phase composition (next-next PR) will add `template_phases.voice_id` and decide ON DELETE semantics there.

## UI surface migration

Six components currently read `<lineage>.enabled_models` from settings. They all switch to reading the voices table via a new `useVoices(opts?)` hook.

**Grouping rule (per round 1 gem-2 BLOCKER 5):** Fleet cards group voices by **provider** (the CLI binary identity), NOT by lineage. So opencode-cli's kimi+deepseek+claude-via-opencode all live in ONE OpenCode card regardless of their underlying lineage. Lineage is reserved for cross-cutting concerns (diversity scoring in the template designer, the lineage-color dot indicator).

| Component | Current | New |
|---|---|---|
| `src/components/cli-status-panel.tsx` | reads `opencode.enabled_models` to badge "3 models" | `useVoices({source:'cli'})` group-by **provider** |
| `src/components/lineage-fleet-card.tsx` | reads `<lineage>.enabled_models` for toggles | `useVoices({provider})` + per-row PUT enabled. Shows ALL voices (enabled + disabled) so user can re-enable. |
| `src/components/opencode-fleet-card.tsx` | gateway-grouped picker reading `opencode.enabled_models` | `useVoices({provider:'opencode-cli'})` grouped by gateway prefix in the model_id; PUT to toggle |
| `src/components/orchestrator-card.tsx` | per-CLI inline picker on /connect | `useVoices({provider})` + PUT toggle |
| `src/app/onboarding/page.tsx` | persists `opencode.enabled_models` on submit | seedCliVoices runs on daemon boot; onboarding calls `PUT /voices/:id` per user toggle. Per deepseek MED — the flow is: (1) onboarding loads, (2) GET /voices returns all detected, (3) user picks/unpicks, (4) on submit, batch PUT individual rows. |
| `src/components/phase-editor.tsx` | model dropdown reads `<lineage>.enabled_models` | `useVoices({lineage, enabled:true})` — explicit enabled filter for the dropdown context |

The substrate keys (`<lineage>.enabled_models`) remain readable for 1 minor — but new writes go to voices table. The migration step above seeds voices from those keys exactly once; after that the keys are inert.

## Test scaffolding

New `tests/voices.test.ts` covering:

| Surface | Cases |
|---|---|
| Schema init | voices table created on fresh DB; idempotent on re-init |
| seedCliVoices | seeds one voice per (claude-code, codex-cli, gemini-cli, kimi-cli) lineage + N voices for opencode-cli; second call doesn't dup; preserves enabled=0 across calls |
| Migration from settings | with `<lineage>.enabled_models` populated and voices empty, first seed migrates correctly; with voices already populated, settings is ignored |
| `voices.upsert/get/list/update/delete` | CRUD round-trip; lineage/source/provider/enabled filters; updated_at bumps; created_at preserved on update; DELETE allowed for cli-sourced rows; auto-heal: deleting a still-detected cli voice and re-running seed re-creates it |
| HTTP routes | GET / GET /:id / PUT / POST / DELETE happy path + error paths (400 on bad input, 404 on missing) |

`tests/db.test.ts`: extend the existing schema-init test to include `voices` in the expected table list. Per cdx-1 round-2 review of libsql migration, the schema-init assertion is the regression net for new tables.

`tests/settings-helpers.test.ts`: unchanged — settings substrate stays read-compatible for the deprecation window.

## Edge cases

1. **Existing user DB without voices table**: schema's `CREATE TABLE IF NOT EXISTS voices` runs in `getDb()` init. Migration from settings is gated on `voices.list().length === 0` so it fires once.
2. **User downgrades chorus to v0.7-pre-voices**: settings keys still exist; old code paths still work. Upgrade-then-downgrade is safe.
3. **Concurrent first-boot seeds**: `dbInitPromise` already serializes init; voices seed runs inside daemon main, which only fires once per process.
4. **CLI versioned model upgrade**: e.g. user's voices row is `claude-code` with `model_id='claude-opus-4-6'`. The CLI ships `claude-opus-4-7` and `UI_LINEAGE_AVAILABLE_MODELS.claude[0]` updates to match. On next seed, the row at id=`claude-code` keeps its primary key but `model_id` + `label` are rewritten to the new model and `updated_at` bumps; `enabled` is preserved. Templates referencing `claude-code` automatically pick up the new model on next dispatch — no user action required, no ghost rows.
5. **OpenCode user with no opencode-cli installed**: `opencode models` shells out — handle ENOENT gracefully (skip opencode voices, log once at info level, don't crash boot).
6. **`opencode models` slow on cold start**: timeout at 10s (already enforced in existing system route). On timeout, skip seeding opencode voices for this boot; retry on next boot. Don't crash daemon.
7. **Lineage + vendor_family extraction from `opencode-go/kimi-k2.6`**: gateway prefix doesn't determine lineage on its own; the model name does. Define a small lookup in `src/lib/voices.ts` that returns `(lineage, vendor_family)`. **Lineage stays in the existing 5-enum** per round 1 cdx-1+gem-2 review:

   | Model name pattern | lineage | vendor_family |
   |---|---|---|
   | `*kimi-*` | `moonshot` | NULL |
   | `*claude-*` | `anthropic` | NULL |
   | `*gpt-*` | `openai` | NULL |
   | `*gemini-*` | `google` | NULL |
   | `*deepseek-*` | `opencode` | `deepseek` |
   | `*llama-*` / `*meta-*` | `opencode` | `meta` |
   | `*mistral-*` / `*mixtral-*` | `opencode` | `mistral` |
   | `*grok-*` / `*xai-*` | `opencode` | `xai` |
   | unmatched | `opencode` | NULL |

   Same pattern map across `opencode-go/`, `opencode-zen/`, `opencode/` gateway prefixes — lookup is on the model-name suffix, not the gateway. Diversity scoring in the template designer (later) reads `lineage` for the canonical 5-axis spread; cost UX surfaces read `vendor_family` for the finer breakdown.

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| UI surfaces break during migration (5 components touched) | Medium | Test each component manually with chrome-devtools MCP; keep commits small (one component per commit if needed) |
| seedCliVoices races against UI reads on boot | Low | Daemon boots seed BEFORE fastify.listen — no HTTP requests can hit /voices until seed completes |
| Settings substrate kept in place creates dual-write risk | Medium | Mitigation: ALL writes go to voices table; settings is read-only post-migration. Document in code comments. |
| OpenRouter inline (next PR) blocks on this | N/A | This is by design — voices table is the dependency |

## Test strategy

1. **Pre-implementation**: confirm `tests/db.test.ts` schema-init assertion list extends to include `voices`.
2. **Implementation commit 1**: schema + db helpers (`src/lib/db/index.ts` voices export) + tests/voices.test.ts CRUD subset GREEN.
3. **Implementation commit 2**: HTTP routes + route tests GREEN.
4. **Implementation commit 3**: seedCliVoices + migration from settings + idempotency tests GREEN.
5. **Implementation commit 4**: UI surface migrations (5 components). Test each manually with chrome-devtools after each.
6. **Pre-merge**: full `pnpm test` GREEN, `pnpm typecheck` clean, `pnpm build:server` succeeds, daemon boot smoke shows seeded voices via `/voices` endpoint, real chat via existing template still works (regression check).

## Out of scope

- OpenRouter inline (next PR — `feat/openrouter-inline`)
- Phase composition (PR after — `feat/phase-composition`)
- Per-voice persona overrides
- Voices marketplace
- Cost-based pre-flight estimation in template designer (depends on voices.cost_*)
- Dropping the `<lineage>.enabled_models` settings keys (deprecation in v0.8)
- ON DELETE FK semantics for future phase_voice_id refs

## Acceptance criteria

- [ ] `voices` table exists in `src/lib/db/schema.sql` + idempotent CREATE in `getDb()` init (incl. `vendor_family` column)
- [ ] `tests/voices.test.ts` covers CRUD + seed idempotency + migration semantics (absent vs empty key) + version bump test (single-model CLI's model_id rewrites without ID change) + auto-disable on CLI uninstall
- [ ] `tests/db.test.ts` schema-init list includes `voices`
- [ ] `seedCliVoices()` runs synchronously pre-listen for single-model CLIs; `seedOpencodeVoicesAsync()` runs post-listen as background warmup
- [ ] First-boot migration from `<lineage>.enabled_models` settings populates voices correctly with the absent-vs-empty distinction; test fixtures cover both
- [ ] All 6 UI components read from `/voices` (grouped by **provider**, not lineage); fleet cards show enabled+disabled rows for re-enable workflow
- [ ] `GET /voices` defaults to all rows; `?enabled=true` for template-dropdown contexts
- [ ] `GET/PUT/POST/DELETE /voices` endpoints work; DELETE allowed on cli-sourced rows
- [ ] `pnpm test` GREEN (existing 127 + new ~30 = ~157)
- [ ] `pnpm typecheck` clean
- [ ] `pnpm build:server` succeeds
- [ ] Daemon boot latency unchanged (synchronous phase doesn't block on opencode shell-out — verify via timing)
- [ ] Real chat dogfood: existing template using `claude` reviewer still runs end-to-end after migration

## Reviewer agreement (round 1)

Multi-LLM plan review fan-out: cdx-1 (gpt-5.5) + gem-2 (gemini-3.1-pro-preview) + deepseek (opencode-go/deepseek-v4-pro). Decision: `disagree` round 1 — 12 findings raised (2 CRITICAL + 3 HIGH + 5 MEDIUM + 2 LOW), all valid architectural concerns.

| # | Sev | From | Finding | Status |
|---|---|---|---|---|
| 1 | HIGH | cdx-1 | `GET /voices` default `enabled=true` hides disabled rows | **Addressed** — flipped to all-by-default; explicit `?enabled=true` for dropdowns |
| 2 | HIGH | cdx-1 | New lineages (deepseek/meta/mistral/xai) require widening daemon `Lineage` type | **Addressed** — added `vendor_family` column; voices.lineage stays in existing 5-enum |
| 3 | CRITICAL | gem-2 | Versioned model upgrades create ghost duplicates for single-model CLIs | **Addressed** — immutable IDs (`claude-code`, `codex-cli`, etc.); model_id+label rewritten on each seed |
| 4 | CRITICAL | gem-2 | Renaming voices.lineage from CLI names to corporate names breaks templates | **Addressed** — voices.lineage uses existing daemon-side enum (already corporate: anthropic/openai/google/opencode/moonshot); no rename |
| 5 | HIGH | gem-2 | Grouping fleet UI by lineage scatters multi-vendor OpenCode | **Addressed** — fleet cards group by `provider`; lineage reserved for diversity scoring |
| 6 | MED | cdx-1 | OpenCode seed should iterate FULL discovered list, not just user's enabled_models | **Addressed** — algorithm revised |
| 7 | MED | cdx-1 | Migration semantics for absent vs empty enabled_models key | **Addressed** — explicit semantics: absent → default model enabled; empty → none |
| 8 | MED | gem-2 | Blocking DELETE for cli rows prevents deprecated-model cleanup | **Addressed** — DELETE allowed for cli; auto-heals on next seed if still detected |
| 9 | MED | deepseek | Onboarding flow underspecified | **Addressed** — explicit (1) seed runs first, (2) GET /voices, (3) user toggles, (4) batch PUT |
| 10 | MED | deepseek | Uninstalled CLI voices should auto-disable, not stay enabled | **Addressed** — seed phase 1 step 3 |
| 11 | LOW | deepseek | Schema/prose contradicted on model_id NULL | **Addressed** — `NOT NULL` throughout |
| 12 | LOW | deepseek | seedCliVoices blocking on `opencode models` adds 10s boot latency | **Addressed** — opencode warmup moved post-listen |

## Reviewer agreement (round 2)

Re-ran with prior-rounds-summary. **gem-1 AGREE** ✓; cdx-1 + deepseek partial.

| # | Sev | From | Finding | Status |
|---|---|---|---|---|
| 13 | HIGH | cdx-1 | Seed indexed `UI_LINEAGE_AVAILABLE_MODELS` with daemon-side lineage names; map is keyed by UI-side names | **Addressed** — explicit `LINEAGE_TO_UI` translation table in §"Auto-populate seed" Phase 1 |
| 14 | HIGH | deepseek | Pack didn't include the revised planning/voices.md (verified — pack-meta `planning_docs: []`) | **Process issue, not plan issue** — work-pack-build heuristic missed --planning-doc on round-2 calls. Filed as work-pack-build follow-up. |
| 15 | MED | cdx-1 | Stale duplicate migration section contradicting new absent-vs-empty semantics | **Addressed** — duplicate section deleted; canonical version remains in §"Auto-populate seed" |
| 16 | MED | cdx-1 | Test scaffolding still said DELETE on cli rejected | **Addressed** — test row updated to "DELETE allowed; auto-heal" |
| 17 | MED | cdx-1 | Edge cases section still described ghost-row behavior | **Addressed** — edge case 4 rewritten with new immutable-ID semantics |

Round 3 was skipped (would chase a pack-build process bug, not catch new content). Implementation proceeds against this revised plan.
