# Hermes Desktop v0.9.0

`v0.9.0` is the Hermes Desktop release I suggest everyone try.

This is the chat release. Sessions now has a more mature split between the
stored transcript and a live Chat view, and that Chat view runs the real Hermes
TUI over the same direct SSH path Hermes Desktop already uses. It feels better,
it is easier to reason about, and it keeps the product aligned with the core
idea: the host remains the source of truth.

No browser wrapper. No extra desktop gateway API. No background sync layer. No
local mirror of Hermes state.

## Highlights

- `Sessions` now separates `Transcript` from `Chat`, so you can inspect stored
  history without confusing it with the live Hermes TUI
- `New Chat` starts the real `hermes --tui` flow on the selected SSH host and
  profile
- existing sessions can be resumed in the embedded Chat TUI with the selected
  Hermes profile preserved
- the sessions layout is calmer and more durable, with pinned sessions, clearer
  session actions, and a better path for long-term maintenance
- Workflows can now run in Chat or Terminal, using the same SSH-first launch
  model
- Overview now surfaces Chat readiness in terms of the embedded Hermes TUI and
  the transcript source Desktop reads back from the host

## Fixes

- fixes `#31` (`Kanban unable to load`) and improves the same failure family
  reported in `#35` (`Unable to load cron jobs and kanban`)
- Desktop now runs remote Python service probes through the same prepared
  Hermes service environment, reducing failures caused by the wrong
  non-interactive Python or PATH
- Cron job loading now tolerates the timestamp precision emitted by newer
  Hermes scheduler metadata

If one of these host-specific loading issues still reproduces after installing
`v0.9.0`, the most useful next diagnostics are from the same SSH target Desktop
uses:

```bash
command -v python3
python3 --version
command -v hermes
hermes --version
```

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
