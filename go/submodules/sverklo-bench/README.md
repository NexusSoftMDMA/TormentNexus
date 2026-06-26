# bench:primitives

A public, reproducible benchmark for code-intelligence MCP servers and code-search baselines.

**150 hand-verified tasks** across **5 OSS codebases** (sverklo, express 4.21.1, lodash 4.17.21, requests 2.32.3, flask 3.0.3), **4 task categories** (definition lookup, reference finding, file dependencies, dead code), **5 baselines** (naive grep, smart grep, [sverklo](https://github.com/sverklo/sverklo), [jcodemunch-mcp](https://github.com/jgravelle/jcodemunch-mcp), [GitNexus](https://github.com/abhigyanpatwari/GitNexus)).

Reproducible from a fresh clone with one npm script.

## Why this repo exists

When you publish a benchmark and your competitor responds by shipping fixes within hours, that's the bench working as intended. This repo is the eval surface that loop runs on. It's split out from [github.com/sverklo/sverklo](https://github.com/sverklo/sverklo) (the primary code-intelligence MCP server) so the methodology gets its own audit trail, its own contributor PRs, and its own credibility signal independent of the tool that wrote it.

If you build a code-intelligence MCP server, code-search tool, or retrieval system in this neighborhood, **you can submit a baseline implementation here** and we'll run it on the same task suite. See [CONTRIBUTING.md](./CONTRIBUTING.md). The category spec the bench is being measured under is [RFC-001: Code-Intel Bracket](./RFC-001-code-intel-bracket.md) (open for comment until 2026-05-21).

## Add your tool in four steps

You don't need to download datasets, ground truth, or anything else manually. The harness fetches everything on demand.

1. Fork [github.com/sverklo/sverklo](https://github.com/sverklo/sverklo) and add `benchmark/src/baselines/<your-tool>.ts` implementing the [`Baseline` interface](./CONTRIBUTING.md#the-baseline-interface).
2. Register it in `benchmark/src/runner/run-primitive.ts` (one line — the existing baselines show the shape).
3. Open a PR. Auto-bench CI runs your baseline against the express dataset (~30 tasks) and posts a results-table comment back within ~10 minutes. You don't run anything locally.
4. Iterate against the comment until you're happy, we merge, results land on the [next refresh](https://sverklo.com/mcp/).

Reference baselines live in [`benchmark/src/baselines/`](https://github.com/sverklo/sverklo/tree/main/benchmark/src/baselines): `sverklo.ts` (MCP stdio), `jcodemunch.ts` (MCP stdio via uvx), `gitnexus.ts` (CLI), `naive-grep.ts` (pure shell). Pick the closest shape to your tool and start from there.

## Where the data lives

Common confusion: there is no "dataset repo" you clone separately. Datasets are fetched automatically by `npm run bench:quick` from upstream at pinned tags:

| dataset | source | tag |
|---|---|---|
| express | [expressjs/express](https://github.com/expressjs/express) | 4.21.1 |
| lodash | [lodash/lodash](https://github.com/lodash/lodash) | 4.17.21 |
| requests | [psf/requests](https://github.com/psf/requests) | v2.32.3 |
| sverklo | [sverklo/sverklo](https://github.com/sverklo/sverklo) (this monorepo) | current HEAD |

Cloned into `benchmark/.cache/<dataset>/` on first run. Re-runs reuse the cache. Task definitions for each dataset are mirrored in [`tasks/`](./tasks/) below.

## Headline results (May 2026, sverklo v0.20.2)

| baseline | n | F1 | P1 (def lookup) | P2 (ref finding) | P4 (file deps) | avg input tokens | tools/task | audit grade |
|---|---|---|---|---|---|---|---|---|
| **sverklo** | 120 | **0.58** | 0.70 | 0.29 | **0.78** | 498 | **1.0** | B |
| smart-grep | 120 | 0.41 | 0.33 | **0.30** | 0.46 | 963 | 4.1 | — |
| jcodemunch | 120 | 0.32 | **0.78** | 0.00 | 0.34 | 1,178 | 1.2 | C |
| naive-grep | 120 | 0.27 | 0.07 | 0.14 | 0.42 | 24,194 | 6.1 | — |
| gitnexus | 120 | 0.24 | 0.23 | 0.00 | 0.25 | **333** | 1.2 | F |

**Headline:** Sverklo leads overall F1 (0.58 vs smart-grep 0.41). Jcodemunch beats sverklo on P1 definition lookup outright (0.78 vs 0.70). Smart-grep beats sverklo on P2 reference finding (0.30 vs 0.29). Sverklo wins file-dependency reasoning by a wide margin (0.78 vs jcodemunch 0.34). GitNexus wins token cost (333 avg). Sverklo's audit grade is B with an F on coupling — `indexer.ts` has fan-in 60. P5 dead-code is omitted from the headline because every baseline returns the same empty-expected score (0.67), which makes the column non-discriminating; per-category numbers are still in the per-task output.

[Full ranking page →](https://sverklo.com/mcp/) · [Per-category breakdown and the slice where each baseline loses →](https://sverklo.com/bench/)

## Methodology

- Tasks are **hand-verified** at ground-truth construction time. Each P1, P2, P4 task points at exact `(file, line)` tuples, validated by reading the file. P5 tasks list expected dead-code names, validated against the actual codebase.
- Tolerances: P1 uses **±3-line** tolerance (parsers disagree on "def line"). P2 uses **±2 lines**. P4/P5 use **set membership**.
- Token cost is measured as **input tokens to the consuming agent**, including all tool-call results. The metric that matters in production is "how many tokens does the agent eat to get an answer," not "how big is the index file."
- The "honesty section" of the writeup lists every task where sverklo loses to a baseline. Same for every baseline. No cherry-picking.
- Cold-start is reported separately from warm-call latency. Both are loud.

Full methodology document: [METHODOLOGY.md](./METHODOLOGY.md).

## Reproducing the bench

The runtime lives in the main sverklo monorepo:

```bash
git clone https://github.com/sverklo/sverklo.git
cd sverklo
npm install
npm run build
npm run bench:quick
```

Results land in `benchmark/results/<timestamp>/` as `raw.jsonl` (per-task records), `summary.json` (aggregate per baseline + category), and `report.md` (human-readable).

To run a single baseline: `BASELINES=sverklo,jcodemunch npm run bench:quick`.

To run a single dataset (e.g. for fast iteration during development): `DATASETS=express npm run bench:quick`.

To re-validate a competitor's published claim, point the harness at their fork and re-run. The task definitions in this repo are authoritative; the harness is a reference implementation.

### Auto-bench CI on baseline PRs

Every PR to [sverklo/sverklo](https://github.com/sverklo/sverklo) that touches `benchmark/src/baselines/**` automatically runs the harness against the express dataset (~30 tasks) and posts a results table back as a comment within ~10 minutes. Idempotent — re-running the workflow updates the same comment in place. Full results upload as a GitHub Actions artifact for raw inspection.

Contributors don't need to run `npm run bench:quick` locally before opening the PR; CI does it. See [sverklo-bench#4](https://github.com/sverklo/sverklo-bench/issues/4) for the implementation notes and [.github/workflows/auto-bench.yml](https://github.com/sverklo/sverklo/blob/main/.github/workflows/auto-bench.yml) for the workflow source.

## Task definitions (copies for reference)

The authoritative ground truth lives at [github.com/sverklo/sverklo/tree/main/benchmark/src/ground-truth/seed](https://github.com/sverklo/sverklo/tree/main/benchmark/src/ground-truth/seed). Mirrored here as a stable read-only reference:

- [tasks/sverklo.jsonl](./tasks/sverklo.jsonl) — 30 tasks against the sverklo monorepo (TS)
- [tasks/express.gen.ts](./tasks/express.gen.ts) — 30-task generator against express 4.21.1 (CommonJS, modular)
- [tasks/lodash.gen.ts](./tasks/lodash.gen.ts) — 30-task generator against lodash 4.17.21 (single-file UMD/IIFE — different shape)
- [tasks/requests.gen.ts](./tasks/requests.gen.ts) — 30-task generator against requests 2.32.3 (Python, the first non-JS dataset in the bench)

Datasets covered:
- **sverklo monorepo** — TypeScript, modular, the most representative real-world repo (it's the project's own codebase)
- **express 4.21.1** — JavaScript, modular CommonJS, well-known structure
- **lodash 4.17.21** — JavaScript, single 17K-line UMD/IIFE wrapper. Pathological for parsers — exposed blind spots in both jcodemunch (size cap, IIFE call-graph fallback) and sverklo (regex brace counter mis-counting inside string literals) within 36 hours of being added to the bench.
- **requests 2.32.3** — Python. Surfaced a real bug in sverklo's own parser within hours of being added: relative imports (`from .adapters import HTTPAdapter`) were being emitted as the literal string `.adapters` instead of being resolved against the importing file's directory. Fix landed in the same commit as the dataset; sverklo P4 on requests went 0.10 → 1.00. The dataset earned its slot by surfacing a parser bug that no JS-only dataset would have.
- **flask 3.0.3** — Python web framework. Added by [@yallalaraja](https://github.com/yallalaraja) in [PR #33](https://github.com/sverklo/sverklo/pull/33) — the first unaffiliated community contribution. Tests decorator-heavy patterns (Blueprint registration, class-based views, dynamic route binding) and module-level LocalProxy globals (`flask.request`) that pure-function Python like requests doesn't exercise. Surfaced three real bugs within the same review cycle: a regex parser that truncated functions whose signatures contained `Array<{...}>` type annotations (sverklo `findBraceEnd`, +464 symbol refs recovered repo-wide), `audit --format json` contaminating stdout with model-download progress (broke CI JSON parsing), and `auto-bench.yml` + `audit-self.yml` not handling fork-PR `GITHUB_TOKEN` write restrictions. All four fixes shipped in sverklo v0.20.17–v0.20.18 alongside the merge.

## How to add a baseline

Quick walkthrough is at the [top of this README](#add-your-tool-in-four-steps). The full process — including the [`Baseline` interface contract](./CONTRIBUTING.md#the-baseline-interface), what we will and won't accept, and how methodology critique gets handled — is in [CONTRIBUTING.md](./CONTRIBUTING.md).

## Contributors

The bench is open infrastructure. Datasets and baselines come from the community as much as from the maintainer.

- [@yallalaraja](https://github.com/yallalaraja) — Flask 3.0.3 dataset (30 hand-verified tasks). First unaffiliated contribution. [PR #33](https://github.com/sverklo/sverklo/pull/33).
- [@HaleTom](https://github.com/HaleTom) — surfaced jcodemunch-mcp as a baseline candidate. [Issue #25](https://github.com/sverklo/sverklo/issues/25).
- [@jgravelle](https://github.com/jgravelle) — maintainer of [jcodemunch-mcp](https://github.com/jgravelle/jcodemunch-mcp), reviewed the early bench methodology, publicly endorsed an LMSYS-style arena evolution on r/mcp.

PRs welcome. See [CONTRIBUTING.md](./CONTRIBUTING.md) for the baseline interface and dataset format. Open [an issue](https://github.com/sverklo/sverklo-bench/issues) before starting work on a baseline so we can flag any methodology concerns early.

## License

MIT. Same as the [main sverklo repo](https://github.com/sverklo/sverklo).

The bench task files are CC-BY-4.0 — reuse them in academic work, in your own evaluations, or to argue with our methodology. Cite the [Zenodo paper](https://doi.org/10.5281/zenodo.19802051) where applicable.

---

**Related:**
- [github.com/sverklo/sverklo](https://github.com/sverklo/sverklo) — the MCP server being benchmarked (the "sverklo" baseline above) and the home of the bench runner + baseline implementations under `benchmark/`
- [sverklo.com/mcp/](https://sverklo.com/mcp/) — the public MCP code-intel ranking page (sortable, with audit grades)
- [sverklo.com/bench/](https://sverklo.com/bench/) — per-category breakdowns and the slices where each baseline loses
- [Issue #25](https://github.com/sverklo/sverklo/issues/25) — original "compare against jcodemunch and GitNexus" thread that drove the May 2026 expansion
- [Issue #29](https://github.com/sverklo/sverklo/issues/29) — late-interaction rerank experiment; close-out writeup at [sverklo.com/blog/late-interaction-rerank-made-our-f1-worse/](https://sverklo.com/blog/late-interaction-rerank-made-our-f1-worse/)
- [Zenodo paper, CC-BY 4.0](https://doi.org/10.5281/zenodo.19802051) — peer-reviewable methodology writeup
