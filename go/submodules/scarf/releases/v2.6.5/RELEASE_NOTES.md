## What's in 2.6.5

A patch release that ships **template discoverability**, **cron observability**, and an **end-to-end UI test harness** that locks the new install path against regression. No breaking changes; every Hermes capability target is unchanged from 2.6.0.

### In-app Template Catalog

The catalog is no longer web-only. **Templates → Browse Catalog…** opens a sheet that fetches the live catalog from `awizemann.github.io/scarf/templates/`, renders one row per published template with name + version + tags, and one-click installs through the existing flow. Search filters across name / description / tags; the category picker constrains to whatever categories the loaded catalog actually carries.

- **Install-state badges** — each row shows "Installed v1.2.0" (green) or "Update v1.3.0" (amber) when the catalog version is newer than what's in `~/.hermes/scarf/projects.json`. Update is "uninstall + reinstall" today; in-place upgrade is on the v3 backlog.
- **24h cache** at `~/.hermes/scarf/catalog_cache.json` so opening the sheet repeatedly doesn't re-hit the network. Refresh icon force-fetches.
- **Bundled fallback** — fresh-install / offline users still see the official templates as a hardcoded list. Network failures serve stale cache with a "refresh failed" hint.
- **Catalog-schema decoder fault tolerance** — one malformed entry on the live catalog can't bring down the whole list. The bad row is dropped with a logged warning; the rest survive.

### HackerNews Daily Digest template

First template added under the new dogfooding-templates loop. Configurable `min_score`, `max_items`, `topics`; one daily-at-08:00 cron job (paused on install) that pulls the HN Firebase API, filters, and prepends a markdown digest to the project's `digest.md`. No API keys required. Live at the catalog URL above.

### Cron observability — auth-error banner + running indicator + log tail

Cron rows now surface the same OAuth-refresh-revoked recovery flow as Chat instead of a generic red dot, plus three previously-missing observability cues:

- **OAuth re-auth.** `ACPErrorHint.classify` runs on `job.lastError`; when it returns `oauthRefreshRevoked(provider)` the detail pane shows the human-readable hint + a **Re-authenticate** button that drops the user into Credential Pools — same wiring ChatView's banner uses. Unrecognized errors fall back to the legacy red `lastError` text.
- **Running indicator.** The row dot turns blue + pulses when `state == "running"` (precedence over disabled / error / success); the detail header gains a "running…" badge next to active/paused. No new polling — `HermesFileWatcher.lastChangeDate` already drives `CronViewModel.load()`.
- **Last run output.** Collapsible panel replacing the inline log: a one-line summary (`<timestamp> — ok|error|running…`) always visible, full monospaced terminal-style scroll on expand, auto-scrolls to bottom when new runs land.

Also fixes a pre-existing bug in `HermesFileService.loadCronOutput` that returned the wrong file under Hermes's per-job-id output nesting.

### Layer B install-drive XCUITest harness

The dogfooding-templates initiative ships its first end-to-end UI test that drives the install pipeline:

```
Launch with --scarf-test-mode → Sidebar → Projects → Install sheet
(via --scarf-test-install-url launch arg) → Configure → Open Project
→ Right-click → Uninstall Template → Confirm Remove → Done
```

Runs ~30 s green on the dev Mac, validates 9 assertion points across the user journey. Covers the new accessibility identifiers wired in this release: `templateConfig.commitButton`, `projects.row.<name>`, `sidebar.section.<rawValue>`, `projects.contextMenu.uninstallTemplate`, `templateUninstall.confirmRemove`, `templateInstall.success.openProject`, `templateUninstall.success.done`. The `--scarf-test-install-url` launch arg + `TestModeFlags.isTestMode` gating lets XCUITest skip SwiftUI Menu / NSToolbarItem accessibility-bridging quirks that otherwise block toolbar-menu driving.

Wiki [Test-Harness](https://github.com/awizemann/scarf/wiki/Test-Harness) documents how to extend the harness for the next template.

### Sentinel-marker test isolation (incident-response hardening)

`SCARF_HERMES_HOME` override now requires the path to contain a `.scarf-test-home-marker` file to activate. Without the marker, production code falls through to the user's real `~/.hermes/`. Lands belt-and-braces protection for cases where a test crashes mid-teardown leaving the env var set, an env var inherits from a parent shell, or a misconfigured launchctl plist exports the variable. The override remains the seam every E2E test relies on; the marker file ensures it can't accidentally pivot a non-test process off the user's data.

### Chat fixes

- **OAuth refresh-revoked surface.** Chat-side error banner now classifies the message via `ACPErrorHint.classify` and offers an in-app **Re-authenticate** button that routes through Credential Pools (#65). Same primitive the new cron banner reuses.
- **Placeholder ghosting fix.** TextEditor's placeholder now clips to the editor's bounds and clears on focus instead of bleeding past the cursor area when the user types fast (#67).

### Profile chip + structured logs

- **Active-profile chip in the sidebar header.** Click → routes to Profiles. Local contexts only (remote SSH would mislead).
- **Switch & Relaunch** flow now writes `~/.hermes/active_profile` and relaunches Scarf in a single click instead of asking the user to quit+reopen.
- Profile-resolver logs are now structured (key=value form) so `log show … | grep ProfileResolver` can pull "which profile did Scarf resolve to and why" out of support requests.

### Swift 6 cleanup

- `MessageSpeechService` — drop `@preconcurrency` on the AVSpeechSynthesizerDelegate conformance now that the protocol's Sendable annotations are upstreamed.
- `ChatView` — `RichChatViewModel.PendingPermission: @retroactive Identifiable`. Quiets the Swift 6 compiler so downstream breakage would be loud if ScarfCore ever adds the conformance upstream.
- `CredentialPoolsView` — `.help(Text(verbatim:))` so backticks render literally instead of being treated as markdown inline-code.

### iOS

- Composer redesigned with HIG touch targets + clear disabled state.
- Portrait lock retained.
- Chat-start preflight moved off MainActor.

### Known caveats

- **Cron-job-uninstall by name is ambiguous** when two projects share the same template id. The Layer B test surfaced this — manifests as: the test passes, but if you've manually installed the same template before running the test, your real cron job can disappear. Recovery is `hermes cron create`. Fix is queued: store cron-job IDs in `<project>/.scarf/template.lock.json` at install time and resolve by ID at uninstall time.
- **Full-suite parallel test runs intermittently hang** — pre-existing flaky test infrastructure unrelated to this release. Individual suites all pass; the hang only manifests on `xcodebuild test` with everything concurrent. The sentinel-marker hardening prevents user-data damage from any race.

### Compatibility

- **Hermes target unchanged from 2.6.0**: v2026.4.30 (v0.12.0). Pre-v0.12 Hermes hosts continue to work — no new capability gates added in this release.
- **Min macOS unchanged**: 14.6.
- **No schema changes** to anything in `~/.hermes/`. The two new Scarf-owned files (`scarf/catalog_cache.json` and the template-installer's `.scarf-test-home-marker` for tests) are additive.
