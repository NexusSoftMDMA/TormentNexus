# ScarfGo v2.13.0 — App Store Connect per-version copy

The set-once **App information** (name, subtitle, bundle ID, categories, support/privacy URLs, full Description, keywords) is unchanged — [`releases/v2.5.0/APP_STORE_METADATA.md`](../v2.5.0/APP_STORE_METADATA.md) remains the source of truth for those fields. This file carries only the per-version fields that change for v2.13.0. (Verify character counts against Apple's limits before pasting.)

## Promotional text (max 170 chars, editable without resubmission)

```
Switch Hermes profiles from your phone, plus rock-solid remote chat and Settings over SSH. Connect to any SSH-reachable Hermes host and run sessions on the go.
```

## What's New text (max 4000 chars)

```
Profile switching comes to ScarfGo, plus a reliability fix for remote chat and Settings.

• Switch Hermes profiles. If you run more than one profile on a host — say an admin profile and a locked-down gateway — pick one in the Profiles tab and ScarfGo points its chat, memory, cron, sessions, gateway, and every command at that profile. It scopes only its own view: it never changes the active profile your Mac, terminal, or running gateway use on that host. (Create, rename, delete, import, and export stay on the Mac app.)

• Reliable remote chat and Settings. Fixes "Couldn't save model.provider… Transport refused" and empty Settings on a host where the config is valid and the same command works over plain ssh. ScarfGo was opening a fresh SSH connection for every read and command; under a Settings load or chat start, the connection itself could fail. It now reuses one pooled connection per server.

• Skills "What's New" tracks each profile. The pill no longer shows wrong "new / changed" counts after you switch profiles.

Privacy is unchanged: no analytics, no telemetry, no developer-controlled servers. SSH keys are generated on-device and never leave it. Full policy at awizemann.github.io/scarf/privacy.
```

## Version

Marketing version **2.13.0**, in lockstep with the macOS Scarf release (project convention). Build number = the next `CURRENT_PROJECT_VERSION` (46+).

## Screenshots

Not required for a TestFlight build. If promoting this version to the public App Store, refresh per the screenshot matrix in [`releases/v2.5.0/APP_STORE_METADATA.md`](../v2.5.0/APP_STORE_METADATA.md) — add a Profiles-switcher capture to the set.
