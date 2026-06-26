## What's in 2.8.0

A coordinated catch-up release adopting Hermes v0.13.0 (v2026.5.7) — "The Tenacity Release" — across Scarf's full surface area. v2.8.0 ships **Persistent Goals**, **ACP `/queue`**, **Kanban diagnostics + recovery UX**, **Curator archive/prune**, **Google Chat (20th platform) + cross-platform allowlists**, a refreshed **provider catalog** with five new models, and a slate of settings + UX polish — all behind capability flags so pre-v0.13 hosts continue to render the v2.7.5 surface unchanged.

No data migrations, no schema changes. `~/.hermes/state.db` columns are unchanged from v0.11/v0.12. Existing `~/.hermes/scarf/` sidecars are untouched. Sparkle picks the update up automatically.

### New features

#### Persistent Goals + ACP `/queue` (chat)

- **`/goal <text>` slash command** — locks the agent on a target that persists across turns. Surfaced via the chat slash menu (gated on `HermesCapabilities.hasGoals`) and rendered as an `info`-tinted "Goal locked: …" pill in the chat header. The pill exposes a "Clear goal" context-menu item that dispatches `/goal --clear`. Optimistic local mirror — Hermes is the authoritative owner; Scarf paints the pill the moment the user sends `/goal …` so the affordance feels instant.
- **`/queue <text>` slash command** — queues a prompt to run after the current turn completes. Joins `/steer` and `/goal` in `RichChatViewModel.nonInterruptiveCommands` (the chat keeps "Agent working…" off when sent). A header chip shows the queued count; tap opens a popover listing prompts + relative timestamps. Per-entry deletion isn't exposed (Hermes has no remove-by-id verb), and the popover header makes that explicit so users understand the local mirror's role.
- **`/steer` on idle** — pre-v0.13 was a no-op when no turn was in flight; v0.13 runs it as a regular prompt. The composer's slash button now greys `/steer` only on pre-v0.13 hosts (gated on `hasACPSteerOnIdle`).
- **Static slash-menu fallbacks** — pre-session, the menu surfaces `/new` (with optional `[<name>]` argument hint on v0.13). Active-session-only fallbacks (`/clear`, `/compact`, `/cost`, `/model`, `/tools`, `/reload-skills`, `/help`, `/exit`) round out resumed sessions where Hermes ACP doesn't re-emit `available_commands_update` after `session/load`. Deduped against the ACP-advertised set so the canonical entry always wins once a session opens.

#### Kanban v0.13 diagnostics + recovery UX

- **Hallucination-gate verify / reject** — worker-created cards land with `hallucination_gate_status: pending`. The inspector renders a yellow banner ("Created by a worker — verify before running") with a Verify and Reject button. Cards in pending state dim 0.6 with a yellow ⚠ glyph in the title row.
- **Diagnostics rendering** — new typed-mirror enum `KanbanDiagnosticKind` with severity (info / warning / critical). Per-task and per-run diagnostics surface in the inspector Runs tab as chip-lists; auto-block reasons render verbatim in the existing red banner. Darwin zombie detections show as a distinct `darwin_zombie_detected` kind.
- **Per-task `max_retries`** — added to the create sheet (default 3) and shown as a header chip in the inspector. Write-once at create time, matching Hermes's pattern.
- **Multiline title/body** — the create sheet's Title field accepts multiline input, capped to four visible rows.
- **Tolerant decoding** — every new field uses `decodeIfPresent`. Pre-v0.13 JSON parses cleanly with the new fields defaulting to nil, and the v2.7.5 board surface is unchanged on older hosts.

#### Curator archive + prune

