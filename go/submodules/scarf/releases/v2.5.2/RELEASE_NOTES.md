## What's in 2.5.2

A patch with one substantial new feature (**iOS chat resilience** — reconnect, cached snapshot fallback, history paging) plus a stack of fixes for issues reported against 2.5.1 and earlier. Drop-in replacement for 2.5.1 on Mac; drop-in TestFlight build on iOS. No data migrations.

### iOS chat resilience

ScarfGo now survives phone-sleep, network handoffs, and SSH socket drops without losing the agent's work. Hermes was already persisting messages to `state.db` in real-time; iOS just had no resync path.

- **5-attempt exponential reconnect** (1s → 2s → 4s → 8s → 16s) via `session/resume` with `session/load` fallback. Reconciles with `state.db` on success and surfaces a *"Resynced N new messages"* toast when the agent kept working through the disconnect.
- **`NetworkReachabilityService`** (NWPathMonitor singleton): suspends reconnect attempts while offline and kicks a fresh cycle on link-up. Two new banner states above the message list — `.reconnecting` and `.offline` — render as slim ScarfDesign-tinted strips so the user always knows what the chat is doing.
- **Scene-phase awareness**: returning to foreground triggers a channel-health check; if dead, the reconnect cycle starts immediately rather than waiting for the next interaction.
- **Draft persistence**: per-server, per-session draft survives force-quit (UserDefaults-backed, 7-day janitor at app launch).

### Cached snapshot fallback (Mac + iOS)

`ServerTransport.cachedSnapshotPath` lets `HermesDataService` fall back to the previously-pulled `state.db` snapshot when a fresh pull fails. `isUsingStaleSnapshot` + `lastSnapshotMtime` surface to views so they render *"Last updated X ago."* Chat-history reload still passes `forceFresh: true` to refuse stale data; everything else (Dashboard, Sessions list, Activity) gets read-while-disconnected for free.

### Bounded message-history paging

`HermesDataService.fetchMessages(sessionId:limit:before:)` paginates by id desc with centralized `HistoryPageSize` constants. `RichChatViewModel.loadEarlier()` walks back through long sessions via `oldestLoadedMessageID` + `hasMoreHistory`. Legacy unbounded overload deprecated.

### Bug fixes

#### Mac

