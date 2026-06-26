---
id: t-aud30
title: **[followup/t-aud24]** Two remaining pieces from t-aud24 (goal #1 — the re-fetch elimination — is DONE): (1) **goal #2 cancellable-load** — apply the t-aud11 pattern (stored `loadTask`, cancel on disappear, check `Task.isCancelled` between SSH round-trips) to the async feature VMs (Plugins/Webhooks/Cron/MCPServers/Models/Settings); deferred because goal #1 already makes an in-flight load populate the cache rather than waste it, so this is a secondary optimization best done WITH runtime verification. (2) **Runtime verification owed** — couldn't drive the GUI headlessly: confirm A→B→A switches don't re-fetch (SSH counters/logs), state + scroll preserved, `let`/`@Bindable` VMs observe correctly, multi-window + server-switch (cache drop) behave, no memory regression (Instruments), and ⌘, / sidebar / window-restoration still work. Optionally extend the cache to Health (currently re-fetches on re-entry; has cancelLoad from t-aud11). Risk: LOW–MED.
status: todo
added: 2026-06-13, source: t-aud24
---

## Description



## Plan



## Artifacts



