# Hermes Desktop v0.8.0

`v0.8.0` adds reusable terminal workflows, tighter SSH behavior, and a more
complete release verification path.

Hermes Desktop still talks directly to the selected Hermes host over SSH. The
host remains the source of truth. There is no gateway API, remote helper
daemon, local mirror, or background sync layer added in this release.

## Highlights

- new `Workflows` workspace for reusable prompt presets scoped to the active
  host/profile, with optional skills and one-click launch into a fresh
  Terminal tab
- workflow presets stay local to the Mac and do not create remote shadow state
- more resilient non-interactive SSH behavior, including clearer `python3`
  PATH errors and better handling when shell startup output breaks app requests
- stronger regression coverage for SSH transport, terminal input submission,
  workflow persistence, connection storage, file editing, and localization
- new macOS CI, release manifest generation, and packaged-release verification
- new public trust docs: `SECURITY.md` and expanded distribution guidance

## Compatibility

- The app still requires SSH access from this Mac to the Hermes host, with
  `python3` available on the host.
- In-app chat, terminal resume, and workflow launch paths still require the
  remote `hermes` CLI to be available on the host's non-interactive SSH
  `PATH`.
- Workflow skill preloading depends on the host exposing the selected skills in
  its Hermes skills store for the active profile.
- Public releases are still ad-hoc signed and not notarized by Apple.

## Still True

- Hermes Desktop still connects directly over SSH.
- The Hermes host remains the source of truth.
- Sessions, Kanban, cron jobs, files, skills, usage, and terminal work stay
  anchored to the selected host and profile.
- Workflow presets are local launch helpers, not a second transport model or
  synchronization layer.

## Notes

- universal macOS build for Apple Silicon and Intel
- ad-hoc signed and not notarized yet, so first launch may still require
  right-click -> Open / Open Anyway
- release archive: `HermesDesktop.app.zip`
- checksum: `HermesDesktop.app.zip.sha256`
- manifest: `HermesDesktop.app.zip.manifest.json`
