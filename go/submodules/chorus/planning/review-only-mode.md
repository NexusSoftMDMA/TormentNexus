# Review-Only Mode (Option A — single pass)

## Problem

Today every Chorus chat assumes Chorus owns the doer seat: a CLI subprocess writes the artifact, reviewers critique it. But many real workflows already have the artifact in hand — a diff produced by `/work`, a draft I just wrote in this conversation, a patch from another tool. Spawning a fresh doer to re-derive it wastes a CLI seat, burns quota, and the re-derived artifact is not the one the caller actually wants reviewed.

## Approach

Add a new phase kind `review_only` that:
- Takes the artifact as **runtime input** at chat-create time (not template config)
- Skips doer spawn entirely; writes the supplied text into the doer-answer-file slot synthetically and marks the doer phase DONE
- Spawns reviewers exactly as today
- **Forces single pass.** `iterate.maxRounds` is ignored — review-only is always 1 round
- Disables Ship phase (no doer = no diff to commit)

Caller workflow: revise the artifact themselves, submit a *new* review-only chat for round 2. Chorus stays stateless about revisions. This matches how PR review tools already behave (push new commit → fresh review pass).

## Alternatives considered

1. **`mode: 'full' | 'review_only'` at template top level.** Rejected — conflates with phase composition (PR 3/3) which is moving toward per-phase source declarations. New phase kind composes cleaner.
2. **Optional doer (infer review-only when missing).** Rejected — implicit, harder for the cockpit picker to surface, and breaks template-portability invariants.
3. **`doer.source: 'spawn' | 'external'` knob inside existing review phase.** Considered. Two-knob shape is more flexible but means UI has to inspect nested fields to decide what to render. Phase-kind discriminator is clearer at every layer (schema validator, runner switch, cockpit picker, MCP tool description).
4. **Multi-pass review-only with caller-revision callback (Option B from chat).** Deferred. Adds chat lifecycle state ("waiting for revision"), a `submit-revision` endpoint, and cockpit affordances. Worth doing only if a UX use-case shows up where the user iterates inside the cockpit.

## Schema changes

### Template YAML — new phase kind

```yaml
phases:
  - id: review
    kind: review_only         # NEW kind (vs existing 'review')
    title: External Review
    description: Reviewers critique an artifact supplied at chat creation.
    # NO doer block — that's the whole point
    reviewer:
      require: 2
      crossLineage: true
      candidates:
        - lineage: openai
          models: [gpt-5.5]
        - lineage: google
          models: [gemini-3.1-pro-preview]
        - lineage: anthropic
          models: [claude-opus-4-7]
    artifact:
      label: "Diff or draft to review"     # cockpit textarea label
      hint: "Paste a unified diff, a markdown draft, or any text blob."
      maxBytes: 1048576                    # 1 MiB cap, guard rail
    # iterate block omitted — runner forces 1 round
```

Schema validator (`src/lib/template-schema.ts`):
- Add `review_only` to the phase-kind union
- For `kind: review_only`: forbid `doer`, forbid `iterate`, require `reviewer` + `artifact`
- For `kind: review`: behavior unchanged

### Top-level template fields

- `ship.enabled` is **ignored** when any phase is `review_only`. Validator warns (does not reject) so templates can theoretically mix kinds in future. Runner short-circuits Ship with a clear log line.

### Chat creation API

`POST /chats` body gains an optional `artifact: string` field. When the resolved template's first review phase is `kind: review_only`, the field is **required** — return 400 if missing or empty. Reject if `artifact.length > phase.artifact.maxBytes`.

### DB

- `chats` table: add `artifact TEXT` column (nullable). Idempotent migration — existing rows unaffected.
- `templates` table: no schema change. The new phase kind is encoded inside the YAML/JSON spec column.
- `phase_events`: no change. Synthetic doer DONE event uses the existing event shape.

## Runner changes

`src/daemon/runner.ts` (or wherever phase dispatch lives):

