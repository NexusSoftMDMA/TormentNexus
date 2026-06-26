## What's in 2.7.0

The biggest release since 2.6.0 — a six-week stretch covering **remote-context performance**, a **new project authoring flow**, **dashboard widgets**, **OAuth resilience**, and a top-to-bottom **performance instrumentation harness** that drove the bulk of the rest. 36 commits, no schema bump, no Hermes capability bump.

The throughline: Scarf got materially faster and more honest on slow remote SSH links, where 30-second sqlite timeouts and silently-empty UI used to be common. The skeleton-then-hydrate pattern, SSH cancellation propagation, and ScarfMon-driven diagnosis are the shape of how that work gets done now.

---

### Remote-context performance — chats and Activity in seconds, not 30s timeouts

Resuming a chat on a slow remote (a 420ms-RTT droplet, an underprovisioned VPS, a tunnel through 4G) used to fetch the full message column set in one shot, which routinely tripped the 30s SSH timeout on chats with multi-page tool result blobs. The 160-message session was broken; the 30-message session was broken too. Activity didn't load at all.

v2.7 introduces a **skeleton-then-hydrate pattern** that bounds the wire payload by what the user actually needs to see RIGHT NOW, then fills in the heavy stuff in the background:

- **Chat skeleton.** [`fetchSkeletonMessages`](https://github.com/awizemann/scarf/blob/main/scarf/Packages/ScarfCore/Sources/ScarfCore/Services/HermesDataService.swift) selects user + assistant rows only (skips `role='tool'`) with `tool_calls` / `reasoning` / `reasoning_content` hard-NULLed at the SQL level. Wire payload bounded by conversational text alone — typically a few KB. The chat appears in seconds. Background `startToolHydration` pages through `hydrateAssistantToolCalls` in 5-id batches to splice tool calls in. Tool-result CONTENT is **opt-in** via Settings → Display → "Load tool results in past chats" (default off); the inspector pane lazy-fetches per-result content via `fetchToolResult(callId:)` when you open a card.
- **Activity skeleton.** [`fetchRecentToolCallSkeleton`](https://github.com/awizemann/scarf/blob/main/scarf/Packages/ScarfCore/Sources/ScarfCore/Services/HermesDataService.swift) returns metadata-only rows (id + session_id + role + timestamp; everything else NULLed). Activity opens in <1s on remote with placeholder rows; real per-call entries swap in as paged hydration completes. New "Loading tool details…" pill in the page header surfaces hydration progress.
- **Single-id whale recovery.** When a 5-id batch trips the 30s timeout (one row carries an oversized `tool_calls` blob — a long Edit's args, a big diff), an L1 single-id retry isolates the offending row so the rest of the batch still hydrates. Whale row stays bare; assistant message stays readable.
- **Lazy tool result loading in the inspector.** Default-off avoids the bulk fetch. When you focus a tool call card, ChatInspectorPane fires `loadToolResultIfMissing(callId:)` which splices a single result into the message stream without re-fetching anything else.

Effect: a 160-message thinking-model session that used to time out at exactly 30s now opens in under 2 seconds with placeholder cards filling in over the next few. Activity loads in 500-800ms.

#### SSH cancellation that actually cancels

`Task.detached { … }` doesn't inherit cancellation from the awaiting parent, and `Task<…> { … }` (unstructured) also drops the signal. Without explicit bridging, cancelling a chat-load Task only unwinds Swift state — the underlying ssh subprocess kept running for the full 30s, pinning a remote sqlite query and a ControlMaster session slot. This produced the "third chat hangs" / "dashboard spins after rapid switching" symptom.

v2.7 wires `withTaskCancellationHandler` through [`SSHScriptRunner.run`](https://github.com/awizemann/scarf/blob/main/scarf/Packages/ScarfCore/Sources/ScarfCore/Transport/SSHScriptRunner.swift) and [`RemoteSQLiteBackend.query`](https://github.com/awizemann/scarf/blob/main/scarf/Packages/ScarfCore/Sources/ScarfCore/Services/Backends/RemoteSQLiteBackend.swift) so parent cancellation reaches the `Process` and calls `proc.terminate()` within 100ms. New `ssh.cancelled` ScarfMon event surfaces this.

#### In-flight coalescing for `loadRecentSessions`

File-watcher deltas during an active stream used to stack 2-3 parallel sessions-list reload tasks (the 500ms `scheduleSessionsRefresh` debounce only suppresses a pending tick, not one already executing). Subsequent callers now await the in-flight load instead of spawning a parallel SSH subprocess. New `mac.loadRecentSessions.coalesced` event tracks dedup hits.

#### Loading-state UX hardening

The Mac chat sidebar greys out and disables row taps the moment a session-switch is initiated (synchronously, before `client.start()` returns), with a floating ProgressView showing the current phase: **"Spawning hermes acp…"** → **"Authenticating…"** → **"Loading session…"** → **"Loading history…"** → **"Ready"**. Pre-fix the sidebar looked engageable while the 5-7 second SSH+ACP boot was still in flight, and the user could queue up a second session-switch behind the first. New `isStartingSession` flag flips on user click for instant feedback.

#### Partial-result + mismatch + pinned-model banners

- **Partial-result banner.** When the skeleton fetch trips an SSH transport failure (rather than a clean empty result), the chat surfaces "Couldn't load full chat history — the connection to *server* timed out" through the existing `acpError` triplet, plus forces `hasMoreHistory = true` so the "Load earlier" affordance shows up. Replaces the pre-fix silent empty transcript.
- **Model/provider mismatch banner.** [`ModelPreflight.detectMismatch`](https://github.com/awizemann/scarf/blob/main/scarf/Packages/ScarfCore/Sources/ScarfCore/Services/ModelPreflight.swift) recognizes when `model.default` carries a `<provider>/...` prefix that disagrees with `model.provider` (e.g. `anthropic/claude-sonnet-4.6` + `provider: nous` after switching OAuth via Credential Pools). Banner offers one-click fix in either direction.
- **Pinned-model failure hint.** ACP error classifier now recognizes `model_not_found` / `404 messages` / `model is not available` and surfaces "This session was created with a model the provider no longer offers — start a new chat to use your current model" so the pinned-model failure mode has a clear recovery path.
- **OAuth-completion provider swap.** After a successful OAuth in Credential Pools, if the just-authed provider differs from `model.provider`, surface "Switch active provider to *name*?" with [Switch] / [Keep current] instead of auto-dismissing.

---

### New Project from Scratch wizard + Keychain-backed cron secrets

A **third project entry point** alongside Browse Catalog and Add Existing Project: a wizard that scaffolds a Scarf-standard project skeleton (`<project>/.scarf/dashboard.json` + AGENTS.md marker block), registers it, and hands off to a chat session that auto-activates the bundled `scarf-template-author` skill. The skill drives the rest conversationally — widgets, optional config schema, optional cron — and writes the final files itself. Wizard stays minimal because the agent does configuration better than a multi-step form. The skill ships bundled inside `Scarf.app/Contents/Resources/BuiltinSkills.bundle/` and copies into `~/.hermes/skills/` on launch (idempotent + version-gated).

**Cron + Keychain — `$SCARF_<SLUG>_<FIELD>` env vars.** Cron prompts that referenced `secret`-typed config fields used to get the literal `keychain://...` URI back when reading `config.json`, producing 401s. v2.7 mirrors resolved Keychain values into `~/.hermes/.env` under a marker-bounded block keyed by template slug:

```sh
# scarf-secrets:begin local-news-aggregator
SCARF_LOCAL_NEWS_AGGREGATOR_API_TOKEN=actual-value
SCARF_LOCAL_NEWS_AGGREGATOR_RSS_URL=https://example.com/feed
# scarf-secrets:end local-news-aggregator
```

Hermes already reloads `~/.hermes/.env` per cron tick, so credential rotation is automatic — just edit the value in Configuration → next tick sees it. The mirror runs at every state-change point: install, post-install Configuration save, uninstall, "Remove from List", and on app launch (reconciliation pass over registered projects). Source of truth stays in the Keychain — `config.json` keeps `keychain://` URIs unchanged. Mode 0600 enforced on `~/.hermes/.env`.

Cron prompts now reference these env vars directly:

```json
{
  "prompt": "Use the terminal: curl -sS -H \"Authorization: Bearer $SCARF_LOCAL_NEWS_AGGREGATOR_API_TOKEN\" \"$SCARF_LOCAL_NEWS_AGGREGATOR_RSS_URL\" -o {{PROJECT_DIR}}/.scarf/feed.xml"
}
```

**Migration.** First launch of v2.7 walks the project registry and writes the managed block per schemaful project — automatic. Existing cron prompts you wrote against the old (broken) `config.json` pattern still need updating: open the cron job in Scarf's Cron sidebar and edit the prompt, or ask the agent in chat ("Update my Local News cron job's prompt to use the new env var convention") — the bundled `scarf-template-author` skill (now v1.1.0) documents the convention with worked examples.

Also fixes [#75](https://github.com/awizemann/scarf/issues/75) — `_NSDetectedLayoutRecursion` on the Configuration form for projects whose form transitioned between stages with different intrinsic heights.

---

### Project dashboards — file-reading widgets, sparklines, typed status

Five new widget types, project-wide auto-refresh, and a structured error card for unknown widgets. Backwards-compatible — every existing `dashboard.json` renders byte-identically.

- **Project-wide auto-refresh.** [`HermesFileWatcher`](https://github.com/awizemann/scarf/blob/main/scarf/scarf/Core/Services/HermesFileWatcher.swift) used to watch each project's `dashboard.json` specifically. v2.7 promotes that to a watch on the entire `<project>/.scarf/` directory. A `markdown_file` or `log_tail` widget pointing at `<project>/.scarf/reports/foo.md` refreshes the moment a cron job rewrites the file. **By convention, place files the dashboard reads inside `.scarf/`** so the watch picks them up.
- **`markdown_file`** — renders a markdown file from disk through the same `MarkdownContentView` pipeline used by inline `text` widgets.
- **`log_tail`** — last `lines` of a file (default 20, max 200), monospaced, ANSI codes stripped.
- **`cron_status`** — last run / next run / state for one Hermes cron job by `jobId`, plus a small inline log tail. Read-only — Run/Pause/Resume controls stay on the Cron tab.
- **`image`** — local file (`path` relative to project root) or remote `url`. Optional `height` cap. Useful for matplotlib/Plotly PNGs the cron job generates.
- **`status_grid`** — compact NxM grid of colored cells, one per service / item, with hover labels.
- **`stat` widget gains inline sparklines.** Optional `sparkline: [Number]` field. SVG-only render, dozens per dashboard cost nothing.
- **Typed status badges.** `list` items and `status_grid` cells share a typed enum (`success`, `warning`, `danger`, `info`, `pending`, `done`, `neutral`) with lenient decode for synonyms (`ok`/`up` → success, `down`/`error` → danger). Unknown strings render as plain text.
- **Structured widget error card.** Replaces the legacy "Unknown: \<type\>" placeholder with a card surfacing the title, specific reason, and a hint.
- **Schema mirror.** The widget vocabulary lives once at [`tools/widget-schema.json`](https://github.com/awizemann/scarf/blob/main/tools/widget-schema.json); the catalog validator reads from it and enforces per-type required fields.

---

### OAuth resilience + Credential Pools

- **Daily OAuth keepalive cron.** Prevents Anthropic OAuth refresh tokens from expiring after weeks of inactivity. New cron job `[scarf:oauth-keepalive]` (managed by Scarf) pings Hermes on a daily cadence; the in-app Refresh All Sessions action mirrors the same path on demand.
- **Remote re-auth.** Re-authenticating against a remote droplet's OAuth provider used to be blocked by the lack of a stdin path through SSHTransport. The OAuth flow now drives a remote `hermes auth add` correctly with stdin forwarded.
- **OAuth remove button.** Per-provider remove action in Credential Pools (auth.json edit), with confirmation dialog. Companion auto-refresh of the view when `auth.json` changes externally (file-watcher).
- **`resolve_provider_client` error classification.** When an auxiliary task references a provider whose credentials aren't loaded, Hermes prints `resolve_provider_client: <name> requested but <Display Name> not configured` to stderr — pre-fix this surfaced in chat as the opaque `-32603 Internal error` with no actionable detail. Now classified into a clear hint pointing at Settings → Aux Models.
- **Aux Tab unknown-task surface.** When `config.yaml` has an `auxiliary.<task>` block for a task Scarf doesn't know about (newer Hermes added it; Scarf hasn't caught up), render it as a plain row with the raw provider/model values instead of dropping it silently.
- **Credential Pools refresh after OAuth sheet dismiss.** Closing the OAuth sheet after a successful add now refreshes the list immediately instead of leaving the just-added pool hidden until the next file-watcher tick.

---

### ScarfMon — performance instrumentation harness

The diagnostic surface that drove the bulk of the v2.7 perf work. Off by default; signpost-only mode (Instruments-friendly) is free; Full mode (4096-entry in-memory ring buffer + os.Logger) is a click away in Settings → Diagnostics → Performance. Wiki: https://github.com/awizemann/scarf/wiki/Performance-Monitoring

- **Phases 1-3** built the core: dispatcher + ring buffer + 3 backends, chat / transport / sqlite measure points, diagnostic counters for chat-render bursts, finalize-burst dampening.
- **Tier A + B** added per-feature instrumentation: iOS file watcher, sessions list, model catalog, dashboard widgets, image encoder, message hydration.
- **Nous picker investigation** localized a 60s + 120s beach-ball to a specific path (Nous catalog `readCache`), then killed the 120s one with dedupe + 5s timeout.
- **Tier C catch-up** (this release): instrumented Memory / Skills / Cron / Curator load paths so future captures show how often these tabs cost multiple sequential SFTP RTTs on remote.
- **Per-call bytes recorded** on transport + sqlite events so captures show payload sizes alongside latencies.
- **`mac.emptyAssistantTurn` event** documents the Nous quirk where the model returns a thought stream with no body (the bubble looks like Hermes is "still thinking" but the turn already finished).

Adding a new measure point is two lines. The harness covers Mac and iOS uniformly. The "Copy as JSON" button exports the ring buffer for paste-into-issue diagnosis.

---

### Other fixes + polish

- **Sessions sidebar reload debounce** — file-watcher deltas during streaming used to flicker the sessions list. Coalesced into one trailing fetch ~500ms after the last tick.
- **Session-load pagination + race guard** — switching to a small chat while a larger one is mid-fetch could last-write-wins the small chat away. Three race-checks against `self.sessionId` prevent the stale fetch from overwriting.
- **Sessions + previews batched** — two separate SSH calls folded into one `queryBatch` round trip, halving the round-trips for every sidebar refresh.
- **Remote SQLite query timeout** bumped 15→30s to better tolerate slow links; in-flight query coalescing dedupes concurrent identical queries.
- **`Thread.sleep` spin replaced** with a kernel-wait via `DispatchGroup` for `runLocal` timeout; under concurrent SSH load the old loop accumulated spin-blocked threads and produced 7-second outliers in `loadRecentSessions`.
- **Window position + size** persists across launches.
- **Sidebar reorder** — Projects promoted to first section; profile chip moved under server name.
- **`stop` badge suppressed** on metadata footer for normal turn ends (it was firing for every clean completion, looking like an error).
- **Nous picker search field** + `model-picker` filter for the long Nous overlay model list.
- **`oauth-keepalive` cron create** — drop the `--silent` flag Hermes doesn't accept.
- **Snapshot pipeline rewritten** — replaced the `sqlite3 .backup`-then-download pipeline with direct SSH-streamed query execution (issue [#74](https://github.com/awizemann/scarf/issues/74)). Eliminates the multi-minute snapshot wait on multi-GB state.db files. Companion fix: pre-expand `~/` in Swift via `resolvedUserHome` so sqlite3 finds the DB without depending on the remote shell's tilde expansion.
- **Aux nested-YAML parser** — corrected the parser so the unknown-task surface works on remote (was previously dropping aux blocks whose `provider:` value lived on a separate line).
- **`ModelPreflight` newline trim bug** — `.whitespaces` doesn't strip newlines; switched both trims to `.whitespacesAndNewlines` so a stray `\n` in a hand-edited config.yaml doesn't false-positive the mismatch banner.

---

### What's measured today

321 ScarfCore tests pass (302 prior + 19 new ModelPreflight). New ScarfMon events documented in the [Performance-Monitoring wiki](https://github.com/awizemann/scarf/wiki/Performance-Monitoring).

### Compatibility

- macOS 14+ (unchanged).
- Hermes target: still **v2026.4.30 (v0.12.0)**. No new Hermes capability gates added.
- Existing `dashboard.json` files render unchanged.
- Existing `.scarftemplate` bundles install unchanged. Catalog manifest schemaVersion stays at 1/2/3 — no bump.
- Existing `~/.hermes/.env` content is preserved byte-identically — Scarf only writes inside its `# scarf-secrets:begin <slug>` / `# scarf-secrets:end <slug>` regions.
- The skeleton-then-hydrate chat loader and SSH cancellation propagation are **Mac-only** in this release; ScarfGo (iOS) keeps its existing chat path.

### What's deferred

- **Per-widget data sources + per-widget refresh granularity.** The general "widget points at a typed data source" abstraction is the next-largest win in dashboards but materially expands the model + JS mirror + validator surface. The project-wide watch covers the common cron-driven workflow without it.
- **Cross-project health digest sidebar rollup.** Counting attention-needed projects across the registry — scoped but didn't pull its weight. The typed status enum makes it cheap to add later.
- **Automatic cron-prompt rewriter on upgrade.** Heuristic rewrites of free-form prompts are risky; the docs + agent-assisted path ships in v2.7. Revisit a "scan + fix" UI in v2.8 if real users miss the migration.
- **iOS New Project wizard + iOS Keychain-env mirror.** ScarfGo's project surface is read-only; the wizard's chat-handoff pattern depends on Mac-only ACP plumbing.
- **iOS skeleton-then-hydrate loaders.** Same data-service surfaces are public, but the iOS chat lifecycle is structured differently. Defer until iOS dogfooding shows the same payload-size pain.
- **Tier C redesigns (Memory/Skills/Cron/Curator).** Instrumented in v2.7; redesign waits for capture data showing which path actually needs the skeleton-then-hydrate treatment.
