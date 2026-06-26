# Hermes Desktop v0.7.2

`v0.7.2` is a release-candidate polish and compatibility update for people who
already work inside Hermes Desktop every day. It fixes a real session
navigation regression, adds more flexible split-view controls, and stays aligned
with recent Hermes Agent changes without weakening the app's SSH-first model.

Hermes Desktop still talks directly to the selected Hermes host over SSH. The
host remains the source of truth. There is no gateway API, remote helper daemon,
local mirror, or background sync layer added in this release.

## Highlights

- fixes issue `#25`: chat transcript scroll position is preserved when you move
  from `Sessions` to `Terminal` and back
- new collapse controls for the workspace sidebar and section browsers, so you
  can give more room to the part of the app you are actively using
- the workspace sidebar now stays visible in `Terminal`, preserving the host
  context while still allowing the rest of the workbench to collapse
- Kanban can now promote triage tasks through the upstream `specify` flow from
  the native app when the host supports it
- new Kanban task creation and inspection support for per-task retry limits,
  matching newer upstream Hermes Agent behavior
- cron jobs recognize and edit the newer `all` delivery target for connected
  channels
- skills discovery now reads `platforms` metadata, exposes it in the detail
  view, and includes it in search
- continued English, Simplified Chinese, and Russian localization coverage for
  strings touched by this release

## Compatibility

- The app still requires SSH access from this Mac to the Hermes host, with
  `python3` available on the host.
- In-app chat still requires the remote `hermes` CLI to be available on the
  host's non-interactive SSH `PATH`.
- The Kanban `Specify` action depends on the newer upstream Hermes CLI path and
  may ask you to update Hermes Agent on the host when that support is missing.
- Per-task Kanban retry limits require a Hermes Agent build that exposes the
  current `max_retries` task fields and CLI arguments.
- The `all` cron delivery target appears correctly only when the host supports
  the current Hermes scheduler delivery semantics.

## Still True

- Hermes Desktop still connects directly over SSH.
- The Hermes host remains the source of truth.
- Sessions, Kanban, cron jobs, files, skills, usage, and terminal work stay
  anchored to the selected host and profile.
- There is no desktop gateway API, remote helper daemon, local mirror, or shadow
  sync layer.

## Notes

- universal macOS build for Apple Silicon and Intel
- open source
- ad-hoc signed and not notarized yet, so first launch may still require
  right-click -> Open / Open Anyway
- release archive: `HermesDesktop.app.zip`
- checksum: `HermesDesktop.app.zip.sha256`
