# Chorus Roadmap

> Living plan for what's shipping. Owns the cross-version product story so nothing gets lost during repo migrations or version cuts.

---

## Core product insight

Chorus's depth comes from a three-axis combinatorial:

```
Voices  ×  Personas  ×  Phases
```

- **Voices** — the model behind the keyboard. CLI subscriptions (Claude Code, Codex, Gemini, OpenCode, Kimi) and API-routed models (anything reachable via OpenRouter, Anthropic API, etc.). Each voice has a lineage tag (anthropic / openai / google / moonshot / deepseek / meta / mistral / xai).
- **Personas** — the *role* a reviewer plays. A system prompt + worldview. Same model produces wildly different output as Sentinel (security) vs Cartographer (cross-platform) vs Accountant (cost).
- **Phases** — the position in the pipeline. Discover, Define, Develop, Deliver, Debate, Decide. User can add, remove, reorder.

A **template** is a chosen path through this 3D space — N phases, each with a (voice, persona) pair. Different combinations produce wildly different outcomes. Users will experiment to find what works for their domain (Rails refactor, React perf, Solidity audit, copy review, …) and share what wins.

---

## Status tracker

| Item | Version | Status | Notes |
|---|---|---|---|
| Audit fixes 1-7 (dogfood pre-publish) | 0.6 | ✅ DONE | shipped 2026-05-01 |
| Cross-platform CLI detection (`where` + fallback dirs + `--version` verify) | 0.6 | ✅ DONE | |
| Manual CLI path entry UI | 0.6 | ✅ DONE | |
| Persona registry — table + 10 built-ins + seed loader | 0.7 | ✅ DONE | |
| MCP `list_personas` + `invoke_persona` | 0.7 | ✅ DONE | 9 MCP tools total |
| Voices abstraction (table + auto-populate + 6 UI surfaces) | 0.7 | ✅ DONE | merged 2026-05-02 (PR #2 c758e80); 71 voices auto-seed across 5 lineages; vendor_family taxonomy |
| libsql migration (better-sqlite3 → @libsql/client, no node-gyp) | 0.7 | ✅ DONE | merged 2026-05-02 (PR #1 1441532); fixes `npm install -g` on Windows / locked-down boxes |
| Review-only mode (artifact in, no doer; substrate + cockpit + verdict persistence) | 0.7 | ✅ DONE | merged 2026-05-02 (PR #4 b21a4b8); unlocks `/work` & MCP harnesses calling chorus as a pure review service |
| Per-phase `timeoutMs` override (round-2-deferred §2) | 0.7 | ✅ DONE | merged 2026-05-02 (PR #5 9a47f23); 30s..1h schema bound, threaded through doer + reviewer + tmux + review-only synthesis |
| Token-capture rails (round-2-deferred §1) | 0.7 | ✅ DONE | merged 2026-05-02 (PR #8 3910607); `AgentEvent.message_done.usage` + capture in doer return shape — per-lineage parser updates + DB persistence + cost math are follow-ups |
| Opt-out telemetry heartbeat (round-2-deferred §4) | 0.7 | ✅ DONE | merged 2026-05-02 (PR #6 bac7445); daemon-side ping to chorus.codes once on boot + every 24h; three opt-out paths; server endpoint deploy + privacy notice are separate infra tasks |
| Structured JSON-line logger substrate (round-2-deferred §3) | 0.7 | ✅ DONE | merged 2026-05-02 (PR #7 7ce6b46); pino-shaped wire format (level/time/pid/hostname); 5 first-class call sites; mechanical migration of remaining ~50 console sites is a follow-up |
| Codex quota_exhausted detection + `CHORUS_CODEX_HOME` override | 0.7 | ✅ DONE | merged 2026-05-02 (PR #9 ee7b5fb); codex emits `ERROR: hit your usage limit` to stderr and exits 1, parseCodexExit now reads stderr+code and emits `error{kind:'quota_exhausted'}` instead of writing 0-byte answer.md. Round-1 review-only dogfood caught the regex was too loose (codex echoes user prompt to stderr → false-positive risk); fixed by anchoring on literal `ERROR:` prefix. Env override is a stopgap; proper UX is multi-account-per-CLI in v0.8. |
| SSE hijack + backpressure drain race + init log path | 0.7 | ✅ DONE | merged 2026-05-02 (PR #13 a5f6db1); fastify `reply.hijack()` before `reply.raw` writes, drain handler clears `paused` first then flushes, `chorus init` log honours `CHORUS_DB_PATH`. cdx-1 round-2 follow-ups from libsql migration. |
| Cockpit retry button + correct reviewer count | 0.7 | ✅ DONE | merged 2026-05-02 (PR #14 0ab62e4) |
| Templates `liveYaml` precedence test coverage | 0.7 | ✅ DONE | merged 2026-05-02 (PR #15 02e0820) — Fix C from PR #10 |
| Time + token chips on participant cards | 0.7 | ✅ DONE | merged 2026-05-02 (PR #16 b1e2f09); reads from `_stats.json` sidecar |
| Personas wire into doer + reviewer slots (per-slot binding) | 0.7 | ✅ DONE | merged 2026-05-02 (PR #17 969dca5); each phase row carries its own persona — removes the §1 "one-persona-overlay-on-all-voices" limitation |
| CLI orphan reap on `chorus stop` + `start` | 0.7 | ✅ DONE | merged 2026-05-02 (PR #18 3eaf166) |
| `/rerun` server-side guard + retry button surfaces ok:false | 0.7 | ✅ DONE | merged 2026-05-02 (PR #19 a682e14) |
| Persona injection fence + missing-persona warning + stats SSE cleanup | 0.7 | ✅ DONE | merged 2026-05-02 (PR #20 0758791); emits `cli_warning{kind:'persona_missing'}` on lookup failure |
| Auto-fire chat runner on POST /chats (closes deferred #11) | 0.7 | ✅ DONE | merged 2026-05-02 (PR #21 b372bec); MCP-driven flows no longer sit `drafting` until a human opens the URL — runChat fires immediately on POST, SSE becomes a passive subscriber on a per-chat event bus that replays from `phase_events`. |
| MCP `wait_for_chat` progress-notification keepalive | 0.7 | ✅ DONE | merged 2026-05-02 (PR #22 1940964); `notifications/progress` every 30s past Anthropic's 60s default tool timeout |
| Per-slot persona picker in PhaseEditor UI | 0.7 | ✅ DONE | merged 2026-05-02 (PR #23 4ac8cdc); completes the per-phase persona-binding UX from PR #17 |
| Per-card cancel button + per-runner AbortController | 0.7 | ✅ DONE | merged 2026-05-02 (PR #24 58603b6); cancel a single voice mid-run without killing the chat |
| Surface opencode + kimi token usage in message_done | 0.7 | ✅ DONE | merged 2026-05-02 (PR #25 15b219a); per-lineage parser updates aggregating `step_finish` events session-wide |
| `cli_warning` banner + USD cost on participant card | 0.7 | ✅ DONE | merged 2026-05-03 (PR #26 3ff046c); compounds with #25 token aggregation — opencode multi-step cost summed independently of token parsing so a malformed-tokens-but-valid-cost step still bills correctly |
| OpenRouter inline flow (validate → fetch models → multi-add) | 0.7 | ✅ DONE | merged 2026-05-03 (PR #27 9934852); validate / catalog / insert flow + cockpit page. Per-pricing per-Mtok USD math; voice id format `openrouter:<model-id>`; `classifyOpencodeModel` reused for lineage classification. Self-review caught two follow-ups (defensive pagination + apiKey plumb-through) shipped in same PR. |
| HTTP shim for OpenRouter chat-completion dispatch | 0.7 | ✅ DONE | shipped 2026-05-03; new agent shim at `src/daemon/agents/openrouter.ts` POSTs `/api/v1/chat/completions` with `stream=true` + `stream_options.include_usage`, parses OpenAI-compatible SSE → AgentEvents (text_delta + message_done with native cost/tokens from OpenRouter), honours abort + 10-min timeout. Dispatch hook `pickShimForVoice(lineage, model)` routes any model with `openrouter:` prefix to the HTTP shim regardless of declared lineage; CLI precheck is skipped (auth = secrets table). Voices now insert with `enabled: true` since dispatch works. |
| Phase composition UI (drag/reorder, edit prompts) | 0.7 | 🟡 PARTIAL | per-slot persona binding shipped (PR #17 + #23). Drag-to-reorder, add/remove phases, fork-from-existing still PLANNED. |
| Default chorus-on-chorus template (Sentinel + Cartographer + Accountant + Translator) | 0.7 | 📐 PLANNED | bakes meta-fix |
| Squashed migration push to `chorus-codes/chorus` | 0.7 | ⏳ NEXT | piece-by-piece audit using personas. **Going live ~2026-05-04.** |
| `npm publish chorus-codes` | 0.7 | ⏳ NEXT | unscoped name (verified free). Rotate token after first publish. **Going live ~2026-05-04.** |
| Cleanup `99xAgency/chorus-ship-e2e` sandbox repo | 0.7 | ⏳ TODO | |
| Pre-audit cleanup sweep (12 findings fixed in one go) | 0.7 | ✅ DONE | 2026-05-01 — runner abort/done race, silent-empty doer, attached_files wire-up, builtin seed re-sync, version drift, brief wall, daemon logs, more. See "Pre-flagged" section. |
| **Per-slot fallback voice chain** (HIGH PRIORITY user-pain) | 0.8 | 🔥 NEXT | **Live pain:** when one voice in a multi-voice template errors (quota_exhausted, network, CLI crash), the partial work from other voices is wasted — they have to re-run from scratch alongside whatever voice replaces the failed one. Fix: each phase slot carries a `fallback: voice[]` chain. On retryable errors (`quota_exhausted`, `network`, `timeout`, `cli_failed`), the runner tries the next fallback voice for THAT slot only — other voices' completed work is preserved. Schema: `phases[].voices[].fallback: ['claude-code', 'openrouter:openai/gpt-4', ...]`. Reuses the same persona binding. Surfaces in PhaseEditor as a "+ add fallback" affordance under each voice slot. **Forces lineage-diversity rule into the chain** — fallback voices should default to same lineage as primary (so quorum math doesn't shift). Dovetails with multi-account-per-CLI: a fallback can be the same CLI on a different account when `CHORUS_CODEX_HOME` rotation is wired up. |
| `chorus audit <persona> <file>` CLI shorthand | 0.8 | 💭 IDEA | wraps `invoke_persona` MCP call; ~30 lines in `src/cli/index.ts`; saves typing the JSON in editors/conversations |
| Cockpit edit UI for builtin templates | 0.8 | ⏳ TODO | POST /templates upsert is now safe (preserves source=builtin); editor itself still missing — designed not built |
| Runner decoupling from SSE — background runChat + event bus replay | 0.8 | ⏳ TODO | surgical fix landed for v0.7 (no auto-abort + chat_done latch); proper fix is fire-on-POST so MCP flows don't sit drafting until a human opens the page |
| Home dashboard (CLI status, usage, reset windows, cost) | 0.8 | 📐 PLANNED | |
| Multi-account per CLI (add N codex / claude / gemini accounts of same vendor) | 0.8 | 📐 PLANNED | first-class UX for what `CHORUS_CODEX_HOME` env hack achieves today; auto-rotate to a non-rate-limited account when one hits quota_exhausted; surfaces in home dashboard alongside reset-window per-account |
| Run history + cost aggregates | 0.8 | 📐 PLANNED | |
| Public launch on npm + chorus-codes/chorus repo | 0.8 | ✅ DONE | shipped 2026-05-04 as `chorus-codes` v0.8.x; landing page live at chorus.codes |
| `chorus diagnose` command + crash hook for bug reports | 0.8 | ✅ DONE | merged 2026-05-08 (PR #1); copy-pasteable markdown bundle, version-mismatch flag, install-mode classifier, latest crash preview, 50-line log tails, SSE noise filter, realpath bin path |
| QUEUED placeholder cards for not-yet-spawned reviewers | 0.8 | ✅ DONE | merged 2026-05-08 (PR #2); 8-reviewer template now shows 8 cards from t=0 instead of revealing them as the daemon-wide CLI semaphore drains |
| Cross-slot fallback collision dedup | 0.8 | ✅ DONE | merged 2026-05-08 (PR #3); per-(chatId,round) in-flight registry stops two slots from independently falling back to the same model in parallel |
| Run-page reviewer card model name regression fix | 0.8 | ✅ DONE | merged 2026-05-08 (PR #6); chats.fromRow now derives `candidatesWithModels` from snapshot's `candidates` so card titles show the model name |
| Node 22 ERR_REQUIRE_ESM crash on `chorus init` | 0.8.29 | ✅ DONE | merged 2026-05-09 (PR #14, community contribution by Julien-Deudon); `open` and `chokidar` (both pure ESM packages) switched to dynamic `await import()` |
| macOS Keychain fallback for Claude Code v2 OAuth creds | 0.8.29 | ✅ DONE | merged 2026-05-09 (PR #8, community contribution by yuga-92); cli-precheck now probes `security find-generic-password -s "Claude Code-credentials"` before falsely reporting `auth_missing` |
| Codex headless isolation (`--ignore-user-config`) | 0.8.29 | ✅ DONE | shipped 2026-05-09; user MCP/plugin/hook configs were causing codex to hang or cancel mid-call when chorus spawned it as a reviewer (#10) AND when codex called chorus as MCP client (#16). Closes both. |
| Template `require ≤ candidates` validation | 0.8.29 | ✅ DONE | shipped 2026-05-09 (#15); zod `superRefine` rejects impossible quorum configs at template-save time instead of failing silently at run-start |
| `lsof` port-probe timeout during `chorus start` | 0.8.30 | ⏳ IN FLIGHT | community PR #13 (austinmao) in draft — 1s timeout on the synchronous lsof probe so a hung lsof can't block daemon startup |
| Auto-disable voice on persistent quota_exhausted | 0.8.30 | 📐 PLANNED | needs schema migration (`disabled_reason='auto_quota'`) + per-voice failure tracking. Pain: a Pro Gemini model on a Flash-only account fails forever; auto-disable stops repeated attempts (#11) |
| Local-LLM adapter (Ollama / LM Studio / vLLM via OpenAI-compatible base URL) | 0.9 | 📐 PLANNED | one shim covers every OpenAI-compatible local stack; reuses the OpenRouter HTTP shim's request/response handling |
| Multi-stage review (write → review → fix → re-review) | 0.9 | 📐 PLANNED | doer + reviewer + fix-doer + re-reviewer in a single template, iterating until quorum agrees or maxRounds hits |
| Per-voice persona overrides | 0.10 | 📐 PLANNED | tighter scope than the v0.7 per-slot persona binding — lets a saved voice carry a default persona that templates inherit unless overridden |
| CI integration (`chorus review --pr 1234`) | 1.0+ | 💭 IDEA | local-only review of an open GitHub PR diff |

Legend: ✅ done · ⏳ in flight · 🔥 high-priority next · 🟡 partial · 📐 designed · 💭 idea

> **Where we are (2026-05-09):** publicly launched on npm + chorus-codes/chorus 2026-05-04; ~120 unique installs by day 3 with ~30% activation. v0.8.x has been a reliability sprint driven by real user reports — diagnose tool, crash capture, queued cards, fallback dedup, node 22 ESM, Keychain creds, codex isolation, template validation. Three community PRs merged. v0.9 starts on multi-stage review + the local-LLM adapter.

---

## v0.6 — Public launch (DONE 2026-05-01)

Shipped:
- Audit-fix sweep — 7 real bugs caught dogfooding pre-publish (build asset copy, fresh-DB migration, MCP id↔chatId, CLI auto-detect, chorus start in dist, web spawn, postinstall hint)
- Cross-platform CLI detection — `where` on Windows + `which` on Unix, 8 fallback install dirs per OS, `--version` smoke test
- Manual path entry — when auto-detect misses, user pastes the path to the binary and chorus validates
- Reset → reinstall verified end-to-end via `npm pack` + `npm install -g <tarball>`
- Postinstall hint UX with state-aware help banner

---

## v0.7 — Voices · Personas · Phases (CURRENT)

The big one. Three independent dimensions (voice × persona × phase) that
users can compose freely.

### 1. Personas (DONE)

**Storage:** `personas` table seeded from `prompts/personas/*.md` on daemon startup. Built-in rows refresh from file source-of-truth on every boot; user-cloned rows (`builtin = 0`) are never overwritten.

**Built-in library** (full prompts in [`prompts/personas/`](prompts/personas/)):

| Persona | One-liner | Recommended lineage |
|---|---|---|
| **Sentinel** | Hunts secrets, injection, broken auth, supply-chain risk | anthropic |
| **Conservator** | Spots when a change fights the existing pattern instead of joining it | openai |
| **Cartographer** | Catches Windows/macOS/Linux assumptions, path separators, line endings, encoding | google |
| **Profiler** | N+1 queries, big-O cliffs, cold-path costs, latency budgets | openai |
| **Translator** | Reviews labels, errors, empty states, help text — for layman users, not engineers | anthropic |
| **Accountant** | Asks "who pays for this, when, and is there bill shock?" | anthropic |
| **Concierge** | Time-to-first-success — install path, error legibility, docs alignment | anthropic |
| **Quartermaster** | Scrutinizes every new dep — maintenance, license, transitive footprint, install scripts | openai |
| **Inspector** | Identifies what's not tested but should be — and what's tested but worthless | openai |
| **Librarian** | Reads README, marketing copy, help text alongside the diff and flags every lie | anthropic |

Each prompt is a *worldview*, not a checklist — single role, list of red flags to actively hunt, and an out-of-scope guard so personas don't bleed into each other's lanes.

**MCP surface** (live now):

- `list_personas()` → `{ personas: [{id, label, oneLiner, recommendedLineage, builtin}] }`
- `invoke_persona({personaId, brief, repoPath?, files?, template?})` → `{chatId, status, url}` — fires a real chat with the persona's system_prompt prepended to the user's brief.

> ✅ **Per-slot persona binding shipped 2026-05-02** (PR #17 969dca5 + PR #23 4ac8cdc). Templates now carry a `persona` field on each phase row, so a single template can route Cartographer → Discover, Sentinel → Develop, Accountant → Decide. The PhaseEditor UI exposes a per-slot persona picker. `invoke_persona` still works as the single-persona-overlay shorthand for ad-hoc audits. Drag-to-reorder + fork-from-existing remain on §5 (planned).

### 2. Voices (DONE 2026-05-02 — PR #2 c758e80)

Shipped: `voices` table with `(id, label, source, provider, model_id, lineage, vendor_family, input_cost_per_mtok, output_cost_per_mtok, enabled)`. Auto-populated on daemon boot — Phase 1 (sync, pre-listen) seeds single-model CLIs with immutable IDs (`claude-code`, `codex-cli`, …); Phase 2 (background, post-listen) shells `opencode models` for multi-model OpenCode voices. First-boot migration from `<lineage>.enabled_models` settings preserves prior toggles. Lineage stays in existing 5-enum; `vendor_family` carries the finer taxonomy (deepseek/meta/mistral/xai). Full GET/POST/PUT/DELETE routes; 6 UI surfaces (home fleet cards, /connect, onboarding, phase-editor) read from `/voices`. See [planning/voices.md](planning/voices.md).

### 3. Review-only mode (DONE 2026-05-02 — PR #4 b21a4b8)

A new phase kind, `review_only`, that takes the artifact as runtime input and skips doer spawn entirely. The unlock for `/work` and other harnesses to call chorus as a pure review service: `MCP create_chat({template: "review-only", artifact: <diff/draft/blob>})`.

Shipped:
- **Schema** — `PhaseSchema` is now a discriminated union by `kind`. `review_only` variant requires `reviewer` + `artifact { label, hint, maxBytes (default 1 MiB) }` instead of a doer block. `.refine` rejects hybrid templates (review_only mixed with standard) at parse time.
- **Runner** — `runReviewOnlyPhase` writes the artifact synthetically as the doer answer, emits synthetic `phase_start`/`phase_progress` events with `agent: 'artifact'`, then runs reviewers via the existing pool with `iterate.maxRounds = 1`. Single pass — no retry. Ship phase force-skipped (no doer diff to commit).
- **DB** — `chats.artifact TEXT` (nullable; capped by template) + `chats.verdict TEXT` (nullable; persisted from `chat_done` so list views can distinguish `verdict='request_changes'` from `verdict='approved'`). Both via idempotent `ADD COLUMN`.
- **Endpoint** — `POST /chats` validates artifact: required when first phase is `review_only`, rejected for non-review-only templates, capped at the template's `maxBytes`.
- **MCP** — `create_chat` gains optional `artifact` param.
- **Cockpit** — "Review only" pill in the template picker. Form swaps task ↔ artifact textarea (monospace, taller, hint placeholder, byte-cap badge). Doer indicator + repo-path block hidden. Run page hides the doer card and round indicators (single pass — no rounds to disambiguate).
- **`buildReviewerAsk` cap raised 2 KB → 256 KB** byte-aware (walks back UTF-8 continuation bytes so multi-byte runes aren't split into U+FFFD). Pre-existing legacy cap was silently truncating any review > ~50 lines; review-only mode just made it impossible to ignore.
- **Verdict surfacing** — `chat_done` now reports the real reviewer consensus (`approved` vs `request_changes`) instead of always `approved`. Standard templates untouched (verdict-tracking only kicks in when a review-only phase ran).
- **Built-in template** — `templates/review-only.yaml` (codex + gemini + claude reviewers, `require: 2`, `crossLineage: true`, `ship.enabled: false`).

Built using its own substrate as the dogfood gate — each commit was reviewed by the new template before the next built on it. Round-2 triage caught 4 real bugs (abort race, ship-skip enforcement, UTF-8 boundary, hybrid validation) all fixed in-branch. 18 new tests; 216/216 passing.

Out of scope (deferred):
- Multi-pass review-only with cockpit-driven revision loop. Today: revise yourself, resubmit a fresh chat.
- `chorus run` / `chorus review` CLI subcommands. Substrate ready; CLI wiring tracked separately in [`planning/cli-task-surface.md`](planning/cli-task-surface.md).

### 4. OpenRouter inline (NEXT — builds on voices)

**UX flow** on the OpenRouter row in onboarding/Settings:

```
┌─ OpenRouter ─────────────────────────────────┐
│ [sk-or-...........]  [Validate & list models] │
│ ✓ Validated. 327 models available.            │
│                                               │
│   Pick which to enable as voices:             │
│   ☐ Search models…                            │
│                                               │
│   Anthropic                                   │
│     ☐ claude-opus-4   $15/$75 per 1M          │
│   Moonshot                                    │
│     ☑ kimi-k2          $0.27/$1.10 per 1M     │
│   DeepSeek                                    │
│     ☑ deepseek-v3      $0.14/$0.28 per 1M     │
│   Meta                                        │
│     ☐ llama-3.1-405b   $0.80/$0.80 per 1M     │
│                                               │
│   2 selected. Save & add as voices.           │
└───────────────────────────────────────────────┘
```

**Mechanics:**

- New endpoint `POST /providers/openrouter/validate` body `{key}`:
  - Calls `GET https://openrouter.ai/api/v1/models` with `Authorization: Bearer <key>`
  - On 200 → returns model list (id, lineage, context, input_cost, output_cost)
  - On 401 → `{ok: false, error: "Invalid OpenRouter key"}`
  - On network fail → graceful timeout error
- Frontend caches the model list for the session (no re-fetch per re-render).
- User selects N models. On Save:
  - Upsert secret `openrouter` with the key
  - Insert N rows into `voices` table
- Voices table becomes the canonical source for the template designer's voice picker. CLI subs auto-populate the same table on first detect.

**Lineage tagging** — each OpenRouter model belongs to a lineage (anthropic / openai / google / meta / moonshot / deepseek / mistral / xai). Critical for chorus's diversity story.

**Cost surfacing** — per-1M-token costs shown in picker. Template designer later shows "this template will cost ~$X per run" (sum of expected tokens × per-model cost). Avoids the bill-shock failure mode.

### 5. Phase composition (PARTIAL — per-slot persona binding shipped 2026-05-02)

The §1 limitation is gone — PR #17 wired persona into doer + reviewer slots and PR #23 added the per-slot picker UI. Each phase row now owns its own `(voice, persona)` pair, so templates route Cartographer → Discover, Sentinel → Develop, Accountant → Decide out of the box.

Still PLANNED for full §5 (drag-reorder, add/remove phases, fork-from-existing):

```
phase {
  position: 0,
  phase_kind: "review",  // doer | review | synthesis | decide
  voice_id: "openrouter:moonshotai/kimi-k2",
  persona_id: "cartographer",
  prompt_overrides: { ... }   // overrides on top of persona system_prompt
}
```

UI: drag-to-reorder, add/remove phases, fork from existing template. Token inserts in prompts: `{{user_brief}}`, `{{prior_phase_outputs}}`, `{{repo_diff}}`, `{{file_path}}`.

### 5b. Round-2 deferred items (4-of-4 DONE 2026-05-02)

The four warmup items from [`planning/round-2-deferred.md`](planning/round-2-deferred.md) all shipped in a single autonomous session, each as its own PR + dogfooded via the new review-only template. Both Claude and Gemini reviewers caught real bugs in 3 of 4 dogfoods; round-1 fixes pushed before merge.

- ✅ **§2 Per-phase `timeoutMs` override** — PR #5 (9a47f23). Optional integer field on both `StandardPhase` and `ReviewOnlyPhase` schemas (30s..1h bounds) that overrides the default phase wait budget. Threaded through `runDoerHeadless`, `runReviewerHeadless`, both tmux paths' `waitForAnswer`, and the review-only synthetic StandardPhase. Round-1 dogfood flagged the magic-literal `5*60*1000` in tmux paths vs the named `DEFAULT_PHASE_TIMEOUT_MS` in headless — fixed by extracting `DEFAULT_TMUX_PHASE_TIMEOUT_MS` sibling constant with a comment explaining why tmux is tighter (file-watch polling vs subprocess wall).
- ✅ **§4 Opt-out telemetry heartbeat** — PR #6 (bac7445). Daemon-side ping to `chorus.codes` once on boot + every 24h. Fixed payload (`schema/installId/version/os/arch/node/daemonUptimeSeconds/chatsLast24h`) — no chat content, no PII, no file/repo paths. Three opt-out paths: `CHORUS_TELEMETRY=0|false|no|off`, `~/.chorus/no-telemetry`, settings flag. Round-1 caught a critical unhandled-rejection bug (`buildPayload` ran outside `sendHeartbeat`'s try/catch — a libsql disconnect during shutdown could crash the daemon); fixed by wrapping the whole body. Plus install-id mode tightened to 0o600 and timers `.unref()`'d so the daemon can exit naturally. Server endpoint at `chorus.codes/api/telemetry` + `/privacy` page deploy are separate infra tasks.
- ✅ **§3 Structured JSON-line logger substrate** — PR #7 (7ce6b46). Pino-shaped wire format (numeric `level`, `time`, `pid`, `hostname`) so a future swap to pino is transparent for downstream consumers. No runtime dependency added. Child loggers bake correlation fields (`chatId`/`phase`/`role`/`lineage`/`requestId`) into every emitted line. `CHORUS_LOG_LEVEL` env gate. Round-1 dogfood caught 3 real bugs: clobberable `ts`/`level` (user fields could overwrite core wire shape), `JSON.stringify` throws on circular refs/BigInt/hostile `toJSON` (logger should never throw — now wrapped with `safeStringify` and a degraded fallback), and `Error` instances serialized to `{}` (now expanded to `{message,name,stack,cause?}`). Five first-class call sites wired in `src/daemon/index.ts`; the mechanical migration of the remaining ~50 `console.log` sites is a follow-up.
- ✅ **§1 Token-capture rails** — PR #8 (3910607). Smallest viable substrate: `AgentEvent.message_done` extended with optional `usage: {inputTokens, outputTokens, cachedInputTokens}`, captured into `runDoerHeadless`'s return shape. Per-lineage parser updates (claude `result.usage`, codex `assistant_done`, gemini `*TokenCount`), DB persistence destination, cost computation from `voices.input_cost_per_mtok`, reviewer-side capture, and cockpit token display all deferred to focused follow-up PRs so each can land independently.

### 6. Default chorus-on-chorus template (PLANNED)

Hard-codes Sentinel + Cartographer + Accountant + Translator as the four reviewer slots in chorus's own audit template. The next person reviewing chorus *can't* skip those four lenses — bakes the meta-lessons into the product.

### 7. Migration to chorus-codes (NEXT)

After v0.7 ships:

1. Piece-by-piece audit via MCP — fire personas against critical files (`src/cli/index.ts`, `src/daemon/runner.ts`, etc.) using real chorus chats. The audit is the dogfood demo.
2. Findings → cleanup branch → squashed push to `chorus-codes/chorus` (single clean snapshot, no 99xAgency history).
3. Update `package.json` `name` → `chorus-codes` (unscoped, verified free on npm 2026-05-03; `@chorus-codes` org also reserved for future sister packages), repository + bugs URLs.
4. `npm publish` (unscoped, no `--access public` needed). Rotate token after first publish.
5. Cleanup `99xAgency/chorus-ship-e2e` sandbox repo.

---

## Pre-flagged for migration audit

Drift bugs / risks spotted in passing — most fixed 2026-05-01 in a single sweep before the migration audit even ran (the AUDIT prep itself surfaced 12 findings).

### Fixed 2026-05-01 (post-first-audit sweep)

The first cartographer audit attempt revealed 5 more findings — most caught while the doer was still running, then the system load spiked because the actual root cause (#17) was hammering the LLM CLIs. All 5 fixed before re-running the audit.

- ✅ **#13 Persona-composed brief dominated every chat title** — `mcp__chorus__invoke_persona` writes `# Persona: <Label>\n\n<system_prompt>\n---\n# User request\n\n<brief>` into `chat.work`. Sidebar/run page truncated by char count, so every Cartographer chat looked identical (200 chars of persona prompt). New helper at [`src/lib/chat-title.ts`](src/lib/chat-title.ts) `chatDisplayTitle()` parses the wrapper and returns `[<Persona>] <user request>`. Wired into sidebar, /runs list, and run page heading. Non-persona chats fall through unchanged.
- ✅ **#15 Phase stepper showed "DONE" while doer was still drafting** — the run-artifacts API ([`src/app/api/run-artifacts/[chatId]/route.ts`](src/app/api/run-artifacts/[chatId]/route.ts)) treated *file existence* as `hasAnswer`, but the runner pre-creates an empty `answer.md` at spawn time so live tail can poll. Now `hasAnswer` requires non-empty content (`answer.trim().length > 0`).
- ✅ **#16 Duplicate `phase_events` rows for the same phase** — observed 4× drafting + 10× blocked rows on a single chat. Root cause was the same as #17 (multiple concurrent runChats); fixing #17 eliminates it. No additional dedup needed in `phaseEvents.create` — the singleton invariant guarantees each event fires exactly once.
- ✅ **#17 (CRITICAL) — every SSE re-attach spawned a fresh runChat** — browser refresh, opening the run page in a second tab, MCP `wait_for_chat`, sidebar polling: each one called `GET /chats/:id/stream` which awaited a fresh `runChat()`. With code-review template that's 1 doer + 2 reviewers = 3 LLM CLI subprocesses *per SSE connection*, all writing to the same `answer.md` file, all consuming hundreds of MB. Hit load avg 133 with 8.5G swap full when 4-5 SSEs piled up on the same chat. Fix in [`src/daemon/index.ts`](src/daemon/index.ts): `Map<chatId, ActiveRun>` registry. First SSE wins → fires `runWithMultiplex` which runs `runChat` once and broadcasts events to a `Set<SubscriberFn>`. Subsequent SSEs subscribe to the existing run + replay past events from `phase_events`. Terminal-state chats just replay + close. POST `/chats/:id/cancel` and DELETE `/chats/:id` now also call `entry.abortController.abort()` so cancel kills the actual LLM CLI processes (was previously DB-only).
- ✅ **Daemon log verified end-to-end** — `[daemon] seeded 10 built-in personas` + `Chorus daemon listening on http://127.0.0.1:7707` now appear in `~/.chorus/logs/daemon.log` (vs previously /dev/null).

---

### Fixed 2026-05-01 (pre-audit sweep)

- ✅ **#1 `seedBuiltinTemplates()` doesn't update existing rows** — [`src/daemon/index.ts`](src/daemon/index.ts) `seedBuiltinTemplates`. Now mirrors the personas seed pattern: re-syncs builtin rows from disk on every daemon boot when YAML differs from DB. User-cloned rows (`source='user'`) are never overwritten.
- ✅ **#2 `POST /templates` silently demoted builtins to user-source** — [`src/daemon/index.ts`](src/daemon/index.ts) POST /templates handler. Now reads existing row first; if `source='builtin'`, the upsert preserves it. New rows still default to `source='user'`. Combined with #1, YAML source-of-truth is now authoritative for builtins.
- ✅ **#4 `attached_files` was dead state** — runner now reads chat.attached_files, packs each file (≤64 KB, ≤256 KB total) into a `## Attached files` block, and inlines it into both doer + every reviewer prompt. `invoke_persona({files: [...]})` finally works as documented.
- ✅ **#5 `readfile()` returns BLOB, Zod rejected** — DB layer at [`src/lib/db/index.ts`](src/lib/db/index.ts) `coerceTemplateYaml` now coerces Buffer→string at the read boundary. Direct SQL writes via `readfile()` no longer poison the daemon.
- ✅ **#6 Cockpit sidebar version was hardcoded** — [`src/components/app-sidebar.tsx`](src/components/app-sidebar.tsx) now fetches version from `/health` on mount and falls back to `—` when daemon is offline. No more drift between bumps.
- ✅ **#7 Run page rendered the entire brief as a wall of text** — [`src/components/live-run-real.tsx`](src/components/live-run-real.tsx) `BriefHeading` collapses briefs >200 chars to a one-line summary with "Show full brief" expander.
- ✅ **#8b Doer reported success on empty output** — [`src/daemon/runner.ts`](src/daemon/runner.ts) `runDoerHeadless`. Returns null when stream ended with no `message_done` AND no accumulated text, regardless of whether an error event fired. Catches CLI exits where stdout was unparseable, abort killed the process early, or the SDK ended silently.
- ✅ **#8c `phase_failed` events lacked `kind`, fell back to `plan`** — both phase_failed emission sites in [`src/daemon/runner.ts`](src/daemon/runner.ts) now include `phaseIdx`, `kind`, `role`. DB rows now reflect actual pipeline phase, not the schema fallback.
- ✅ **#9 `POST /chats` hardcoded `phase_kind: 'plan'`** — [`src/daemon/index.ts`](src/daemon/index.ts) now reads `template.phases[0].kind` for the initial drafting event. Falls back to 'plan' only on parse error. Verified live with a `code-review` chat creating `phase_kind: review` correctly.
- ✅ **#11 Runner only fired when SSE was opened, not on POST /chats** — combined with #12, was the proximate cause of the silent-approval bug. Mitigated by **#12 fix** below; full runner-decoupling deferred to v0.8 (see below).
- ✅ **#12 Abort/chat_done race overwrote `cancelled` with `approved`** — [`src/daemon/runner.ts`](src/daemon/runner.ts) `emitChatDone` is now a one-way latch: first emission wins. Plus [`src/daemon/index.ts`](src/daemon/index.ts) SSE handler no longer auto-aborts on connection close — closing a tab ≠ cancelling a chat. Explicit cancel still goes through `POST /chats/:id/cancel`. Plus a new `anyPhaseDoerFailed` flag terminates the chat as `failed` (not `approved`) when no doer ever produced real output.
- ✅ **Daemon stdio was `/dev/null`** — debugging silent failures was impossible. [`src/cli/index.ts`](src/cli/index.ts) `chorus start` now pipes daemon + cockpit stdio to `~/.chorus/logs/{daemon,web}.log`. Confirmed: `[daemon] seeded 10 built-in personas` now visible.

### Deferred — substantial work

- ⏳ **#3 Cockpit UI for editing builtin templates** — `src/app/templates/page.tsx` is still read-only. With #2 now safe (POST won't demote builtins), the missing piece is the editor itself. Designed but not implemented: in-page YAML editor with validate-on-save, "Fork to user template" button, voice/persona/quorum picker (depends on Voices abstraction in v0.7-final). Target: v0.8.
- ✅ **Runner decoupling from SSE (#11 properly)** — shipped 2026-05-02 in PR #21 (b372bec). `runChat` now fires in a background async task on POST /chats; SSE is a passive subscriber on a per-chat event bus that replays past events from `phase_events`. MCP-driven flows (`invoke_persona` from Claude Code/Codex/Cursor) no longer sit in `drafting` until a human opens the URL.

---

## v0.8 — Per-slot fallback voice chain (HIGH PRIORITY)

**The pain (live, observed 2026-05-03):** when one voice in a multi-voice template errors mid-run (codex `quota_exhausted` until next reset, gemini network blip, claude CLI crash, opencode subprocess kill), today the runner reports the slot as `errored` and the *other voices' completed reviews* are wasted — re-running the chat re-fires every voice from scratch. With a 4-voice review template that's 3 wasted ~$X reviews because of one bad subscription.

**The fix:** every phase slot owns a `fallback: voice[]` chain. On a retryable error class (`quota_exhausted | network | timeout | cli_failed`), the runner swaps in the next fallback voice **for that slot only**. The completed work from other voices is preserved — no global re-run.

**Schema** (additive on top of the per-slot persona binding from PR #17 + #23):

```yaml
phases:
  - kind: review_only
    reviewers:
      - voice: codex-cli           # primary
        persona: sentinel
        fallback:                  # tried in order on retryable error
          - openrouter:openai/gpt-5.5
          - claude-code
      - voice: gemini-cli
        persona: cartographer
        fallback:
          - openrouter:google/gemini-3-pro
      - voice: opencode-go         # no fallback declared — error stays errored
        persona: accountant
```

**Behaviour rules:**

- **Retryable errors only** — `cli_warning` and non-terminal events do NOT trigger fallback. Only the four error kinds above. Schema validation errors, hostile-stream writer-died, etc. stay terminal because the underlying problem is the prompt or environment, not the voice.
- **One fallback exhaustion at a time** — runner tries fallback[0]. If that also errors retryably, tries fallback[1]. Once the chain is empty, the slot is terminal `errored`. Quorum math then runs on whoever did succeed (e.g. `require: 2 of 3` still passes if 2 succeeded and 1 exhausted its chain).
- **Lineage diversity preserved by default** — PhaseEditor warns if a fallback differs in lineage from the primary (so quorum math doesn't silently shift from "2 lineages agree" to "all 3 are anthropic"). User can override with explicit acknowledgement.
- **No re-run of already-finished voices** — the partial event log is authoritative. The runner inspects `phase_events` for the slot, sees primary errored, dispatches the fallback to the same persona+brief without touching the others.
- **Cost attribution** — both primary attempt and fallback attempt log their own `cost_usd` rows. Run history shows "voice swap due to quota_exhausted" inline so the user understands why their bill has two charges for one slot.
- **MCP visibility** — `chat_done` payload gains `voiceSwaps: [{slot, from, to, reason}]` so MCP harnesses can decide whether the swap fundamentally changes the verdict semantics.

**UX surfaces:**

- PhaseEditor: each voice row gets a "+ add fallback" chevron. Drag-to-reorder within the chain. Lineage warning badge when chain crosses lineages.
- Run page: when a swap fires, the participant card shows a small "→ swapped to <fallback>" pill instead of disappearing. Both attempts visible in the round expander.
- Onboarding: when a user has only ONE voice in a lineage, the post-add nudge suggests "Add an OpenRouter fallback so a quota hit doesn't kill your runs."

**Dovetails with:**

- **Multi-account per CLI** (also v0.8): a fallback can be the same CLI on a different `CHORUS_CODEX_HOME` account. So Codex with quota A → Codex with quota B → OpenRouter `openai/*`. First-class UX once the `cli_account` table lands.
- **OpenRouter inline** (in flight PR #27): API-routed fallbacks become trivial once the dispatch shim ships. The roadmap-level value of OpenRouter is partly about being the universal-fallback layer.

**Effort:** ~3 days. Schema + parser + runner dispatch swap + PhaseEditor UI + 2 cockpit surfaces. Tests cover: retryable-classification, chain-exhaustion-then-quorum-still-passes, lineage-warning, cost-double-attribution, no-re-run-of-others.

---

## v0.8 — Home dashboard (PLANNED)

A live overview of your model fleet. Answers "what do I have, what am I burning, when do my limits reset."

**Scope:**
- **CLI status panel** — per CLI: installed/healthy, version, last-used, recent error count
- **Usage & reset windows** — for each subscription:
  - Tokens consumed this period
  - Period reset time (Claude Pro 5h windows, Codex daily caps, Gemini quotas)
  - Visualised as filling-bar with countdown
- **Cost tracking** — per-run, per-template, per-day:
  - CLI sub usage shown in "shadow $" (list-price equivalent for capacity planning)
  - API usage shown in real $
  - Aggregate dashboard with breakdown by voice, by template, by phase
- **Health pings** — heartbeat per CLI every N minutes, surfaces "claude is unhealthy" before a run fails
- **Multi-account per CLI** — power users hit weekly/daily plan limits and want to round-robin across multiple ChatGPT/Claude/Gemini accounts. Today the `CHORUS_CODEX_HOME` env override (PR #9) is a single-account stopgap. v0.8 brings: a `cli_account` table `(id, cli, label, home_path, enabled, priority)`, an Add Account UI on `/connect` that scaffolds `~/.codex-<label>/` and runs `codex login` in a tmux side-pane, automatic rotation when a voice's active account emits `quota_exhausted` (the in-flight PR #9 already surfaces this event), and per-account reset-window display in the dashboard above. Same shape extends to claude (`~/.claude-<label>/`) and gemini (`~/.gemini-<label>/`).
- **Run history** — last 50 chats with status, voices used, total cost, duration

**Storage:** new tables `voice_usage`, `cost_event`, `cli_health_ping`, `cli_account`.

**Effort:** ~1 week.

---

## v0.9 — Template sharing (PLANNED)

Foundations for shareable templates.

- Public `chorus.codes/templates` directory
- Author profiles, install count, ratings, tags
- One-click install: `chorus template install <slug>`
- Signing & verification (templates are config + prompts, but supply chain still matters)

---

## v1.0+ — Future (IDEAS)

- **Custom phase types** — users can define new phase kinds beyond review/doer/synthesis
- **Conditional phases** — "only run Sentinel if repo has API endpoints"
- **Multi-repo templates** — review changes across a monorepo or set of services
- **CI integration** — `chorus review --pr 1234` runs your template against any GitHub PR
- **Local LLM voices** — Ollama, llama.cpp as voices (offline / private)

---

## Website plan

`chorus.codes` needs a **Templates** section to sell the v0.7+ story:

1. **Hero update** — replace generic "multi-LLM peer review" with "Voices × Personas × Phases. Build your own. Or use one that wins."
2. **Templates page** (`/templates`) — gallery view:
   - Featured templates (curated): "React Refactor", "Solidity Audit", "API Design", "Copy Review", "Migration Plan"
   - Each card: voices used (lineage badges), personas, expected runtime, expected cost
3. **Personas page** (`/personas`) — the 10 built-in personas, what they hunt for, what they ignore. Sells the worldview-as-prompt idea.

---

## Storage schema (target — v0.7 final)

```sql
-- Voices: every routable model
CREATE TABLE voices (
  id TEXT PRIMARY KEY,
  label TEXT NOT NULL,
  source TEXT NOT NULL,                -- "cli" | "api"
  provider TEXT NOT NULL,
  model_id TEXT,
  lineage TEXT NOT NULL,
  input_cost_per_mtok REAL,
  output_cost_per_mtok REAL,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at INTEGER NOT NULL
);

-- Personas: the worldview a reviewer wears (LIVE)
CREATE TABLE personas (
  id TEXT PRIMARY KEY,
  label TEXT NOT NULL,
  one_liner TEXT NOT NULL,
  system_prompt TEXT NOT NULL,
  recommended_lineage TEXT,
  builtin INTEGER NOT NULL DEFAULT 0,
  forked_from TEXT REFERENCES personas(id),
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

-- Phases: per-template phase definitions (replaces hardcoded phases)
CREATE TABLE template_phases (
  id TEXT PRIMARY KEY,
  template_id TEXT NOT NULL REFERENCES templates(id),
  position INTEGER NOT NULL,
  phase_kind TEXT NOT NULL,            -- "doer" | "review" | "synthesis" | "decide"
  voice_id TEXT NOT NULL REFERENCES voices(id),
  persona_id TEXT NOT NULL REFERENCES personas(id),
  prompt_overrides TEXT,
  UNIQUE (template_id, position)
);
```

Migration of pre-v0.7 templates: convert each existing phase to a row with `persona_id = 'generic-reviewer'` (an 11th built-in for backwards compat), preserving voice assignments.

---

## MCP surface (target — v0.7 final)

```typescript
// LIVE
list_personas() → { personas: Persona[] }
invoke_persona({personaId, brief, repoPath?, files?, template?}) → ChatRef
create_chat({work, template?, files?, artifact?}) → ChatRef
   // artifact required when template's first phase is review_only

// PLANNED
list_voices() → { voices: Voice[] }
list_templates() → { templates: Template[] }   // already live
run_template({templateId, brief, repoPath?, voiceOverrides?}) → ChatRef
```

Use case: from inside Claude Code, user types *"Ask chorus to have Sentinel and Cartographer review this branch"* → MCP fires two `invoke_persona` calls in parallel → cockpit shows both reviews live → results flow back into the editor.
