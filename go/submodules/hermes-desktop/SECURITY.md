# Security Model

Hermes Desktop is a native macOS client for Hermes that operates directly on
this Mac or connects to another machine over SSH.

The active Hermes machine stays the source of truth. The app does not introduce
a gateway API, background daemon, synchronized mirror, or separate sync layer.

This document describes the current implementation in this repository. It is
not a promise about future packaging, signing, or infrastructure changes.

## What Executes Locally

Hermes Desktop runs as a normal macOS app on your Mac.

Local execution includes:

- the app UI and local state handling
- direct `/bin/sh` service commands when the connection type is `This Mac`
- a local embedded terminal when the connection type is `This Mac`
- `/usr/bin/ssh` and an embedded SSH shell when the connection type is
  `SSH Host`
- the built-in update check, which requests the latest Hermes Desktop release
  metadata from the GitHub Releases API

The app does not install a helper service on the host or on your Mac.

Direct-local commands run with the current macOS user's permissions against
that account's real Hermes files. The app reuses the same `HERMES_HOME`, named
profile, custom-home, PATH, Hermes CLI, service, and terminal bootstrap rules
used for SSH connections.

## What Can Execute Remotely Over SSH

For an `SSH Host` connection, Hermes Desktop uses SSH to execute commands on
the selected host.

Current remote execution includes:

- non-interactive `python3 -` invocations for app features that query or modify
  Hermes state on the host
- remote shell startup for the embedded terminal
- remote `hermes` CLI invocations for in-app chat and session resume flows

SSH mode therefore depends on the remote SSH environment you already trust:

- SSH access must already work from Terminal on your Mac
- `python3` must exist on the host for service-style RPC requests
- `hermes` must be available on the host's non-interactive SSH `PATH` for
  in-app chat and terminal resume workflows

## Local State Stored On Your Mac

Hermes Desktop stores a small amount of local state under:

`~/Library/Application Support/HermesDesktop`

Current files written there include:

- `connections.json`
  Saved connection definitions including connection type, label, SSH routing
  fields when applicable, and the selected Hermes profile or custom home
- `preferences.json`
  App preferences and lightweight workspace state such as last-used
  connection, terminal theme and font preferences, app appearance preference,
  background image metadata, update-check preference, bookmarked Hermes files,
  pinned sessions, sidebar order, and workflow presets

If you choose a custom appearance background image, Hermes Desktop copies that
image into its app support directory so the preference remains available after
the original file moves. That image is local to your macOS user account and is
not uploaded by the app.

These files are written with private file permissions (`0600`) and the app
support directory is created with private directory permissions (`0700`).
They are ordinary JSON files under your macOS user account, not Keychain
entries.

For SSH connections, the app also creates control sockets under:

`/tmp/hd-<uid>`

That directory is also created with private directory permissions (`0700`).

## What Is Not Stored Locally

Hermes Desktop does not maintain a synchronized mirror of Hermes state.

In the current implementation, it does not store these Hermes artifacts as a
second source of truth:

- session databases and transcripts
- Kanban databases
- cron job definitions
- Hermes skill directories
- workspace files as a synchronized mirror

Unsaved edits can still exist transiently in app memory while you are working,
but the app's design is to read and write the canonical state on the active
Hermes machine.

Workflow presets are the notable local exception: they are intentionally stored
as lightweight launch helpers on your Mac, scoped to the active
connection/profile, and used only to seed a fresh local or SSH Terminal
session. They are not a mirror of Hermes state.

## Secrets And Credentials

Hermes Desktop does not ask you to enter an SSH password into the app.

Current connection profiles store the connection type, Hermes profile details,
and—when using SSH—routing details such as alias, host, user, and port. They do
not contain SSH private keys, API keys, or a stored SSH password.

The app assumes SSH authentication is already handled by your existing macOS
and SSH setup.

## Network Calls

Hermes Desktop keeps its network surface intentionally small.

In the current implementation, network calls are:

- SSH connections only for hosts you explicitly configure as `SSH Host`
- an optional GitHub API request to
  `https://api.github.com/repos/dodo-reach/hermes-desktop/releases/latest`
  when checking whether a newer Hermes Desktop version exists

The built-in update check does not update Hermes Agent, does not install
anything automatically, and does not send your host, profile, session, file, or
Kanban content to GitHub.

## What You Can Verify Yourself

If you want to validate the app before trusting it:

- inspect this repository and build from source with
  `./scripts/build-macos-app.sh`
- compare the published release checksum with
  `shasum -a 256 HermesDesktop.app.zip`
- verify the installed bundle with
  `codesign --verify --deep --strict /Applications/HermesDesktop.app`
- observe live connections with Little Snitch, LuLu, `lsof`, or `nettop`
- compare this document with the current code, especially the local/SSH
  transport switch, local storage, update check, and packaging scripts

For release-specific details and the limits of those checks, see
[docs/distribution.md](docs/distribution.md).

## Distribution Limitation

The current public app bundle is ad-hoc signed and not notarized by Apple.

That means:

- macOS can verify bundle integrity after signing
- the signature does not identify a specific Apple Developer account
- the release does not carry Apple notarization attestation
- cautious users may reasonably prefer to build from source instead of trusting
  the published zip

This limitation is documented so users can make an informed trust decision. It
is not hidden, and it should not be described as stronger assurance than it is.

## Reporting Security Issues

If you discover a security issue, avoid posting sensitive exploit details in a
public issue first.

Use GitHub's private vulnerability reporting flow for this repository if it is
available. Otherwise, contact the maintainer privately through GitHub before
public disclosure.
