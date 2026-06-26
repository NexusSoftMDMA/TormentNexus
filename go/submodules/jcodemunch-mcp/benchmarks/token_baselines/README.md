# Token Baselines

Per-release snapshots of token-savings + latency per tool. Used by
`analyze_perf(compare_release="1.74.0")` to detect regressions in compression
ratio or per-tool latency drift across releases.

## Schema

```jsonc
{
  "version": "1.74.0",
  "captured_at": "2026-04-25T08:34:00Z",
  "session": {
    "session_calls": 137,
    "session_tokens_saved": 1264476,
    "session_duration_s": 412.3
  },
  "tools": {
    "search_symbols": {
      "calls": 42,
      "tokens_saved": 308124,
      "p50_ms": 42.1,
      "p95_ms": 188.4
    },
    "...": "..."
  }
}
```

## Capturing

Run any representative workload against the indexer (the README's
"benchmark commands" or the harness in `benchmarks/harness/run_benchmark.py`),
then snapshot:

```bash
python benchmarks/harness/capture_token_baseline.py
```

The script writes `benchmarks/token_baselines/v{VERSION}.json` derived from
the live `get_session_stats` reading.

## Comparing

```python
analyze_perf(window="session", compare_release="1.74.0")
# returns _meta.baseline_diff with per-tool deltas
```

The compare path is read-only — it never mutates the saved baseline file.
