# Scarf v2.12.0

A coordinated catch-up to **Hermes v0.17.0 (2026.6.19)** — by far the largest Hermes release yet, though Scarf only needed a focused slice of it — bundled with a **remote-chat performance fix** that anyone running Scarf against an SSH host will feel immediately. The audit of v0.17 also turned up several pre-existing bugs in Scarf's own CLI plumbing, which are fixed here. Every new v0.17 surface is capability-gated, so pre-v0.17 hosts render byte-identical to v2.11.0, and all flag/config/wire shapes were verified against the live Hermes v0.17 source before implementation.

## Remote chat is responsive again

If you chat against a remote Hermes over SSH, typing into an active session could lag and spike CPU. Root cause: during a live chat, every persisted message bumps `state.db-wal`'s modification time, which fires Scarf's file watcher — and several view models were doing **synchronous** SSH/scp reads **on the main thread** on every tick. The biggest offender was the chat's credential preflight (reading `.env` + `auth.json` over SSH per tick), alongside the sessions-list, platforms, and projects refreshes. A ScarfMon capture showed `loadRecentSessions` taking 2.7–4.2 s while the chat render itself measured ~0 ms — the lag was main-thread I/O contention, not rendering.

Every watcher-driven read now runs off the main actor (debounced where appropriate), with cancel-prior + recency guards so a slower older read can't clobber fresher state. Two adversarial review passes verified the concurrency ordering. (Split from the original Dashboard CPU spin fixed in v2.10.3; thanks to the reporter whose ScarfMon dump localized it.)

## Four broken Health/Settings actions, fixed

Auditing v0.17 surfaced four Scarf actions that had been quietly broken — in some cases since they shipped — because they invoked the `hermes` CLI with the wrong verb or flags:

- **Run supply-chain audit** called a bare `hermes audit`, which isn't a command — it was being interpreted as a chat prompt and burning an agent turn instead of running the scan. It now calls `hermes security audit`.
- **Migrate retired xAI model** ran `hermes migrate xai`, which is dry-run by default — so it printed a plan, wrote nothing, and still reported success while the retired-model warning never cleared. It now passes `--apply` (and no longer claims success on a no-op).
- **Set up browser tools** passed `--assume-yes`, which the CLI rejects; the correct flag is `--yes`.
- **WhatsApp recipient allowlist** wrote `whatsapp.allowed_chats`, a key Hermes never reads (WhatsApp gates senders via `allow_from`), so the allowlist silently did nothing. It's been removed from the chat-id editor rather than left lying. Saving a setting that fails now surfaces the real reason (e.g. an administrator-managed key) instead of a generic "Failed to save."

## Curator "Prune" is now "Archive idle skills" — and it actually works

Scarf's Curator screen had a destructive **"Prune Archived"** action that claimed to *permanently delete archived skills from disk*. It turns out no such Hermes command exists — `hermes curator prune` does the opposite: it **bulk-archives _active_ skills that have been idle for N days**, and archiving is fully reversible via Restore. On top of the wrong mental model, Scarf invoked it with a flag that doesn't exist and without the confirmation-skip flag, so a real run would hang until it timed out.

It's been rebuilt to match reality: a **"Archive idle skills"** action in the Curator menu with a 30/60/90/180-day threshold, a preview of which idle skills would be archived (and how long they've been idle), and a confirm sheet that's correctly framed as reversible — not a red "permanent delete" gate. The CLI invocation and output parsing now match the real verb.

## New gateway platforms: WhatsApp Cloud + SimpleX

- **WhatsApp Business Cloud API** (Meta's hosted webhook path — no bridge process) joins the platform list with a full setup form: phone number ID, access token, webhook verify token + app secret, and a direct-message allowlist. It's distinct from the existing QR-paired WhatsApp web bridge.
- **SimpleX Chat** finally has a setup form (it had been in the platform list since v0.14 with no way to configure it): daemon WebSocket URL, contact/group allowlists, auto-accept, and the home channel.

## More v0.17 surfaces

- **Telegram rich messages** (Bot API 10.1, on by default) and an **online/offline status indicator** (opt-in) are exposed as toggles.
- **Curator consolidation is now opt-in.** v0.17 turned the curator's LLM skill-merge pass off by default; Settings → Advanced now has a toggle (with a note) so anyone who relied on the automatic pass can turn it back on. Deterministic pruning is unaffected.
- **Max concurrent sessions** — the new optional cap on simultaneously-active chat sessions is exposed as a stepper (0 = unlimited).

## Under the hood

- **Verified against the live v0.17 source.** The headline finding of the audit: v0.17 required **zero** mandatory compatibility changes — the `state.db` schema, ACP wire protocol, CLI verbs, config keys, and model catalog are all stable — so this release is a focused feature catch-up plus the pre-existing-bug fixes above, not a forced migration.
- **New capability flags** (`hasCuratorConsolidate`, `hasMaxConcurrentSessions`, `hasTelegramRichMessages`, `hasPhotonPlatform`, `hasWhatsAppCloudPlatform`) plus the `isV017OrLater` predicate gate every new surface; an adversarial fresh-eyes audit of the whole branch caught a real bug before release (WhatsApp Cloud is a built-in platform, so it needed an explicit `enabled` write to actually start).
- **642 ScarfCore tests** pass, including a full v0.17 capability cluster (flags-on / pre-v0.17-degradation / patch-still-on) and the v0.16 cluster that the prior cycle had shipped without.
- **Deferred** (tracked for later): iMessage via Photon (the underlying protocol is still changing upstream), the per-MCP-server keepalive knob, and a pre-existing Google Chat allowlist fix.

## Upgrade notes

- Sparkle will offer this update automatically on next launch (or **Scarf → Check for Updates**).
- macOS 14.6+ (Sonoma) deployment target unchanged.
- **Fully backward compatible.** Every v0.17 surface is capability-gated; on Hermes v0.16.x or earlier, Scarf renders exactly as v2.11.0 did. There are no data migrations.
- **iOS testers:** the shared-core changes (the curator/config/platform model updates and the off-main read fixes) ride in ScarfCore; a ScarfGo TestFlight build carrying them is queued separately on the iOS track.
