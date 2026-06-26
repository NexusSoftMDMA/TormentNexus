# A/B Test: Naming Audit Task — JCodeMunch vs Native Tools

**Date:** 2026-03-18
**Author:** @Mharbulous
**Source:** [GitHub Issue #128](https://github.com/jgravelle/jcodemunch-mcp/issues/128)
**Iterations:** 50 (25 per variant)
**Model:** Claude Sonnet 4.6 (both variants)
**Codebase:** Vue 3 + Vite + Vuetify 3 + Firebase + Cloud Functions
**Platform:** Windows 11 (MINGW64), Claude Code CLI
**Timeout:** 600s per iteration

---

## Task Design

Each iteration was a fresh Sonnet 4.6 session that scanned source files for misleading, ambiguous, or inconsistent names, then applied fixes via a three-subagent consensus vote (Haiku/Sonnet/Opus). Variants alternated to control for session order effects.

- **Variant A (Native):** Grep / Glob / Read
- **Variant B (JCodeMunch):** `get_file_outline`, `get_symbol`, `find_importers`, `search_symbols`

---

## Results

### Top-line

| Metric | Variant A (Native) | Variant B (JCodeMunch) | Delta |
|--------|-------------------|----------------------|-------|
| Successful iterations | 18/25 (72%) | 20/25 (80%) | +8 pp |
| Timeouts | 10/25 (40%) | 8/25 (32%) | −8 pp |
| Mean cost/iteration | $0.783 | $0.738 | −5.7% |
| Mean output tokens | 12,230 | 11,347 | −7.2% |
| Mean cache creation | 104,135 | 93,178 | −10.5% |
| Mean duration (s) | 318 | 299 | −6.0% |

### Isolating tool-layer savings

The blended 5.7% cost reduction understates the actual tool-layer advantage. Each iteration includes variant-independent fixed overhead: three subagent consensus calls (Haiku/Sonnet/Opus) per finding, plus system prompt and skill loading costs. This constant overhead dilutes the A-vs-B comparison.

Isolating iterations with 0 findings (no subagent overhead) and matching by file count:

| File count bucket | A avg cost | B avg cost | B savings |
|-------------------|------------|------------|-----------|
| 2–3 files (n=4 vs 5) | $0.400 | $0.329 | **17.7%** |
| 4–5 files (n=1 vs 3) | $0.394 | $0.365 | **7.1%** |
| 6–8 files (n=2 vs 2, avg 7.0 both) | $0.739 | $0.562 | **23.9%** |

**Estimated tool-layer savings: 15–25%.** The cost mechanism is cache creation — native `Read` returns entire raw files as prompt content, while `get_file_outline` and `get_symbol` return smaller structured payloads.

---

## Qualitative findings

### `find_importers` — structural capability, not just speed

B-004 detected two orphaned files (`peek-event-handlers.js`, `peek-positioning.js`) with zero live importers. This finding category did not appear in any Variant A iteration. Native tools cannot answer "what imports this file?" without scripting; `find_importers` makes it a single tool call.

### `get_symbol` — focused context changes reasoning depth

B-016 did not just rename `isLawyerReversed`. It recognized the computed property was a double-inversion and eliminated 26 lines of code. The hypothesis: `get_symbol` returning a focused implementation (rather than an entire file) prompted deeper reasoning about what the code was actually doing.

### `get_file_outline` — pre-parsed symbol lists

Pre-parsed symbol lists meant the agent did not need to read entire files for name extraction. This is the likely primary mechanism behind the cache creation savings.

### Vue 3 / `<script setup>` — no blind spots

Symbol extraction from `<script setup>` blocks, computed properties, and component APIs worked correctly across all 20 successful Variant B iterations. No index blind spots, failed queries, or tool limitations affected outcomes.

---

## No tool gaps found

This test did not reveal any JCodeMunch-specific gaps. Finding quality was equivalent across variants on shared finding categories. The only finding category unique to one variant was orphaned file detection (Variant B only, via `find_importers`).

---

## Relationship to synthetic benchmark

The [synthetic token benchmark](results.md) measures token reduction on pure retrieval tasks (symbol search + fetch) against a "read all files" baseline, showing 95%+ reduction. This A/B test measures end-to-end task cost on a real agentic workflow that includes fixed overhead independent of the tool choice. Both are valid measurements of different things:

- **Synthetic benchmark:** upper bound on retrieval-layer savings (no fixed overhead)
- **This A/B test:** real-world blended savings on a realistic task (15–25% tool-layer isolated, 5.7% blended)