1. On chat start, inspect first phase. If `kind: review_only`:
   - Skip `precheckLineage` for doer (no doer)
   - Write `chat.artifact` to the doer answer file via `atomicWriteJsonSync`-equivalent for plain text
   - Emit synthetic `phase_event` rows: `doer_started`, `doer_output` (the artifact), `doer_done`
   - Proceed to reviewer spawn exactly as today
2. After reviewers converge, **do not loop**. Regardless of agreement verdict, mark the phase complete and return findings.
3. Ship phase: `if (template.hasReviewOnlyPhase) skipShip('review-only template — no doer diff to commit')`.

The reviewers themselves see no difference — they read an answer file like always.

## UI changes (cockpit)

### Template picker
- Show "Review only" pill next to templates whose first phase is `kind: review_only`
- Sort/group: full-pipeline templates above review-only templates

### Chat creation form
- When picked, replace "Task" textarea with **"Artifact to review"** textarea (taller, monospace, placeholder = template's `artifact.hint`)
- Hide the "Repo path" input (Ship is disabled, repoPath has no effect)
- Submit button label becomes "Send for review"

### Run page
- **No doer card.** Render the artifact as a static "Input" card at top — collapsed by default with a "Show artifact" toggle. Visually distinct from agent cards (no streaming animation, no model badge).
- Round indicators (`Round 1 of 3` etc.) hidden — render only the reviewer cards
- Progress bar: collapse to single segment ("Reviewing…" → "Done")
- Convergence display: drop "Round 1" prefix on findings; just show "Findings"

### Run page — completion state
- Verdict block shows only the agreement summary (`agree / disagree / overridden`)
- No "Revise" or "Continue" button. If user wants another pass, they re-submit a fresh chat (visible CTA: "Submit revised artifact" → opens new-chat dialog pre-filled with current template + last artifact)

## Built-in template

Add `templates/review-only.yaml`:

```yaml
id: review-only
name: Review Only
description: |
  Paste a diff, draft, or text blob. Three reviewers (Codex + Gemini + Claude)
  critique it independently. Single pass — revise yourself and resubmit for
  another round.
author: chorus
agreementThreshold: 0.66
onThresholdMet: ask
yoloDefault: false

ship:
  enabled: false             # explicit, even though runner ignores it for review_only

phases:
  - id: review
    kind: review_only
    title: External Review
    description: Three lineages critique the supplied artifact independently.
    reviewer:
      require: 2
      crossLineage: true
      candidates:
        - lineage: openai
          models: [gpt-5.5]
        - lineage: google
          models: [gemini-3.1-pro-preview]
        - lineage: anthropic
          models: [claude-opus-4-7]
    artifact:
      label: "Artifact to review"
      hint: "Paste a unified diff, a markdown draft, code, or any text blob."
      maxBytes: 1048576
```

Lineage diversity holds (codex + gemini + claude as reviewers, no doer to overlap with).

## MCP entry point

`mcp__chorus__create_chat` tool gets an optional `artifact` parameter. Description updated: *"Required when the chosen template has a `review_only` phase."*

This is the unlock for `/work` and other harnesses to call Chorus as a pure review service.

## CLI entry points (future — captured here so it stays consistent)

Today Chorus has no `run` / task CLI subcommand. Tasks are started via cockpit or MCP only. When a CLI task surface is added, the doer-vs-no-doer split maps cleanly to two subcommands:

```
chorus run "fix the bug in foo.ts"        # full-pipeline templates (doer required)
chorus review --file diff.patch           # review-only templates (artifact, no doer)
chorus review < diff.patch                # same, via stdin (pipe-friendly for /work etc.)
chorus review --template <id> --file ...  # explicit template pick (default = review-only built-in)
```

Routing rules:
- `chorus run` → resolves a template; rejects with clear error if the template has a `review_only` phase ("use `chorus review` for this template")
- `chorus review` → resolves a template; rejects with clear error if the template's first phase is NOT `review_only` ("use `chorus run` for this template")
- `chorus review` requires `--file` OR stdin; reject if both empty; reject if size exceeds the template's `artifact.maxBytes`

The subcommand encodes user intent — `run` means "do work for me", `review` means "critique what I have". Same daemon, same templates table, same chat lifecycle. Only the entry point differs.

**Out of scope for this PR.** The CLI subcommands belong to a follow-up. This PR focuses on the substrate (phase kind + runner + cockpit + MCP). When `chorus run` / `chorus review` get added later, the substrate is ready.

## Test strategy

Unit:
- Template schema accepts `kind: review_only`, rejects `doer` block under it, rejects `iterate` block under it
- Template schema rejects existing `kind: review` if `doer` is missing (regression guard)
- Runner: when first phase is `review_only`, no `precheckLineage` call for doer, no spawn of doer subprocess
- Runner: synthetic doer DONE event emitted with artifact text as output
- Runner: `iterate.maxRounds` ignored — runner exits after round 1 regardless of agreement verdict
- Ship: skipped with logged reason on review-only chats
- Chat-create endpoint: 400 when `artifact` missing/empty for review-only template; 400 when over `maxBytes`

Integration (with mock CLI reviewers):
- End-to-end review-only chat: artifact in → 3 reviewer subprocesses spawned → findings out
- Disagreement verdict in round 1 → chat ends, no round 2 attempted
- Reviewer dies mid-stream → `cli_warning` emitted, partial findings delivered (same as full-mode behavior)

Regression:
- Existing `code-review.yaml` template still produces a doer + reviewers + multi-round loop unchanged
- Existing `red-green.yaml`, `bug-diagnose.yaml`, `architect-review.yaml`, `tri-review.yaml` unchanged
- `tests/db.test.ts` schema-init assertion: add `artifact` column to expected `chats` table shape

UI (manual smoke for v1):
- Cockpit picker shows "Review only" pill on the new template
- New-chat form swaps task textarea for artifact textarea
- Run page hides doer card, hides round indicators, shows artifact as collapsed input card

## Out of scope (explicitly deferred)

- **Multi-pass review-only with callback loop.** That's Option B from the design chat. Revisit if a cockpit UX use-case demands it.
- **Hybrid templates** (some phases `review_only`, some `review` with doer). Phase composition (PR 3/3) territory — let that PR decide the substrate, then this can compose with it.
- **Diff-aware reviewers.** Today reviewers see raw text. A future enhancement could detect unified-diff format and render context-aware findings ("on line 42 of foo.ts you …"). Not blocking v1.
- **Streaming artifacts.** Artifact comes in all at once via the create-chat body. No partial uploads.
- **Re-arm UX.** A "submit revised artifact" button that pre-fills the new-chat dialog from the previous chat — nice-to-have, not blocking.

## Risks

- **Schema migration**: `chats.artifact` column add must be idempotent on existing chorus.db files. Standard `ALTER TABLE … ADD COLUMN IF NOT EXISTS` pattern from libsql migration PR.
- **Cockpit run page already assumes a doer card exists.** Some components may NPE when phase has no doer participant. Need to grep `participants.find(p => p.role === 'doer')` and gate every site on null.
- **MCP tool description drift.** The `create_chat` MCP tool advertises a fixed parameter schema; adding `artifact` as conditionally-required is awkward. Document it as always-optional, validate at runtime, return a clear error message if missing.
- **Quorum math with 3 reviewers and `require: 2`.** Need to confirm `crossLineage: true` still works when no doer is in the lineage pool. Today the doer's lineage is excluded from reviewer candidates — with no doer, all three candidates are eligible. Audit the candidate-resolver function before shipping.

## Order of work

1. Schema validator + DB migration + runner phase dispatch (substrate)
2. Built-in `review-only.yaml` template + seed loader
3. Chat-create endpoint validation + MCP `artifact` parameter
4. Cockpit picker pill + new-chat form swap
5. Run page hide doer card + collapse round indicators
6. Tests at every layer
7. Manual smoke through cockpit + via MCP from this very conversation

Estimated: ~1 day of focused work. Smaller than Phase composition because it doesn't restructure templates — it adds one new phase kind alongside the existing one.
