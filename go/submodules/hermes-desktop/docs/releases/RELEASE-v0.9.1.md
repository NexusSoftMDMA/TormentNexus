# Hermes Desktop v0.9.1

`v0.9.1` is a small fixes and hardening release.

It keeps the same direct SSH-first model from `v0.9.0`, while smoothing a few
rough edges reported by early users. The goal is simple: make the desktop app a
little calmer and more ready for the next bigger step.

No extra gateway layer. No local mirror. The selected Hermes host remains the
source of truth.

## What Changed

- SSH service checks now retry once without connection multiplexing when the
  first attempt fails with a reachability-style error. This helps cases where a
  stale SSH control socket makes Desktop look less healthy than Terminal.
- Kanban board loading is more resilient when a host has Hermes Agent code and
  an on-disk Kanban database that are out of sync. Desktop can show a direct
  database view when possible and explains that the host schema needs attention.
- Session transcripts hide leaked terminal control fragments, including partial
  ANSI color sequences that can appear as stray text in the message column.
- Session model labels now prefer the latest model metadata found in the
  transcript, so a live model switch is less likely to leave stale labels in the
  session list.
- Embedded Chat startup no longer sends the initial fallback input twice.

## Compatibility

- macOS 14 or newer
- SSH from this Mac to the Hermes host must already work without interactive
  prompts
- `python3` must be available on the host
- Chat and Terminal resume require the remote `hermes` CLI on the
  non-interactive SSH `PATH`
- public releases are still ad-hoc signed and not notarized by Apple

## Still True

- Hermes Desktop connects directly over SSH
- the Hermes host remains the source of truth
- sessions, Kanban, cron jobs, files, skills, usage, Chat, and Terminal all stay
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
