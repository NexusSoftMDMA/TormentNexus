# Benchmarks

mnem is measured head-to-head against mem0 and MemPalace on six public datasets. All numbers are reproducible with the shipped harness.

## Results summary

| Dataset | mnem R@5 | mem0 R@5 | delta |
|---|---|---|---|
| LoCoMo | - | - | +0.218 |
| MemBench | - | - | +0.120 |
| ConvoMem | - | - | +0.047 |
| LongMemEval | 0.966 | - | tied with MemPalace |

Full numbers with confidence intervals: [`benchmarks/results/v0.1.0/`](../../benchmarks/results/v0.1.0/).

## Methodology

- **Embedder**: ONNX MiniLM-L6-v2, same bytes on every system. No LLM rerank.
- **mem0 columns**: our own reproduction under the same harness. mem0 does not publish R@K headline scores on these datasets.
- **MemPalace columns**: public headline numbers cross-verified under our harness.
- **FinanceBench**: uses Ollama bge-large (1024-dim) on all systems for fair comparison.

This is disclosed, not hidden. Reproducible artifacts ship alongside the binary.

## Reproduce

```bash
mnem bench fetch longmemeval          # download datasets (one-time, 264 MB)
mnem bench                            # TUI; select benchmarks interactively
mnem bench run --benches longmemeval --limit 50 --non-interactive

# Legacy bash harness (canonical path for headline numbers)
bash benchmarks/harness/run_bench.sh
```

## See also

- [Benchmark methodology](../src/benchmarks/methodology.md)
- [Reproduce benchmarks](../src/benchmarks/reproduce.md)
- [Raw results](../../benchmarks/results/v0.1.0/)
