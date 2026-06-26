# jcodemunch-mcp — Token Efficiency Benchmark

**Result: 99.6% average token reduction · tiktoken cl100k_base · 15 task-runs · 3 repos**

## What this measures

How many tokens a code-retrieval tool consumes versus an agent that reads every source file before acting.

**Baseline:** concatenate all indexed source files and count tokens. This is the *minimum* cost for a "read everything first" agent — real agents typically read files multiple times, so production savings are higher.

**jcodemunch workflow:** `search_symbols` (top 5 results) + `get_symbol_source` × 3 hits per query. Total = search response tokens + 3 × symbol source tokens.

**Tokenizer:** `tiktoken cl100k_base` — the GPT-4 / Claude family encoding. Consistent across runs regardless of model.

## Reproducing the results

```bash
pip install jcodemunch-mcp tiktoken

# Index the three canonical repos
jcodemunch index_repo expressjs/express
jcodemunch index_repo fastapi/fastapi
jcodemunch index_repo gin-gonic/gin

# Run the benchmark (prints markdown table + grand summary)
python benchmarks/harness/run_benchmark.py

# Optional: write results to file
python benchmarks/harness/run_benchmark.py --out benchmarks/results/my_run.md
```

## Task corpus

Tasks are defined in [`tasks.json`](tasks.json) — 5 queries × 3 repos = 15 measurements.

| ID | Query | Description |
|----|-------|-------------|
| `router-route-handler` | `router route handler` | Core route registration / dispatch logic |
| `middleware` | `middleware` | Middleware chaining and execution |
| `error-exception` | `error exception` | Error handling and exception propagation |
| `request-response` | `request response` | Request/response object definitions |
| `context-bind` | `context bind` | Context creation and parameter binding |

Repos: `expressjs/express`, `fastapi/fastapi`, `gin-gonic/gin`

## Canonical results

Full per-task tables are in [`results.md`](results.md).

| Repo | Files | Baseline tokens | Avg reduction |
|------|------:|----------------:|--------------:|
| expressjs/express | 165 | 137,978 | **99.4%** |
| fastapi/fastapi | 951 | 699,425 | **99.8%** |
| gin-gonic/gin | 98 | 187,018 | **99.4%** |
| **Grand total** | — | 5,122,105 | **99.6%** |

**99.6% average token reduction** across 15 task-runs · 263.9x ratio · tiktoken cl100k_base.

To regenerate:

```bash
python benchmarks/harness/run_benchmark.py --out benchmarks/results.md
```

## Benchmarking a different tool

The task corpus in `tasks.json` is tool-agnostic. To evaluate another tool:

1. Use the same 3 repos and 5 queries.
2. Use the same baseline: all indexed source files concatenated, tokenized with `tiktoken cl100k_base`.
3. Measure total tokens consumed by your retrieval workflow per query (tool calls + responses).
4. Report per-task rows and the grand average using the same formula: `(1 - tool_tokens / baseline_tokens) * 100`.

If you publish results against this corpus, open an issue or PR and we'll link them here.

## Methodology notes

- The baseline is a lower bound. Agents that re-read files mid-task spend more.
- The jcodemunch workflow counts `search_symbols` + `get_symbol_source` responses only — it does not count system prompt or tool description tokens, which are identical for both approaches.
- Token counts are from serialized JSON responses, not raw source, so they include field names and structure overhead. This slightly understates the reduction.

## Related harnesses (v1.74.0+)

- **`benchmarks/replay/`** — replayable retrieval-quality benchmark.
  Fixtures pin `(query, expected_top_k_ids)` tuples; the harness runs
  each query through `search_symbols` and reports nDCG@k, MRR@k, and
  Recall@k. CI gate: `run_replay.py --baseline X.Y.Z --gate 0.02` exits
  non-zero if any aggregate metric drops by more than 2% vs the saved
  baseline. The shipped `self_v1_75_0` fixture is locked at 1.0 across
  all metrics — every release runs against it.
- **`benchmarks/token_baselines/`** — per-release token-savings + latency
  snapshots. `capture_token_baseline.py` reads
  the live session's `get_session_stats` + `latency_stats` and writes
  `benchmarks/token_baselines/v{VERSION}.json`. The `analyze_perf` tool
  consumes these via `compare_release="X.Y.Z"`.
