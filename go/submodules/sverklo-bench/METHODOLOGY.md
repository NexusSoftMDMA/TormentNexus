# Methodology

How `bench:primitives` measures code-intelligence retrieval, and the load-bearing decisions that produced the current task set.

## What we measure

Four task categories, each testing a distinct retrieval primitive that AI coding agents actually need:

### P1 — Definition lookup
Given a symbol name, return the location where it's defined. Tolerance: ±3 lines (parsers disagree on whether "def line" means the signature or the function body opening). Scoring: exact match (1.0) or miss (0.0); precision/recall collapse to {0,1}.

**Why it's load-bearing:** every refactor starts with "where is this thing defined." This is the most basic primitive. A tool that fails P1 fails everything downstream.

### P2 — Reference finding
Given a symbol name, return all locations where it's used. Tolerance: ±2 lines. Scoring: bag-of-(file, line) — recall = matched / expected, precision = correct_preds / total_preds.

**Why it's load-bearing:** this is "what breaks if I rename this." Critical for refactor blast-radius. Note: tools that track *import* references (jcodemunch, GitNexus) score 0 here on purpose — they expose import-graph references, not call-site references. We chose the call-site definition because it's what users mean when they say "find every caller of X." The bench page documents this design choice.

### P4 — File dependencies
Given a file path, return what it imports and what imports it. Scoring: set membership over normalized paths (strip `./`, `.ts`, `.js` extensions). F1 is the harmonic mean of imports-side F1 and importers-side F1.

**Why it's load-bearing:** the test that AI agents need to answer "if I change this file, what else needs reviewing." Closer to the question users actually ask than P2.

### P5 — Dead code
Return a set of symbol names that look dead in the codebase (no references). Scoring: set overlap. Where the expected set is empty (express + lodash), recall is vacuous-1 and precision becomes the only signal — punishes false-positive floods.

**Why it's load-bearing:** dead code detection is the unique signal hybrid-retrieval systems can offer over pure search. Failing P5 means the call graph is broken.

## What we DON'T measure (yet)

- **Latency under concurrent load.** All measurements are single-process, single-thread. Real-world MCP usage is bursty.
- **Memory pressure on huge repos.** All datasets are ≤500K LOC. The 5M-LOC monorepo case is unmeasured.
- **Cross-language references.** All datasets are JS/TS monolingual. Python-calling-Rust via subprocess, or TS-calling-Python-via-FFI, isn't tested.
- **Update freshness.** All datasets are static checkouts. Incremental re-indexing under live edits isn't measured.
- **Quality of explanations.** We measure correctness of returned `(file, line)` tuples, not whether the agent can explain *why*.

These are roadmap items. Open issues in [sverklo/sverklo](https://github.com/sverklo/sverklo) tagged `bench` if you want to argue for prioritization.

## How the ground truth was constructed

Three categories of construction:

### Hand-authored static (sverklo monorepo)

`tasks/sverklo.jsonl` is hand-authored. Every `(file, line)` was checked by reading the file. The expected dead-code set was constructed by:

1. Listing all exported symbols in `src/`
2. For each, counting file references via `grep -rln`
3. Filtering for `count == 1` (the def file only) and excluding TypeScript types, constants, and dynamically-dispatched MCP tool handlers
4. Hand-verifying each candidate is not reached via dynamic dispatch

Methodology documented in [`tasks/sverklo-curation.md`](./tasks/sverklo-curation.md). The current 6-name expected set is in [issue #27](https://github.com/sverklo/sverklo/issues/27).

### Runtime-resolving generators (express, lodash)

`tasks/express.gen.ts` and `tasks/lodash.gen.ts` are generators. Hand-authoring `(file, line)` for a third-party tag would be brittle (each minor release shifts line numbers). Instead:

1. Declare *what* we want (symbol names + heuristics for definition/reference shapes)
2. The generator runs grep against the actual checkout to resolve names to `(file, line)` at bench startup
3. Resolution happens once, deterministically, before any baseline runs
4. All baselines compete on the *same resolved* tasks
5. The grep used for resolution is **not** the same as the smart-grep baseline — it can read files freely and uses regex patterns the baselines never see

This is honest because resolution is single-pass deterministic; baselines can't game it.

### What we did NOT do

- We did not let an LLM construct the ground truth. The bench's credibility depends on ground truth being human-validated.
- We did not pick tasks that we knew sverklo would win. The express P5 set is empty (vacuous recall) specifically because there's no clean dead-code signal in express's tightly-curated codebase. We kept it because the empty-expected pattern punishes false positives, which is informative.
- We did not artificially balance task counts to be flattering. P1 has 30 tasks because it's load-bearing. P5 has 15 because the signal is small relative to the others.

## Why these specific datasets

| Dataset | Why |
|---------|-----|
| **sverklo monorepo (TypeScript, modular)** | Most representative real-world repo because we wrote it. Allows ground truth verified by deep knowledge of the codebase, including dynamic-dispatch call sites that static analyzers miss. |
| **express 4.21.1 (JavaScript, modular CommonJS)** | Well-known, modest-size (≈30K LOC), uses CommonJS `module.exports` patterns that exposed jcodemunch's pre-v1.80.7 import-resolution blind spot. |
| **lodash 4.17.21 (JavaScript, single-file UMD/IIFE)** | The pathological case. lodash.js is 17,209 lines, 548 KB, all wrapped in a single IIFE. Stress-tests parsers' size caps, brace counters, and call-graph fallbacks for monolithic structures. Both jcodemunch and sverklo shipped lodash-specific parser fixes within 36 hours of this dataset being added. |

Glaring gaps acknowledged in [CONTRIBUTING.md §3](./CONTRIBUTING.md): no Python, no Rust, no Java/Kotlin, no multi-language. Roadmap item.

## Scoring details

Per-task records emitted as `raw.jsonl` with this shape:

```json
{
  "task_id": "ld-p1-01",
  "category": "P1",
  "dataset": "lodash",
  "baseline": "sverklo",
  "metrics": {
    "input_tokens": 871,
    "tool_calls": 1,
    "wall_time_ms": 45,
    "cold_start_ms": 17860,
    "warm_call_ms": 45,
    "recall": 1.0,
    "precision": 1.0,
    "f1": 1.0,
    "exact_match": true,
    "tokens_per_correct_answer": 871,
    "raw_payload_chars": 3770
  },
  "predicted_summary": "1 loc"
}
```

Aggregate `summary.json` averages F1, recall, precision, tokens, tools, and wall-time per `(baseline, category, dataset)` triple. The "gated tokens-per-correct-answer" column averages only over runs where F1 ≥ 0.8 — refuses to reward "found nothing cheaply" or "guessed correctly with garbage scoring."

Cold-start is reported separately because for MCP servers it dominates first-call latency (model load + index build). After the first task, all baselines are warm-call.

## Versioning

Bench task definitions are versioned with the main sverklo repo. Significant methodology changes (new categories, new datasets, scoring revisions) trigger a Zenodo paper revision and are documented in [CHANGELOG-bench.md](https://github.com/sverklo/sverklo/blob/main/benchmark/CHANGELOG.md).

The current bench version is **bench:primitives v2** (May 2026, 3-dataset, 5-baseline). The previous version (bench:primitives v1, April 2026, 2-dataset, 3-baseline) is preserved on the bench page for diff.

## Critique invited

This document is wrong somewhere. We don't know where yet. If you find it, [open an issue](https://github.com/sverklo/sverklo/issues/new) with the `bench-methodology` label.
