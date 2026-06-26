# Security Policy

## Supported versions

Chorus is pre-1.0 and ships from `main`. We patch the latest minor
release; older versions don't receive backports.

| Version | Supported |
|---|---|
| 0.7.x | ✅ |
| < 0.7 | ❌ |

## Reporting a vulnerability

**Please do not open a public GitHub issue for security reports.**

Email **security@chorus.codes** with:

- A description of the issue and impact (data exposure, code execution,
  privilege escalation, etc.)
- Steps to reproduce, ideally with a minimal repro
- Affected versions
- Your contact details if you'd like credit

We aim to acknowledge within 72 hours and to ship a fix within 14 days
for high-severity issues. We'll coordinate disclosure timing with you.

## Threat model (in scope)

Chorus runs as a local daemon that spawns AI CLI subprocesses on the
user's machine. The threat surface includes:

- Daemon HTTP routes (loopback only; no auth in 0.7 — see
  [#issues](https://github.com/chorus-codes/chorus/issues) for the
  hardening roadmap)
- Manual CLI path validation (`/onboard/save-cli-path`)
- Reviewer subprocess sandboxing (per-CLI sandbox profile)
- Local SQLite DB at `~/.chorus/chorus.db` (stores API keys)
- MCP stdio JSON-RPC entry point

Out of scope: vulnerabilities in upstream CLIs (claude / codex / gemini /
opencode / kimi) — please report those to the respective vendors.

## Safe-harbor

We won't take legal action against good-faith research that complies
with this policy. If in doubt, ask before testing.
