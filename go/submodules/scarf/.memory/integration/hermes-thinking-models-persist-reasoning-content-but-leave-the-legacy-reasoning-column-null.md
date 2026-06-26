---
title: Hermes thinking models persist reasoning_content but leave the legacy reasoning column NULL
type: note
permalink: scarf/integration/hermes-thinking-models-persist-reasoning-content-but-leave-the-legacy-reasoning-column-null
tags:
- hermes
- reasoning
- state-db
- thinking-models
- gh-aud27
---

The `messages` table in Hermes' `state.db` has BOTH a `reasoning` (v0.7) and a `reasoning_content` (v0.11) column, but for v0.16 thinking models only `reasoning_content` is populated on the live write path. Verified against the real source at `~/.hermes/hermes-agent/` (build 2026.6.5).

## Observations

- [write-path] `agent/chat_completion_helpers.py` accumulates streaming deltas as `getattr(delta, "reasoning_content", None) or getattr(delta, "reasoning", None)` and stores the result ONLY as `msg["reasoning_content"]`. `hermes_state.py add_message` writes both columns from `msg.get("reasoning")` / `msg.get("reasoning_content")`, but `msg["reasoning"]` is never set in the live path — it's populated only on DB-restore (`hermes_state.py:2903`). #hermes
- [consequence] On chat resume, thinking-model assistant rows have a non-empty `reasoning_content` and a NULL `reasoning`. Any Scarf logic that decides "does this message have reasoning?" from the lightweight `reasoning` column alone will miss them. This is the root of t-aud27. #scarf
- [scarf-fix] Scarf's light/skeleton message fetch deliberately excludes the heavy `reasoning_content` blob (issue #74) and carries only `reasoning`. To still show the REASONING disclosure on resume, the fetch now also selects a cheap `(reasoning_content IS NOT NULL AND reasoning_content != '') AS hasReasoningContent` boolean → `HermesMessage.reasoningContentAvailable` → `hasReasoning`. The blob lazy-loads on disclosure-open (t-aud21). #scarf
- [column-caution] `HermesDataService.messageFromRow` reads positionally: `reasoning` at index 10, `reasoning_content` at index 11 (both schema-gated). When adding columns to light/skeleton, keep index 11 == reasoning_content (use `NULL AS reasoning_content` as a placeholder) and read NEW columns BY NAME (`row["hasReasoningContent"]`) to stay order-safe. This is the t-aud01 fragility zone — lock changes with `HermesDataServiceBackendTests` SQL-shape + row-parse tests. #gotcha

## Relations

- relates_to [[hermes-acp-image-handling-scarf-s-wire-shape-is-correct-ignored-images-are-model-vision-routing]]
- relates_to [[Scarf Architecture Rules]]
