---
id: t-5ab57e6f
title: Photon / iMessage gateway platform support (v0.17) — decide support level
status: todo
added: 2026-06-21
priority: low
---

## Description

Deferred from the Hermes v0.17 Tier 2 catch-up by decision (2026-06-21): iMessage protocols keep changing/being added upstream, so hold on supporting photon at a deep level until it stabilizes.

Findings (verified against v0.17 source): photon = iMessage via "Photon Spectrum", 24th gateway platform. Config is all env-var driven (PHOTON_PROJECT_ID, PHOTON_PROJECT_SECRET, PHOTON_ALLOWED_USERS, PHOTON_SIDECAR_PORT, PHOTON_HOME_CHANNEL, PHOTON_MARKDOWN/REACTIONS/REQUIRE_MENTION). Credentials are minted by `hermes photon setup` (plugins/platforms/photon/cli.py), which: (1) runs a device-code OAuth flow (photon/auth.py request_device_code → verification_uri_complete + user_code + poll — SAME shape as Nous, reusable with NousAuthFlow/OAuthFlowController), then (2) interactively registers a project + user (input() prompts; pass --project-name/--phone/--first-name/--last-name/--email as flags to avoid blocking), then (3) runs `npm install` for a Node sidecar (Node 18; host-side, awkward over SSH). Separate `hermes photon install-sidecar` / `status` subcommands exist.

When ready to support: capability flag `hasPhotonPlatform` already exists (HermesCapabilities, test/doc pattern). Reuse the device-code flow (generalize the parser for photon's `Open this URL:/Enter the code:` format), add a form that collects project/user/allowed-users and drives `hermes photon setup <flags> --no-browser`, gate local-first with a Node/SSH note. NOT in the roster yet (intentionally omitted from KnownPlatforms.all).

## Plan



## Artifacts



