# Hermes Desktop v0.6.0

`v0.6.0` is the release where Hermes Desktop becomes a fuller host workbench
without changing its core promise: the Hermes host remains the source of truth,
and the app stays on the direct SSH path.

Since `v0.5.5`, the app has gained a much stronger Files workspace, a native
session workbench, upstream Kanban support, more resilient terminal and SSH
behavior, refreshed public docs, and tighter localization coverage.

## Highlights

- SSH-backed workspace file browsing and bookmarks for selected remote text
  files, alongside the canonical Hermes files, with conflict-aware saves and a
  10 MB editable file limit
- native session workbench with richer search, pinned sessions, readable
  transcripts, compact tool-output summaries, in-app chat continuation, and
  terminal resume
- upstream Kanban board support for the host-wide `~/.hermes/kanban.db`,
  including create, assign, comment, block, unblock, complete, archive, delete,
  run history, event history, worker log visibility, and dispatcher nudging
- terminal and SSH reliability improvements, including resize/reflow fixes,
  isolated interactive shells, shorter SSH control socket paths, and
  recreation of the temporary control socket directory if macOS prunes it
- broader UI polish across workbench layouts, creation buttons, search affordances,
  and detail panes
- English, Simplified Chinese, and Russian localization tables kept in sync by
  release-support tests
- refreshed README preview gallery, install guidance, release artifact guidance,
  and GitHub Pages website copy for the v0.6.0 surface

## Still True

- Hermes Desktop still connects directly over SSH
- the host remains the source of truth
- there is no gateway API, remote daemon, local mirror, or shadow sync layer
- workspace files are opened and saved on the host; bookmarks are local pointers,
  not a second copy of the remote filesystem
- the app stays focused on the real Hermes workflow instead of inventing a
  second transport model

## Notes

- universal macOS build for Apple Silicon and Intel
- open source
- ad-hoc signed and not notarized yet, so first launch may still require
  right-click -> Open / Open Anyway
- release archive: `HermesDesktop.app.zip`
- checksum: `HermesDesktop.app.zip.sha256`