- **Archived skills section** in `CuratorView` showing `hermes curator list-archived` output. Each row exposes Restore (returns to the active leaderboard) and Prune (destructive — opens a custom confirm sheet matching the template-uninstall pattern, with `ScarfDestructiveButton` "Prune permanently" and Cancel as the default keyboard action).
- **Bulk prune** — a header action (gated on archived list non-empty) that enumerates every archived skill in the confirm sheet before a single-tap destructive action. Per-skill prune buttons are present per row when Hermes supports `prune <name>`; otherwise only the bulk action is exposed.
- **Synchronous "Run Now"** — v0.13 `hermes curator run` blocks until done. The Run Now button shows a progress affordance for the duration; pre-v0.13 falls back to fire-and-forget.
- **New `CuratorService` actor** in ScarfCore ([scarf/Packages/ScarfCore/Sources/ScarfCore/Services/CuratorService.swift](scarf/Packages/ScarfCore/Sources/ScarfCore/Services/CuratorService.swift)) — pure-I/O Sendable actor mirroring `KanbanService`'s shape, with defensive `--json` retry-without-flag fallback for verbs that may not support it on all v0.13 patch releases.
- The legacy `CuratorRestoreSheet` flow (SAFE-list-restore for v0.12) is preserved; it predates the v0.13 archive surface and serves a distinct case.

#### Messaging Gateway expansion

- **Google Chat** — 20th platform. New entry in the Mac Platforms tab, gated on `HermesCapabilities.hasGoogleChatPlatform`.
- **Cross-platform allowlists** — per-platform editor for `allowed_channels` (Slack / Mattermost / Google Chat), `allowed_chats` (Telegram / WhatsApp), and `allowed_rooms` (Matrix / DingTalk). New `AllowlistEditor` component plus the `GatewayAllowlistKind` / `GatewayPlatformSettings` ScarfCore types. Persisted to `~/.hermes/config.yaml` via a new `GatewayConfigWriter` since `hermes config set` doesn't write list blocks.
- **Per-platform behavior toggles** — `busy_ack_enabled` (suppress per-message "agent is working…" acks), `gateway_restart_notification` (post a "Gateway restarted" notice on boot), and a slash-command auto-delete TTL (seconds, 0 to disable). Each appears in the new `GatewayBehaviorSection` component.
- **`hermes gateway list` cross-profile digest** — inline status row in `MessagingGatewayView` showing which profile is running which platform across all profiles. New `HermesGatewayListService` actor parses `hermes gateway list --json`. Hidden when the verb fails (pre-v0.13 hosts) or no profiles are registered.
- **`MessagingGatewayViewModel`** — internal rename from `GatewayViewModel` to disambiguate from the v0.10 Tool Gateway feature. The user-facing label was already "Messaging Gateway" since v0.10.
- **`[[as_document]]` hint** — informational tooltip in skill detail surfaces explaining the new media-routing directive for skills that reference it.

#### Provider catalog refresh

- **Five new models** — `deepseek/deepseek-v4-pro`, `x-ai/grok-4.3`, `openrouter/owl-alpha` (free tier), `tencent/hy3-preview`, and `arcee/trinity-large-thinking` (with temperature + compression overrides). Surfaced through `models_dev_cache.json`; no manual entries required.
- **Grok rename** — `x-ai/grok-4.20-beta` → `x-ai/grok-4.20`. Implemented via read-time alias resolution in `ModelCatalogService.modelAliases` so existing user configs with the `-beta` suffix keep validating without YAML rewrites. Three composite-keyed aliases cover the openrouter / xai / vercel routes.
- **Vercel AI Gateway demoted** — sort comparator change in `loadProviders()` puts Vercel last, after the alphabetical group.
- **`image_gen.model` honored** — pre-v0.13 the key was advertised but ignored; v0.13 actually drives the image-generation path. Surfaced in `Settings → Auxiliary` with a curated picker (`OpenAI gpt-image-1`, `Imagen 3/4`, `Stable Image Ultra`, `FLUX 1.1 Pro`, `DALL·E 3`); free-form entry is also accepted. Gated on `hasImageGenModel`.
- **OpenRouter response caching** — toggle in `Settings → Auxiliary` writing `openrouter.response_cache.enabled` to `config.yaml`. Off by default in Scarf's parser. Gated on `hasOpenRouterResponseCache`.

#### Settings tab additions

