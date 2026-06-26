# Contributing to jCodeMunch-MCP

Thanks for your interest in contributing! A few things to know before you submit a PR.

## Contributor License Agreement

This project is dual-licensed — free for non-commercial use, with paid licenses for commercial use. To keep that model legally sound, **all contributors must sign the CLA before their PR can be merged.**

The CLA is short and plain-English: you keep your copyright, you grant the project the right to sublicense your contribution commercially, and you confirm the work is yours to submit.

**[Sign the CLA](https://cla-assistant.io/jgravelle/jcodemunch-mcp)**

CLA Assistant will prompt you automatically when you open a PR. It takes about 30 seconds.

## Commercial Licensing

If you're using jCodeMunch in a commercial context, see the [license section in the README](README.md#license-dual-use) for options.

## Getting Started

```bash
git clone https://github.com/jgravelle/jcodemunch-mcp
cd jcodemunch-mcp
pip install -e ".[test]"
pytest tests/ -q
```

## Guidelines

- Open an issue before starting large features — saves everyone time if direction needs discussion
- Keep PRs focused; one feature or fix per PR
- Include tests for new functionality
- Run the full test suite before submitting

## Quality gates that run on every release

- **Schema budget** — `tests/test_schema_budget.py` fails when `tools/list` token count grows more than 5% above `benchmarks/schema_baseline.json`. If you intentionally grow the schema (new tool / longer description), regenerate the baseline in the same PR with justification:
  ```bash
  PYTHONPATH=src python benchmarks/harness/capture_schema_baseline.py
  ```
- **Retrieval-quality replay (v1.76.0+)** — `benchmarks/replay/run_replay.py` runs golden queries through `search_symbols` and reports nDCG@10 / MRR@10 / Recall@10 against the locked v1.75.0 baseline. Any aggregate metric drop > 2% fails the gate:
  ```bash
  PYTHONPATH=src python benchmarks/replay/run_replay.py \
      --fixture benchmarks/replay/fixtures/self_v1_75_0.json \
      --baseline 1.75.0 --gate 0.02
  ```
  If your change legitimately moves a metric, capture a new fixture/baseline and document the reason in the PR description.
