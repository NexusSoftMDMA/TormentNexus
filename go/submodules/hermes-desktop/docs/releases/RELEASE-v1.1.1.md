# Hermes Desktop v1.1.1

`v1.1.1` is a focused Settings and connection-hardening release.

It keeps the same direct SSH-first model from `v1.1.0`, while making host
management calmer when you move between network paths, SSH aliases, or saved
hosts. The selected Hermes host remains the source of truth: no gateway service,
no local mirror, and no background sync layer.

## What Changed

- Settings now serializes automatic host discovery refreshes for the active
  workspace, reducing duplicate probes and avoiding misleading transient
  connection-health states.
- Switching hosts now clears the previous "Last Checked" timestamp so Settings
  always reflects the currently selected host.
- Host management now includes a confirmed "Remove Host" action. Removing a host
  only deletes the saved local connection from Hermes Desktop; remote Hermes
  files and profiles are left untouched.
- The first "Edit Host" action now opens the selected host details immediately,
  so the sheet shows the connection being edited from the first attempt.

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
