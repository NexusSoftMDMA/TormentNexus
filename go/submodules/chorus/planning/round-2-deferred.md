# Round-2 Deferred Items

Smaller robustness improvements raised during the PR #3 round-2 review and follow-up
design discussions. None blocking. Each is roughly half a day. Listed in rough
priority order — pick from the top when warmup work is wanted.

## 1 — Completion-detection: structured terminal events over text sentinels

**Problem.** Today the runner watches each agent's answer file for a `## DONE` text
sentinel injected by prompt. Failure modes:
- Model forgets to emit it (Gemini, Kimi do this regularly)
- Model writes `## DONE` inside a code block as an example → false positive cuts off real output
- Format drift (`## Done`, `**DONE**`, `[DONE]`) misses strict regex
- Structurally fragile: the prompt that enforces it is the prompt the model is allowed to ignore

**What the proxy repos do.** CLIProxyAPI / OmniRoute / 9router never use a text
sentinel. They watch the wire-level SSE terminal event from the upstream API:
- OpenAI-shaped: `data: [DONE]`
- Anthropic-shaped: `event: message_stop`
- Gemini-shaped: terminal chunk with `finishReason: "STOP"`

Lesson: **prefer structured terminal signals from the source over text patterns in the
rendered output.** The model can't forget structured terminal events, can't hallucinate
them mid-stream, can't muddle the formatting.

**Approach.** Per-lineage terminal-event parser, with a hierarchy:
1. **Stream-JSON terminal event** (claude `result`, codex assistant-done) — preferred when available
2. **Subprocess clean exit + stdout EOF** — universal backstop
3. **`## DONE` text sentinel** — explicit fallback for kimi + opencode (weak/absent stream-json)
4. **Idle timeout** — stall detection (NOT completion). Distinguishes "thinking" from "hung".

Isolate per-lineage detection in one parser module so the runner sees a single
normalised "stream complete" event regardless of source.

**Wins.**
- Kills the "Gemini ate the sentinel" tail-of-output bug class
- Catches mid-stream sentinel hallucination
- Cleaner state machine: terminal event = done, idle timeout = stuck, exit-without-terminal = crashed-mid-stream — three distinct states instead of one fuzzy one
- Sets up format translation if an HTTP-proxied lineage (OpenRouter inline) is added later

**Pairs with.** The existing `error-detector` taxonomy — same shape, per-lineage
adapters feeding a normalised internal event.

**Bundle: token usage capture.** The same terminal event carries token-usage data.
Capturing it costs nothing extra at parse time and unlocks card-level display + cost
calc later.

Per-lineage shapes (from stream-json terminal events):

| Lineage | Terminal event | Token fields |
|---|---|---|
| Claude (anthropic) | `{"type":"result", "usage":{...}}` | `input_tokens`, `output_tokens`, `cache_creation_input_tokens`, `cache_read_input_tokens` |
| Codex (openai) | `{"type":"assistant_done", ...}` | `prompt_tokens`, `completion_tokens`, `total_tokens` |
| Gemini (google) | terminal chunk | `promptTokenCount`, `candidatesTokenCount`, `totalTokenCount` (when present) |
| Kimi / OpenCode | text sentinel fallback | none — record `null`, no shadow cost |

Normalise into a single internal shape stored on the `phase_event` row (or a new
`phase_tokens` row keyed by `(chat_id, phase, participant_id)`):

```ts
{
  inputTokens: number | null,
  outputTokens: number | null,
  cachedInputTokens: number | null,    // anthropic-only today; null elsewhere
  totalTokens: number | null,           // computed if missing: input + output
  costUsdActual: number | null,         // 0 when CLI is on subscription plan
  costUsdShadow: number | null,         // list-price equivalent for capacity / ROI math
  costSource: 'subscription' | 'metered' | 'unknown'
}
```