- **MCP SSE transport** — MCP add-server flow gains a Transport picker (`stdio` / `http` / `sse`) with `sse_read_timeout` field for SSE servers. The YAML round-trip preserves OAuth + headers identically to the existing `.http` shape. Gated on `hasMCPSSETransport`.
- **Cron `--no-agent` watchdog mode** — toggle in the Cron edit sheet that maps to `hermes cron create/update --no-agent`. When ON, the prompt + context fields hide (the AI call is skipped). Defensive write-path strips the flag on pre-v0.13 hosts mirroring the `--workdir` pattern. New `HermesCronJob.noAgent: Bool` field with `decodeIfPresent` so pre-v0.13 reads keep parsing. Gated on `hasCronNoAgent`.
- **Web Tools per-capability backends** — new `Settings → Web Tools` tab with separate pickers for `web_search` and `web_extract`. SearXNG appears in the search picker only. The legacy single `web_tools.backend` is still readable for round-trip safety on mixed-version installs. Gated on `hasWebToolsBackendSplit`.
- **Profiles `--no-skills`** — "Empty profile (no skills)" toggle in the create-profile flow that appends `--no-skills` to `hermes profile create`. Disabled when "Clone all" is on (mutually exclusive). Gated on `hasProfileNoSkills`.

#### UX polish

