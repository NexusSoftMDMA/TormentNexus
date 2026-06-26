# Hermes Desktop v0.7.1

`v0.7.1` is a focused reliability release for people already working inside
Hermes Desktop every day. It improves in-app chat approval handling, adds more
complete Kanban operations, and makes cron jobs more flexible while preserving
the same direct SSH-only model.

Hermes Desktop still talks to the selected Hermes host over SSH. The host
remains the source of truth. There is no gateway API, remote helper daemon,
local mirror, or background sync layer added in this release.

## Highlights

- safer in-app chat behavior when Hermes requests command approval during a
  non-interactive turn
- clearer `Auto-approve commands` copy: it approves command requests for the
  current turn, and without it approval-required commands may be blocked in
  chat
- a new approval-needed state when a chat turn cannot continue usefully because
  manual approval is required; users can retry with auto-approve enabled or
  resume the same session in Terminal to review the command themselves
- no forced interruption when Hermes receives a denial and can still continue:
  in that case the transcript shows Hermes' normal response
- richer Kanban task editing, including dependency links and task metadata
  updates from the native board UI
- Kanban recovery actions for tasks that need operator attention after warning
  or worker recovery states
- script-only cron jobs, which run a host script and deliver stdout directly
  without creating an agent turn
- cron job details now show script and working-directory metadata where
  available, making scheduler state easier to inspect from the app
- menu command intent handling is more stable when switching sections during
  app-level flows
- continued localization coverage for English, Simplified Chinese, and Russian
  strings touched by this release

## Compatibility

- The app still requires SSH access from this Mac to the Hermes host, with
  `python3` available on the host.
- In-app chat still requires the remote `hermes` CLI to be available on the
  host's non-interactive SSH `PATH`.
- Auto-approve uses Hermes' own auto-approval mode for that chat turn. Leave it
  off for safer default chat turns; use it only when you intentionally want the
  turn to approve command requests automatically.
- If you want to review commands manually, use `Resume in Terminal` and continue
  the session from Hermes' interactive terminal surface.
- Multiple Kanban boards and newer Kanban operations require a Hermes Agent
  build with the upstream board and task APIs used by this release.
- Script-only cron jobs appear and save correctly when the host supports the
  cron job fields used by current Hermes Agent builds.

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
