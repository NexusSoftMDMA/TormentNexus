# Hermes Desktop v0.5.0

`v0.5.0` is the release where Hermes Desktop starts to feel like a complete
native Mac workspace for Hermes Agent over SSH.

Since `v0.4.1`, the app has grown from a strong SSH-first companion into a much
fuller daily environment for living with Hermes on macOS: cron jobs are now
first-class, Hermes profiles on the same host stay coherent across the app, the
terminal is more mature with tabs and appearance controls, and usage can now
surface host-wide totals when more than one readable profile is available.

Just as important, the broader Hermes ecosystem is clearer now. Nous Research
ships the official Hermes web dashboard, and that makes the split cleaner:
Hermes Desktop is not trying to replace the browser dashboard. It is the native
Mac workspace for direct SSH-based daily use.

## Highlights

- first-class cron job workflows on the live Hermes host, including browse,
  create, edit, pause, resume, run-now, and delete flows
- profile-aware host workflows across overview, usage, cron, skills, files, and
  terminal behavior
- terminal tabs, appearance presets, and color customization for a stronger
  long-running shell experience on macOS
- host-wide usage totals across readable Hermes profiles when that data is
  available
- tighter overview and skill workflows, with better workspace clarity and
  remote editing support
- version-stamped universal macOS packaging for Apple Silicon and Intel Macs

## Still True

- Hermes Desktop still connects directly over SSH
- the host remains the source of truth
- there is no gateway API, remote daemon, local mirror, or shadow sync layer
- the app stays focused on the real Hermes workflow instead of inventing a
  second transport model

## Notes

- universal macOS build for Apple Silicon and Intel
- open source
- not notarized yet, so first launch may still require right-click -> Open /
  Open Anyway