- **Context compression count** in the chat status bar. v0.13 emits the count alongside the token tally on the `session/prompt` response; Scarf renders a `🗜 ×N` chip next to the token count when `count > 0`. Gated on `hasContextCompressionCount`.
- **`/new <name>` argument hint** — bracket-aware so v0.13 hosts show `[<name>]` and pre-v0.13 hosts show no hint.
- **`HermesUpdaterCommandBuilder`** — forward-compat plumbing for `hermes update --yes`. No in-app surface in v2.8.0 (Scarf doesn't currently expose a "Run hermes update" command); the builder is wired so a future Settings affordance can opt in cleanly.
- **Redaction default-flip awareness** — the existing `Settings → Advanced → Redaction` toggle hint copy now branches on `HermesCapabilities.isV013OrLater`. v0.13+ hosts read "Recommended: ON. Hermes v0.13 defaults to redacting secrets unless you opt out"; pre-v0.13 keeps the v2.7 hint.
- **`display.language` picker** — new `Settings → General → Locale` row. 8 options: default, zh, ja, de, es, fr, uk, tr. Hermes does the actual translation; Scarf just persists `display.language` to `config.yaml`. Gated on `hasDisplayLanguage`.
- **xAI Custom Voices badge** — `Settings → Voice` shows a "Cloning supported" `ScarfBadge` next to the xAI TTS provider entry. Informational only; voice management itself happens via `hermes voice` CLI. Gated on `hasXAIVoiceCloning`.

#### ScarfGo iOS catch-up (read-only)

Following the Phase H precedent, iOS mirrors selected v2.8 surfaces as read-only — write parity is deferred to v2.8.x.

- **Goal pill + queue chip** in the iOS chat header (`projectContextBar`). Tap is a no-op; the Mac app owns mutations.
- **Kanban v0.13 diagnostics** in `ScarfGoKanbanDetailSheet` — `retries: N` chip, "Worker-created — verify on Mac" hallucination badge, red `auto_blocked_reason` banner, tappable diagnostics chip-lists with severity-tinted badges and a new `DiagnosticDetailSheet` (replacing Mac's `.help()` tooltip on touch).
- **Curator Archived list** in `Scarf iOS/Curator/CuratorView.swift` — read-only, with footer pointing users to the Mac app for Restore / Prune actions.
- **Settings → Platforms extension** — Google Chat status row, busy-ack and restart-notification summary rows across `gatewayPlatforms` (handles disagreement with "mixed (N platforms)"), allowlist DisclosureGroups with monospaced "platform: id" entries when expanded.
- **"v0.13 features active" badge** in iOS Settings (gated on `caps.isV013OrLater`). Tap presents `V013FeaturesSheet` listing the new affordances.

### Capability gating

v2.8.0 adds 22 new flags on `HermesCapabilities` (each gating one v0.13 surface), plus an `isV013OrLater` convenience predicate. Every new affordance is gated; pre-v0.13 hosts see the v2.7.5 surface byte-identical to before. The HermesVersionBanner threshold remains pre-v0.12 — v0.12 → v0.13 nudging happens via the iOS Settings badge (positive surface) rather than a global yellow banner (which was reserved for "missing every new feature" cases).

### Bug fixes uncovered during v0.13.0 dogfooding

- **Dashboard flicker on v0.13 hosts** — Hermes v0.13 writes to `state.db-wal` and rotating logs at ~10 Hz during gateway activity. Each FSEvents fire ticked `lastChangeDate`, every observing view re-fired its load handler against it, and on Local hosts the dashboard stacked 5+ concurrent `dashboardSnapshot` calls in 200 ms — sqlite contention on the read-only handle surfaced as `BackendError error 3`, plus visible flicker. Two-part fix: `HermesFileWatcher.scheduleCoalescedTick` coalesces FSEvents into one observable mutation per 500 ms quiet window with a 1.5 s max-wait floor (so a coincident `gateway_state.json` Start/Stop touch can't be starved indefinitely under sustained WAL writes); `DashboardViewModel.load()` holds a single in-flight `Task<Void, Never>` handle so concurrent triggers await the in-flight load instead of stacking.
- **Sparse slash menu on resumed sessions** — Hermes ACP only emits `available_commands_update` after `session/new`, not after `session/load`. Combined with `RichChatViewModel.reset()` clearing `acpCommands` on every session switch, resumed sessions landed at a 4-command fallback even though the agent identity hadn't changed. Fix: stop wiping `acpCommands` in `reset()` (they're agent-level, not session-level), and add an active-session-only static fallback set covering the standard agent commands so cold-start LOAD users see a rich menu immediately.

### Migrating from 2.7.5

Sparkle delivers the update automatically. No config migration, no schema changes — same `~/.hermes/state.db` columns as v0.11/v0.12, same Scarf-owned sidecars at `~/.hermes/scarf/`. Existing v2.7.5 Kanban tenants stay valid; existing project manifests are unchanged. Settings tabs grow new rows; existing rows render identically.

If you're connecting to a Hermes v0.13.0 host for the first time after this update, the new surfaces light up automatically — no flag flip in the app. Pre-v0.13 hosts continue to render the v2.7.5 surface; nothing breaks if you upgrade Scarf before upgrading Hermes.

### Known limitations

- **iOS write surfaces** (Verify hallucination gate, Reject, Curator archive/prune actions, allowlist editor, `/goal` send, `/queue` send) are explicitly out of scope for v2.8.0 and slated for v2.8.x. iOS surfaces are read-only mirrors per the Phase H precedent.
- **Auto-resumed-from-checkpoint indicator** — Hermes v0.13's "auto-resume after gateway restart" feature is server-side; whether the ACP adapter advertises a Scarf-visible signal is unclear pending live host verification. Deferred to v2.8.1.
- **xAI voice cloning management UX** — only the "Cloning supported" badge ships in v2.8.0. A full voice-management surface is a follow-up.
- **Bulk re-tag for legacy NULL-tenant Kanban tasks** — carryover from v2.7.5; Hermes still has no `tenant` mutation verb post-create.
- **Cluster A wire-shape TODOs** — 25 `// TODO(WS-N-Q<n>)` markers across the codebase flag fields and CLI flags whose exact shape couldn't be verified from release notes alone. Each has a tolerant-decode default that fails closed (hides the affordance rather than throwing); a pre-merge sweep on a v0.13 host can confirm or fix each in seconds.

### Acknowledgements

v2.8.0 was driven by a 9-stream coordinated multi-agent build: WS-1 capability flag foundation through WS-9 iOS catch-up, with planning artifacts archived under [scarf/docs/v2.8/](scarf/docs/v2.8/) for future reference. Bug fixes for the dashboard flicker and sparse-slash-menu issues were caught during a fresh end-to-end dogfood pass against a live Hermes v0.13.0 install — the kind of surface-level UX bugs that only show up under real-world `state.db-wal` write rates and real-world resume flows. As always, real bugs come from doing instead of speculating.
