# Hermes Desktop v0.7.0

`v0.7.0` is a release about staying close to upstream Hermes while making the
desktop workspace feel more deliberate.

Since `v0.6.1`, the biggest change is the Kanban upgrade: Hermes Desktop now
understands the newer upstream board model, while keeping the same host-first
SSH design. The default board still lives at `~/.hermes/kanban.db`, and newer
Hermes Agent builds can expose additional boards under the host-wide Kanban
home.

This release also adds a small but useful app update check, refreshes the
README preview gallery, and gives the embedded terminal a more polished set of
appearance controls.

## Highlights

- upstream Kanban board management, including board selection, board creation,
  board archive, and fallback behavior for hosts that only expose the default
  `~/.hermes/kanban.db` board
- richer Kanban task operations over SSH: task creation, search, filters,
  assignment, comments, block, unblock, complete, archive, delete, run history,
  event history, worker log tailing, and dispatcher nudging when the host
  supports it
- Kanban compatibility checks that clearly tell users when a newer Hermes Agent
  build is needed for multiple boards or home-channel subscriptions
- terminal appearance polish with six stable presets, custom background and
  text color tuning, and anchored ANSI palettes so prompts, git output, and
  command-line tools stay readable
- built-in Hermes Desktop update checks against GitHub Releases, with automatic
  checks limited to the app itself and manual checks available from the Hermes
  menu
- refreshed README preview gallery with six current screenshots: Sessions,
  Kanban, Files, Usage, Skills, and Terminal
- documentation updates that describe the host-wide Kanban home, app-only update
  checks, and the `0.7.0` release packaging path
- continued localization coverage for English, Simplified Chinese, and Russian
  strings touched by this release

## Compatibility

- Multiple Kanban boards require a Hermes Agent build with the newer upstream
  board APIs.
- Hosts with only the default upstream Kanban database still use
  `~/.hermes/kanban.db`.
- The update checker checks Hermes Desktop releases only. It does not update
  Hermes Agent on the host and does not send host, profile, file, session, or
  Kanban content to GitHub.

## Still True

- Hermes Desktop still connects directly over SSH.
- The Hermes host remains the source of truth.
- There is no desktop gateway API, remote helper daemon, local mirror, or shadow
  sync layer.
- Kanban, sessions, cron jobs, files, skills, usage, and terminal work stay
  anchored to the selected host.
- The app continues to favor a native Mac surface for the real Hermes workflow
  over a second transport model.

## Notes

- universal macOS build for Apple Silicon and Intel
- open source
- ad-hoc signed and not notarized yet, so first launch may still require
  right-click -> Open / Open Anyway
- release archive: `HermesDesktop.app.zip`
- checksum: `HermesDesktop.app.zip.sha256`
