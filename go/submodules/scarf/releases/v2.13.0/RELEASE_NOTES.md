# Scarf v2.13.0

ScarfGo — the iOS companion — gets its headline feature this release: **switch between your Hermes profiles right from your phone**, without disturbing the profile your Mac, terminal, or running gateway is using on that host. It ships alongside a **reliability fix for remote chat and Settings on iOS** that anyone running ScarfGo against an SSH host will feel immediately — connections no longer churn under load. These are iOS-track changes; the shared-core pieces ride into the Mac app too.

## Switch Hermes profiles from ScarfGo

If you run more than one Hermes profile on a host — say an `admin` profile and a locked-down `gateway` profile — ScarfGo could *list* them but not operate against them; switching was Mac-only. Now the Profiles tab is a **switcher**: pick a profile and ScarfGo points its chat, memory, cron, sessions, gateway, and every `hermes` call at that profile's `HERMES_HOME` for the selected server.

Crucially, this uses **per-connection scoping (Design B)**: ScarfGo scopes *its own* view via `HERMES_HOME` / `remoteHome` and **never runs `hermes profile use`**, so it does **not** touch the host's `active_profile`. Your Mac app, terminal, cron daemon, and the running gateway on that host keep using whatever profile they were on — switching on the phone is invisible to them. (Verified on a live host: a profile switch leaves `~/.hermes/active_profile` unchanged.) Profile create / rename / delete / import / export remain Mac-only by design.

## Reliable remote chat & Settings on iOS

Some ScarfGo users hit **"Couldn't save model.provider to config.yaml via hermes config set — Transport refused the command,"** with Settings showing the model and config source empty even though the config was valid and the same command worked over plain `ssh`. It reproduced across unrelated hosts — which pointed at ScarfGo, not the host.

Root cause: ScarfGo opened a **brand-new SSH connection for every file read and CLI call**. A Settings load or chat pre-flight fans out a burst of those, and under the churn the SSH handshake itself starts failing — surfaced as "Transport refused" on writes and silently swallowed to "empty" on reads. A single manual `ssh` never reproduced it because it's one connection.

ScarfGo now **pools one SSH connection per server and coalesces concurrent opens**, so reads and writes reuse one warm connection — the iOS equivalent of the Mac app's SSH ControlMaster. Verified against a real sshd: a 24-way concurrent burst went from 20-of-24 connection failures to **zero**. The chat pre-flight also now surfaces the real connection error instead of a generic line. (This is separate from the Docker `~/.hermes`-in-container read path, which is still tracked.)

## Under the hood

- **Skills "What's New" now tracks each profile independently.** After per-connection switching landed, the Skills pill could show bogus "new / changed" counts by diffing one profile's skills against another profile's baseline (on iOS it bled across servers too). The last-seen snapshot is now keyed per (server, profile); the skills list itself was always correct — only the diff indicator was wrong.
- The ScarfIOS package is now unit-testable on macOS (`swift test`), and a gated live integration test drives the real Citadel SSH transport through the pool against an ephemeral sshd to guard the churn fix against regressions.
- iOS direct-file reads now honor the selected profile (previously they ignored profiles and always read the host default), closing a latent half-switch.

## Upgrade notes

- **iOS / ScarfGo:** these are iOS-track features — a ScarfGo TestFlight build carrying them is queued on the iOS track (independent of the Mac update path).
- **Mac:** the shared-core changes (profile scoping in ScarfCore) ride into the Mac app; Sparkle will offer the update automatically (or **Scarf → Check for Updates**). macOS 14.6+ deployment target unchanged. No data migrations.