Cost is computed from `voices.input_cost_per_mtok` / `output_cost_per_mtok` (already
shipped in the voices table from PR #2). Subscription-backed voices set
`costUsdActual = 0` and `costUsdShadow = list-price-equivalent` per the existing
`feedback_cost_actual_vs_shadow_split.md` rule.

**UI surfaces.**
- Participant cards in the run page get a small bottom-row stat: `↑12.3k ↓4.5k · $0.08`
- Chat header gets a totals badge: total tokens + cost-actual + cost-shadow
- Voices table can later show "tokens used last 7d" / "cost last 30d" rollups

**Storage decision.** New `phase_tokens` table (vs adding to `phase_event`) — token
data is one-per-participant-per-phase, not one-per-event. Keeps `phase_event` lean and
lets us aggregate without scanning event rows. Idempotent migration; nullable columns;
existing `chorus.db` files unaffected.

**Out of scope for this PR.** Cost-rollup queries, voices-page cost columns, and
spend dashboards. We're capturing the data here; surfacing aggregates is a follow-up
once a few weeks of dogfood data exists.

**Estimate revision.** ½ day for completion detection alone → **1 day** with token
capture bundled. Same parser, but adds the new table + migration + card UI bits.

---

## 2 — Per-phase timeout override in template schema

**Problem.** Today the spawn timeout is a single global value. A long-running architect
phase (which legitimately takes 5+ minutes for big-picture analysis) hits the same
timeout as a quick code-review phase (which should fail fast at 60s if a CLI hangs).
Setting the global high makes hung CLIs sit longer than needed; setting it low kills
legit slow phases.

**Approach.** Optional `timeoutMs` on phase definitions in template YAML. Runner reads
`phase.timeoutMs ?? settings.defaultPhaseTimeoutMs` per phase. Schema validator enforces
sensible bounds (30s ≤ x ≤ 1h).

```yaml
phases:
  - id: review
    kind: review
    timeoutMs: 600000   # 10 min for this slow phase
    ...
```

**Wins.**
- Templates can self-tune per workload
- Hung-CLI detection stays tight on fast phases
- No change to existing templates (field optional, default falls through)

**Risks.**
- Need to make sure timeoutMs is passed to BOTH the doer subprocess and each reviewer
  subprocess in that phase (not just the phase itself). Today subprocess spawn has its
  own timeout knob — these need to align.

**Estimate.** ½ day.

---

## 3 — Structured logging with request IDs

**Problem.** Daemon log today is line-based plain text. Tracing a single chat through
its lifecycle (chat-create → precheck → spawn N CLIs → stream events → converge → done)
means grepping for chat ID and reading interleaved lines from N concurrent chats.
Fine at 1 chat/sec, painful at 10.

**Approach.** Adopt a structured logger (pino is the obvious choice — already used in
the Node ecosystem, fast, JSON-line output). Every log line carries:
- `chatId` — primary correlation key
- `phase` (`precheck` | `spawn` | `stream` | `converge` | `ship`)
- `participantId` — which CLI seat this log relates to (when applicable)
- `lineage` — which CLI family
- `requestId` — for HTTP-shaped operations (chat-create, /voices CRUD)

Output goes to stdout as JSON lines. Cockpit log viewer reads JSON, renders pretty
with filter chips. Standalone CLI users get a `chorus logs --chat <id>` that filters
the JSON stream by chatId.

**Wins.**
- One-grep tracing per chat: `chorus logs --chat abc123` → entire timeline in order
- Cockpit can render filtered views without reparsing free-text
- Alerts/metrics pipelines can consume structured fields directly
- Foundations for OpenTelemetry export later (request IDs are the seed of trace IDs)

**Risks.**
- Migration: every existing `console.log`/`console.error` site needs to move to the
  structured logger. Mechanical refactor, ~50 sites. Low risk per site, tedious in aggregate.
- Log volume: JSON lines are ~3× plain-text bytes. Worth it but noting.

**Estimate.** ½ day for the substrate + first-class call sites; another ½ day for the
mechanical migration of remaining call sites. Can be split into two PRs.

---

## 4 — Opt-out telemetry (install heartbeat)

**Problem.** Today Chorus has zero usage signal. We don't know how many installs are
active, what versions people are running, what OS mix the user base has, or whether
v0.7 fixed the install regressions v0.6 caught. Decisions about deprecation, platform
support, and where to invest are flying blind. npm download counts give a rough
adoption baseline but say nothing about whether the daemon ever boots.

**Why opt-out, not silent.** Chorus's audience is devs who tcpdump for fun. A silent
phone-home — even something as innocuous as a logo fetch — gets discovered, posted to
Reddit, and the credibility hit outlasts any signal value (Audacity / Sentry / npm
have all paid this tax). Explicit + disclosed + easy-to-disable is the only shape that
trades cleanly for a project still building reputation.

**Approach.** Daemon-side heartbeat (not cockpit-side — more reliable, fewer pings
per user, doesn't depend on browser activity).

```
Endpoint:  POST https://chorus.codes/api/telemetry
Frequency: Once per daemon boot + once per 24h while running
Transport: HTTPS, payload <500 bytes
```

**Anonymous install ID.** Generated once via `crypto.randomUUID()` on first daemon
boot, persisted in `~/.chorus/install-id`. Not derived from anything machine-specific
(no MAC, no hostname). User can `rm` it to reset; daemon will mint a new one next boot.

**Payload (exhaustive — what we send):**

```json
{
  "schema": 1,
  "installId": "8c1a3f4e-...",
  "version": "0.7.0",
  "os": "linux",
  "arch": "x64",
  "node": "22",
  "daemonUptimeSeconds": 86400,
  "chatsLast24h": 12
}
```

**What we never send:**
- Chat content, prompts, artifacts, file paths
- Repo paths, branch names, commit SHAs
- Hostnames, usernames, IPs (the server logs the IP for rate-limiting only,
  immediately discards after request)
- API keys, model IDs from BYO endpoints, anything from `~/.chorus/dx_secret`-equivalent stores
- Voice / template / persona names — which CLIs and templates a user picks is
  behavioural data, not infrastructure data. Out.
- Per-chat detail beyond the aggregate count

**Opt-out paths (any one disables):**

```bash
# Env var (immediate, persistent across boots)
export CHORUS_TELEMETRY=0

# Settings flag (UI toggle in cockpit Settings page)
chorus settings set telemetry.enabled false

# Touch-file (matches OSS conventions: cargo, brew, etc.)
touch ~/.chorus/no-telemetry
```

Daemon checks all three on every send; any one disables. No retry queue — skipped
heartbeats are skipped, no buffer-and-replay.

**Disclosure (must ship together with the telemetry, not after):**
- README "Telemetry" section with the full payload list + opt-out instructions
- First-run notice on `chorus init`:
  ```
  Telemetry: Chorus pings chorus.codes once a day with version + OS + chat
  count (no chat content). Disable with CHORUS_TELEMETRY=0 or `chorus settings
  set telemetry.enabled false`. Full details: chorus.codes/privacy
  ```
- Cockpit Settings page: toggle with same disclosure inline
- The endpoint itself returns the disclosure JSON when GET'd, so users can curl
  it to see exactly what's being collected
- **Website (chorus.codes):**
  - New `chorus.codes/privacy` page — full payload spec, opt-out instructions,
    retention policy, contact email. Single source of truth that all other
    surfaces link to.
  - Footer link from homepage + docs to `/privacy` (small, but visible)
  - FAQ entry: *"Does Chorus collect any data?"* → short answer + link
  - Endpoint URL (`/api/telemetry`) deployed alongside the website

**Legal — terms of service vs privacy notice.** A privacy notice is required (we're
collecting a pseudonymous identifier + activity signal, even if minimal). A full
Terms of Service is **not** required for an OSS CLI distributed under MIT/Apache —
the license already governs the use grant. Decision: ship a privacy notice page only.
Revisit ToS only if/when paid hosted Chorus features ship (account systems, billing,
SaaS surface) — that would be a future event, not blocking this PR.

The privacy notice should cover, in plain language:
- Exactly what's sent (link to the JSON payload spec)
- What's never sent (the "never send" list above)
- Where data goes (chorus.codes server, no third parties, no ad networks)
- Retention (e.g. "raw heartbeats: 30 days; aggregate counts: indefinite")
- Opt-out (three paths, copy-paste commands)
- Contact for data deletion requests (since installId is pseudonymous, deletion =
  user provides their installId from `~/.chorus/install-id` and we purge matching rows)
- Version + last-updated date on the page itself, so future changes are auditable

**Server side.** Tiny endpoint behind chorus.codes — Cloudflare Worker or Vercel
function appending to an analytics store (D1, BigQuery, or a flat log file with daily
rotation). No personal data persisted. Aggregate dashboard for project decisions only;
not exposed publicly. Out of scope for this plan — covered in a separate
infra/deployment task.

**Wins.**
- Real install adoption signal (per-version, per-OS, per-day)
- Engagement signal (`chatsLast24h` distinguishes "installed once and forgot" from
  "actively used")
- Regression detection — if v0.7.1 ships and `chatsLast24h` median drops, something
  broke

**Risks.**
- **Trust.** Mitigated by opt-out + disclosure + minimal payload + auditable endpoint.
  But a bad first impression on Hacker News could still sting; the README disclosure
  needs to be the first or second thing in the file.
- **Endpoint as availability dependency.** If chorus.codes is down, telemetry POSTs
  shouldn't block the daemon. Use fire-and-forget with a 5s timeout; failure is logged
  at debug level only and never surfaced to the user.
- **GDPR / privacy regs.** InstallId is technically a pseudonymous identifier. The
  privacy notice + opt-out covers the consent angle for personal use; for enterprise
  users behind firewalls the network-layer block already opts them out implicitly. No
  data is sold or shared with third parties.
- **Schema evolution.** Bake `schema: 1` in from day one so future fields are additive.
  Server tolerates unknown fields; older clients keep working when the schema bumps.

**Estimate.** ½ day for the daemon-side heartbeat + opt-out wiring + disclosure copy.
Server endpoint deploy is separate (~½ day, tracked in infra).

---

## How to pick

These are all warmup-shaped — small blast radius, isolated, no schema migration
(except #4's `~/.chorus/install-id` file).

- **Pick #1 (completion detection + token capture)** if hung-stream / cut-off-output
  bugs come up in dogfood, or if cost-per-chat visibility becomes interesting
- **Pick #2 (per-phase timeout)** if a template author needs different timeout
  budgets in different phases
- **Pick #3 (structured logging)** if debugging multi-chat sessions becomes the bottleneck
- **Pick #4 (opt-out telemetry)** when the project is ready for real adoption
  signal — typically right before a public push (npm publish, HN post, etc.) so the
  numbers from launch onward are captured
- **Pick all four before Phase composition** if you want a buffer of robustness +
  observability work before touching the template schema in a bigger way

Each has its own PR. They don't depend on each other.
