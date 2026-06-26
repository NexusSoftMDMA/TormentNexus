# Hermes Desktop v0.8.1

`v0.8.1` is a focused polish release for workflows, startup reliability, and
custom Hermes setups.

Hermes Desktop still connects directly to the selected Hermes host over SSH.
The host remains the source of truth. This release does not add a gateway API,
helper daemon, local mirror, or background sync layer.

## Highlights

- workflows now support even longer prompts, giving the new workflow launcher
  more room for richer setup and handoff
- workflow startup has been hardened for more reliable prompt delivery into
  Terminal, with better diagnostics around the handoff path
- custom Hermes home paths are now supported more cleanly across path
  resolution, command launch, and terminal bootstrap flows

## Compatibility

- the app still requires SSH access from this Mac to the Hermes host, with
  `python3` available on the host
- in-app chat, terminal resume, and workflow launch paths still require the
  remote `hermes` CLI to be available on the host's non-interactive SSH `PATH`
- public releases are still ad-hoc signed and not notarized by Apple

## Still true

- Hermes Desktop still connects directly over SSH
- the Hermes host remains the source of truth
- sessions, Kanban, cron jobs, files, skills, usage, and terminal work stay
  anchored to the selected host and profile
- workflow presets remain local launch helpers, not a second transport model or
  synchronization layer

## Notes

- universal macOS build for Apple Silicon and Intel
- ad-hoc signed and not notarized yet, so first launch may still require
  right-click -> Open / Open Anyway
- release archive: `HermesDesktop.app.zip`
- checksum: `HermesDesktop.app.zip.sha256`
- manifest: `HermesDesktop.app.zip.manifest.json`
