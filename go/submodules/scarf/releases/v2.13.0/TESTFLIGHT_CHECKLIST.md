# v2.13.0 TestFlight submission checklist

Incremental ScarfGo TestFlight build. The one-time setup (privacy URL, App Store Connect record, signing, public group) is done — see [`releases/v2.5.0/TESTFLIGHT_CHECKLIST.md`](../v2.5.0/TESTFLIGHT_CHECKLIST.md) for the full first-time walkthrough. This file is the per-build delta.

## Pre-flight

- [ ] `main` carries v2.13.0 (merged + pushed).
- [ ] `CURRENT_PROJECT_VERSION` is greater than the last uploaded build number (currently 45 → use **46+**). Apple requires the build number to increase monotonically per `MARKETING_VERSION`.
- [ ] ScarfCore (`swift test`) and `scarf mobile` build green for **Any iOS Device (arm64)**.
- [ ] Capabilities unchanged: Keychain Sharing only; **Push stays OFF** (`NotificationRouter.apnsEnabled = false`).

## Archive + upload

- [ ] Xcode → scheme **`scarf mobile`** → destination **Any iOS Device (arm64)** → Product → **Archive**.
- [ ] Organizer → **Distribute App** → **App Store Connect** → **Upload** (defaults; strip Swift symbols ON).
- [ ] Wait for processing (~5–15 min); App Store Connect emails when the build is ready.

## What to test (paste into TestFlight → What to Test)

```
v2.13.0 — Hermes profile switching + remote-chat reliability.

- Switch profiles: Profiles tab → pick a different profile. Confirm chat,
  memory, cron, sessions, and skills all follow it — and that your host's
  active profile (what the Mac app / terminal use) is UNCHANGED.
- Remote chat & Settings: on a host that previously failed with "Transport
  refused" or showed empty Settings, confirm chat connects and Settings
  populates.
- Skills "What's New": switch profiles and confirm the pill counts reflect
  the new profile (no bogus carry-over from the previous one).

Known limitations: no push notifications; English only.
Report issues via TestFlight feedback.
```

## Submit

- [ ] TestFlight → External Testers → **Public Beta** group → add the new build → **Submit for Review** (Beta Review ~24–48h).
- [ ] On approval, the existing public link `https://testflight.apple.com/join/qCrRpcTz` serves the new build automatically — no new URL.

## Rollback

- [ ] Expire the build in App Store Connect → TestFlight → Builds.
- [ ] Fix, re-archive with a higher `CURRENT_PROJECT_VERSION`, re-upload, re-add to Public Beta.
