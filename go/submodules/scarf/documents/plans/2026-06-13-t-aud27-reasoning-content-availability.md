# Plan — t-aud27: Show the REASONING disclosure for reasoning_content-only messages

> Status: NOT STARTED · Created 2026-06-13 · Owner: next session · Risk: MEDIUM
> (touches the message column-index mapping fixed in t-aud01) · Source: t-aud21

## Problem

t-aud21 wired lazy-loading of the rich v0.11 `reasoning_content` when the user opens the
REASONING disclosure. But the disclosure only renders when
`HermesMessage.hasReasoning` is true, which (after t-aud01) is driven by the lighter
`reasoning` channel that the skeleton/light fetch carries. So for a message that has
**only `reasoning_content` and no `reasoning`**, `hasReasoning` is false → the disclosure
is hidden → t-aud21's lazy-load can never trigger. Those messages show no reasoning at all
on resume.

**Open question (resolve first):** do Hermes v0.11+ thinking models populate the legacy
`reasoning` column too, or *only* `reasoning_content`? If they always populate `reasoning`,
this ticket is low-value (t-aud21 already covers it). If they write only `reasoning_content`,
this is the real fix for v0.11 reasoning visibility. **Confirm against a live v0.11 host /
a real `state.db` before building** (inspect a thinking-model session's rows).

## Goal

Make the disclosure appear whenever `reasoning_content` exists, even when not yet loaded,
so t-aud21's on-open lazy fetch can populate it — without shipping the heavy blob in the
bulk fetch (issue #74).

## Approach

1. Add a CHEAP availability boolean to the message fetch — NOT the blob:
   `reasoning_content IS NOT NULL AND reasoning_content != '' AS hasReasoningContent`
   in `messageColumnsLight` and `messageColumnsSkeleton`
   (`HermesDataService.swift`). It's a tiny int, not the 20KB+ text.
2. Add `reasoningContentAvailable: Bool` (default false) to `HermesMessage`; read it in
   `messageFromRow`. **CAREFUL with column indices** — t-aud01 was exactly about the
   skeleton/light/full column-shape mismatch. Append the new column LAST in each SELECT and
   read it by the matching index, relying on `Row`'s bounds-safe subscript (returns `.null`
   when absent) for the full-column path. Add/extend the `HermesDataServiceBackendTests`
   SQL-shape tests (the t-aud01 pattern) to lock the new column in.
3. `HermesMessage.hasReasoning` → also true when `reasoningContentAvailable`.
4. No UI change needed beyond t-aud21: once the disclosure shows, the existing
   `.onChange(of: isExpanded)` lazy-load (`hasFullContent == false` →
   `RichChatViewModel.reasoningContent(for:)`) fetches it. Verify `hasFullContent` is keyed
   off the loaded `reasoningContent` (empty) so it still fetches.

## Affected files

- `scarf/Packages/ScarfCore/Sources/ScarfCore/Services/HermesDataService.swift`
  (`messageColumnsLight`, `messageColumnsSkeleton`, `messageFromRow`).
- `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesMessage.swift`
  (`reasoningContentAvailable`, `hasReasoning`).
- `scarf/Packages/ScarfCore/Tests/ScarfCoreTests/HermesDataServiceBackendTests.swift`
  (SQL-shape + row-parse regression tests).

## Risks / gotchas

- Column-index mapping is the danger zone (see t-aud01). Add columns at the END of each
  SELECT and unit-test the exact SQL + a parsed-row case for skeleton/light/full.
- If the answer to the open question is "models always populate `reasoning`," skip the
  build and just close the ticket as not-needed.

## Verification

- New `HermesDataServiceBackendTests`: skeleton + light SELECT include
  `hasReasoningContent` (not the blob); full still has `reasoning_content`; a parsed row
  with reasoning_content-only yields `hasReasoning == true`, `reasoningContent == nil`,
  `reasoningContentAvailable == true`.
- ScarfCore `swift test` green.
- Manual (needs a v0.11 host): resume a thinking-model chat whose messages have only
  `reasoning_content`; the REASONING disclosure appears and, on open, loads the content.
