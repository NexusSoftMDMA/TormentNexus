# Hermes Desktop v1.2.0

Hermes Desktop 1.2.0 is the local Hermes release.

If Hermes Agent runs on your Mac, you can now choose `This Mac` and use the
whole Desktop workspace without setting up SSH to localhost. Sessions, Chat,
Workflows, Files, Skills, Usage, Cron, Kanban, and Terminal work directly
against your current macOS account's real Hermes installation.

The idea stays the same: Hermes Desktop does not create a second copy of your
Hermes state. It now meets Hermes where it already runs, whether that is this
Mac, a Raspberry Pi on your network, another Mac, a VPS, or a remote server.

## What Changed

- Connections now have an explicit `This Mac` or `SSH Host` mode.
- `This Mac` needs only a connection name. There is no SSH alias, hostname,
  user, port, host-key setup, or passwordless SSH requirement.
- Every main workspace surface supports local Hermes: Sessions and transcripts,
  the real Hermes TUI in Chat, Workflows, Files, Skills, Usage, Cron Jobs,
  Kanban, discovery, and Terminal.
- The embedded Terminal opens a real local shell and uses the same Hermes
  profile, custom home, PATH preparation, and launch rules as the rest of the
  app.
- Named profiles and custom `HERMES_HOME` paths work in both local and SSH
  mode.
- Actions that edit files, skills, scheduler state, sessions, or Kanban make it
  clear when they are changing this Mac's real Hermes data.

## Existing SSH Connections

Nothing about the remote workflow is being replaced. Existing saved
connections continue to load as `SSH Host`, keep the same workspace identities,
and retain the established SSH arguments, multiplexing behavior, retry path,
terminal environment, and profile scoping.

Direct-local mode is also intentionally separate from an SSH connection aimed
at `localhost`, so workflows, bookmarks, pinned sessions, and preferences do
not collide between the two.

## Also Included

- Saved connection loading is more resilient: malformed or future connection
  entries no longer prevent valid profiles from loading.
- LAN SSH failures now point to the macOS Local Network permission when that is
  the likely blocker.
- English, Simplified Chinese, and Russian copy now describes local and SSH
  operation accurately.

## Requirements

- macOS 14 or newer
- `python3` available on the machine where Hermes runs
- Hermes data under `~/.hermes`, a named profile, or a configured custom home
- Chat, Terminal resume, and Workflows require the `hermes` CLI through the
  prepared shell PATH
- SSH mode additionally requires non-interactive SSH authentication and an
  accepted host key

## Distribution

- universal macOS build for Apple Silicon and Intel
- ad-hoc signed and not notarized by Apple
- on most supported macOS versions, first launch may require right-click →
  Open / Open Anyway
- release archive: `HermesDesktop.app.zip`
- checksum: `HermesDesktop.app.zip.sha256`
- manifest: `HermesDesktop.app.zip.manifest.json`

### macOS 26.5.1 Gatekeeper note

On some Macs running macOS 26.5.1 (build 25F80), Gatekeeper may say that the
downloaded app "is damaged and can't be opened" and may not offer `Open
Anyway`. This note applies only to that version and build when that exact alert
appears.

Before continuing, verify the v1.2.0 archive:

```bash
shasum -a 256 ~/Downloads/HermesDesktop.app.zip
```

The v1.2.0 result must be:

```text
34ef72ea39e76659f81ceecbc1c42fd970ba1478bed9257d7f3e68306917057d
```

Do not continue if it differs. After extracting the verified zip and moving
the app into Applications, remove the browser quarantine from Hermes Desktop
only and open it:

```bash
xattr -dr com.apple.quarantine "/Applications/HermesDesktop.app"
open "/Applications/HermesDesktop.app"
```

This does not disable Gatekeeper globally and does not require `sudo`.

Alternatively, inspect the source and build locally. The resulting app does
not inherit browser download quarantine:

```bash
git clone https://github.com/dodo-reach/hermes-desktop.git
cd hermes-desktop
./scripts/build-macos-app.sh
open "dist/HermesDesktop.app"
```
