# Scarf v2.10.2

A small bug-fix release. The headline fix lands a long-running paper cut on remote-host workflows, and the version pin catches Scarf's tested-against Hermes range up to the latest patch.

## ACP permission prompts: "Allow Once" / "Allow For Session" actually work now

When Hermes asked for permission to run a command (typically a `sudo` or any policy-gated tool), tapping **Allow Once** or **Allow For Session** in the Mac permission sheet — and the same buttons in the ScarfGo iOS sheet — was reaching Hermes as if the user had **cancelled** the prompt. Hermes correctly refused to run the command and printed *"blocked from executing"*; the user saw a sheet that responded to taps but had no effect downstream.

Root cause was a wire-format mismatch in Scarf's response to ACP's `session/request_permission`. The Zed-spec `RequestPermissionOutcome` is a union keyed on a literal field named `outcome` with values `"selected"` and `"cancelled"`. Scarf was sending `{kind: "allowed"|"rejected"}` instead — fine-looking JSON, but no spec-compliant discriminator, so Hermes (correctly) fell through to its cancellation default. Every Allow tap rode the same code path as a dismissal.

The fix sends the spec-correct shape: `{outcome: {outcome: "selected", optionId: "<id>"}}` for any picked option (Allow Once, Allow For Session, Reject Once, Reject Always — Hermes reads the user's intent from the optionId, not the discriminator), and `{outcome: {outcome: "cancelled"}}` for a true dismissal. The bug affected both Mac and iOS; the report came in from an iOS TestFlight user running sudo against a remote SSH host where prompts are unavoidable.

Coverage tightened to lock the shape: `M4ACPIOSTests` now decodes the actual response JSON and asserts the discriminator field name + values, so the pre-fix shape can't return via substring-passing test.

## Hermes v0.15.2 compatibility

Scarf's verified-compatibility range is extended to **v0.15.2** (released 2026-05-29), the latest patch on the v0.15 Velocity Release line. v0.15.1's hotfix wave (dashboard 401 reload fix, Docker `--insecure` opt-in, MCP bare-command resolution under Docker, Kanban worker SIGTERM, full skills.sh catalog restored, `/yolo` mid-session, `/model`/`hermes model` parity) and v0.15.2's packaging fix are transparently back-compat — no Scarf-side gating changes were needed, but the test suite now exercises a v0.15.x parse line and asserts that all 14 v0.15 capability flags stay active across patch releases.

Users running Hermes in Docker should re-read v0.15.1's release notes for the `HERMES_DASHBOARD_INSECURE=1` opt-in — that's an explicit Hermes server-side change unrelated to Scarf.

## Upgrade notes

- Sparkle will offer this update automatically on next launch.
- macOS 14.6+ (Sonoma) deployment target unchanged.
- iOS testers: this build is Mac-only. A ScarfGo TestFlight build carrying the same ACP permission fix is queued separately.