- **[#46](https://github.com/awizemann/scarf/issues/46) — chat O(n)-per-token bog-down (already shipped in 2.5.1 for the trailing-group patch; this release retains the fix and pairs with the new history paging so chats with thousands of messages stay smooth).**
- **[#19](https://github.com/awizemann/scarf/issues/19) layer-3 — sqlite3 false-negative in diagnostics.** Already in v2.5.1; kept here.
- **[#44](https://github.com/awizemann/scarf/issues/44) — pill / diagnostics agreement** via shared `SSHScriptRunner`. From v2.5.1; the tier-2 probe now also checks `state.db` (not just `config.yaml`) so a healthy fresh install reports green.
- **[#59](https://github.com/awizemann/scarf/issues/59) — Settings → Model and Credential Pools no longer freeze.** Both views called `ModelCatalogService.loadProviders()` synchronously from `.onAppear` on the MainActor; on a remote SSH context that's a multi-megabyte SSH file read on the main thread, freezing the UI for 1–2 minutes. New `loadProvidersAsync()` / `loadModelsAsync(for:)` wrappers dispatch off the main thread; both views now use `.task` + `await` with a `ProgressView("Loading providers…")` overlay. Per-provider switching in the picker is also async now, so clicking a different provider doesn't re-freeze the UI.
- **Diagnostics tri-state.** Hermes v0.11+ doesn't materialize `config.yaml` until the user changes a setting from defaults — so the diagnostics view was reporting *"12/14 passing"* on healthy fresh installs. The probe now distinguishes `.pass` / `.fail` / `.skipped`; a missing `config.yaml` emits SKIP and is excluded from the summary's denominator. Reads as *"12/12 passing (2 optional skipped)"* instead of the misleading 12/14.
- **Credentials: OAuth providers visible.** `hasAnyAICredential()` only probed `credential_pool.<provider>` in `auth.json`; OAuth-authed providers land under `providers.<name>.access_token` (Nous, Spotify, GH Copilot ACP, Qwen, Gemini all use that path). The chat banner kept showing *"No AI provider credentials"* even after a successful Nous sign-in. Now both shapes count. Credential Pools view gains a parallel "OAuth providers" section listing OAuth-authed providers with token tail, expiry badge, and portal URL.
- **Project-shadowed Hermes detection.** New `ProjectHermesShadowDetector` (ScarfCore) probes each registered project at chat-start; if a `.hermes/` dir or `hermes.yaml` is found inside the project, the user gets a banner explaining that project-local Hermes config will shadow the server-level one (a quiet failure mode for users who didn't realize Hermes prefers project-local config).
- **[#58](https://github.com/awizemann/scarf/issues/58) — Mac chat side panes are hideable.** Two toolbar buttons next to the View picker (`sidebar.left` / `sidebar.right`) toggle the sessions list and tool inspector with a slide animation; both default visible (today's behavior). Clicking a tool card auto-shows the inspector if hidden so the click never silently dies. Settings → Display → Chat density gains parity Toggle rows.

#### ScarfGo (iOS)

- **[#56](https://github.com/awizemann/scarf/issues/56) — *"Citadel.SSHClient.CommandFailed error 1"* on dashboard.** `asyncSnapshotSQLite` was missed during the v2.5.0 Citadel hardening — used raw `executeCommand` (which discards stderr on non-zero exit) and didn't prepend the Citadel-friendly `PATH=$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:$PATH`. Now uses `executeCommandStream` and the same PATH prefix. `HermesDataService.humanize` already translates `sqlite3: command not found` / `permission denied` / `no such file` into actionable user copy — the bug was that the snapshot path never fed it real stderr.
- **[#57](https://github.com/awizemann/scarf/issues/57) — keyboard-dismiss chevron over send button.** The keyboard accessory dismiss button added in v2.5.1 (#51) was placed at the trailing edge of the keyboard toolbar, directly above the trailing-edge send button. Moved to the leading edge — matches the iOS convention (Notes, Mail, Reminders).

### New features (Mac)

- **Chat-start model preflight ([commit](https://github.com/awizemann/scarf/commit/2aab9da)).** Catches a missing `model.default` / `model.provider` in `config.yaml` *before* the ACP session starts. Pre-fix the user typed a prompt, hit send, and got an opaque *"Model parameter is required"* HTTP 400 from the upstream provider. Now `ChatModelPreflightSheet` wraps the existing model picker so the same selection / validation / Nous-catalog branch is single-sourced; the chat the user originally opened lands without re-clicking the project row.
- **Nous Portal live model catalog.** `NousModelCatalogService` fetches `GET /v1/models` from `inference-api.nousresearch.com` using the bearer token in `auth.json`. Cached at `~/.hermes/scarf/nous_models_cache.json` with a 24h TTL. The picker's nous-overlay detail view switches from a free-form TextField to a real model list, with a *"Custom…"* escape hatch for IDs not yet in the API response.
- **Remote-aware admin sheets.** Three sheets gained the same context-aware Verify pattern that Add Project got in v2.5.1 (#54):
  - **Profiles → Import / Export.** Buttons that drive `hermes profile import <zip>` / `hermes profile export <name> <zip>` over SSH. Local context picks via `NSOpenPanel`; remote context shows a path-input + Verify button.
  - **Settings → Advanced → Restore.** Pick a local backup zip OR enter+verify a remote path.
  - **Templates → Install destination.** The parent-directory step in the install sheet branches on context — local Browse, or remote text-input + Verify.

### Translations

`Localizable.xcstrings` adds strings for all the new copy across the seven supported locales (English, Simplified Chinese, German, French, Spanish, Japanese, Brazilian Portuguese).

### Notes for users running 2.5.1

No data migrations needed. `~/.hermes/scarf/nous_models_cache.json` is created lazily on first use of the Nous picker; everything else is forward-compatible with existing config / Keychain / project registries.
