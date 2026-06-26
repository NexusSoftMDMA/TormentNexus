# Contributing to bench:primitives

This bench is designed to be **adversarial** — we want competing tools, hostile critique of our methodology, and PRs from people who think they can beat us. Three ways to contribute.

## 1. Submit a baseline implementation

If you maintain a code-search tool, code-intelligence MCP server, or retrieval system, you can have it benchmarked here on the same task suite as the existing baselines. **No "I'll evaluate it on my own data" excuses.**

The baseline interface is one TypeScript file in the main sverklo repo. Reference implementations:

- [`benchmark/src/baselines/sverklo.ts`](https://github.com/sverklo/sverklo/blob/main/benchmark/src/baselines/sverklo.ts) — MCP stdio client
- [`benchmark/src/baselines/jcodemunch.ts`](https://github.com/sverklo/sverklo/blob/main/benchmark/src/baselines/jcodemunch.ts) — MCP stdio client (Python via uvx)
- [`benchmark/src/baselines/gitnexus.ts`](https://github.com/sverklo/sverklo/blob/main/benchmark/src/baselines/gitnexus.ts) — CLI client
- [`benchmark/src/baselines/naive-grep.ts`](https://github.com/sverklo/sverklo/blob/main/benchmark/src/baselines/naive-grep.ts) — pure shell baseline

You implement `Baseline` (`name`, `setupForDataset`, `run(task)`, optional `teardownForDataset`). The runner takes care of timing, scoring, and report generation.

### The Baseline interface

You don't need to clone the runner repo just to read the contract. Mirrored here for convenience (authoritative source: [`benchmark/src/baselines/base.ts`](https://github.com/sverklo/sverklo/blob/main/benchmark/src/baselines/base.ts)):

```typescript
export type TaskCategory = "P1" | "P2" | "P4" | "P5";

export interface Task {
  id: string;                    // e.g. "ex-p4-04"
  category: TaskCategory;
  dataset: string;               // "express" | "lodash" | "sverklo" | "requests"
  query: string;                 // symbol name (P1/P2) or file path (P4)
  expected: unknown;             // shape depends on category — runner handles scoring
}

export interface BaselineOutput {
  prediction: unknown;           // your structured answer (def location, ref list, deps map, etc.)
  payload: string;               // raw text the agent would actually see
  toolCalls: number;             // how many tool calls you made for this task
  inputTokens: number;           // tokens the agent consumed across all tool results
  warmCallMs: number;            // wall time of the answering call (excluding cold start)
  coldStartMs: number;           // wall time of any one-time setup before the first task
}

export interface Baseline {
  name: string;
  setupForDataset(d: { name: string; rootPath: string }): Promise<void>;
  run(task: Task): Promise<BaselineOutput>;
  teardownForDataset?(): Promise<void>;
}
```

`setupForDataset` is called once per dataset (good place to spawn an MCP subprocess and wait for indexing). `run(task)` is called per task and must return the structured `prediction` plus the metrics. `teardownForDataset` is optional — use it to kill subprocesses or clean up temp state.

The harness handles: dataset cloning, task generation, ground-truth scoring, F1/P/R math, per-category aggregation, report rendering. You only own the bridge from your tool's API to `prediction`.

**Process:**

1. Open a PR to [github.com/sverklo/sverklo](https://github.com/sverklo/sverklo) adding `benchmark/src/baselines/<your-tool>.ts`
2. Register it in `benchmark/src/runner/run-primitive.ts`
3. **Auto-bench CI runs on the PR** — see [#4](https://github.com/sverklo/sverklo-bench/issues/4). The workflow detects which baseline files changed, runs the harness against the express dataset (~30 tasks), and posts a results-table comment back to your PR within ~10 minutes. You don't need to run `npm run bench:quick` locally before opening the PR; CI does it.
4. Inspect the comment, iterate. Full results upload as a GitHub Actions artifact for raw inspection.
5. We'll merge if (a) the baseline runs cleanly, (b) the implementation is faithful to your tool's intended use, and (c) it doesn't break existing baselines.

For the full 120-task suite (express + lodash + sverklo + requests): run `npm run bench:quick` locally — too slow for free-tier CI runners.

We will not refuse to bench a tool because it makes sverklo look bad. **The bench losing teaches the bench more than the bench winning.**

## 2. Challenge the methodology

If you think a task is wrong, the scoring is unfair, or a tolerance is too generous, **open an issue with a specific test case**. Examples of legitimate critique we've already merged:

- [#27](https://github.com/sverklo/sverklo/issues/27) — sv-p5 expected set was a thin 5-name set; refined to 6 confirmed-dead exported functions with documented verification methodology
- [#26](https://github.com/sverklo/sverklo/issues/26) — original 2-dataset bench (express + sverklo monorepo) was too narrow. Added lodash 4.17.21 to expose monolithic-IIFE blind spots. Both jcodemunch and sverklo had to ship parser fixes within 36 hours.
- [#28](https://github.com/sverklo/sverklo/issues/28) — sverklo P2/P4 regression after the parser fix. Tracked publicly with bench numbers attached.

A well-formed methodology critique points at:

- A specific task ID (`sv-p1-04`, `ex-p4-02`, `ld-p5-01`, etc.)
- The expected vs predicted output you observe
- Why your interpretation differs from the bench's

Critiques without a reproducible task ID can't be resolved and will be politely closed.

## 3. Add a dataset

The current 4-dataset matrix (sverklo monorepo / express / lodash / requests) covers TypeScript modular, JavaScript CommonJS modular, JavaScript monolithic-IIFE, and Python modular. Glaring gaps:

- **Rust** — none.
- **Java/Kotlin** — none. A monorepo platform like Gradle's source tree would be a hard test.
- **Multi-language repos** — none with cross-language reference resolution.
- **Real monorepos** — nx, turborepo, yarn workspaces. The current "monorepo" is just the sverklo repo itself.

Adding a dataset means: pick the repo + ref, add to `benchmark/src/datasets/manifest.json`, write a `<dataset>.gen.ts` ground-truth generator (10 P1 + 10 P2 + 5 P4 + 5 P5 = 30 tasks). Reference: [`lodash.gen.ts`](https://github.com/sverklo/sverklo/blob/main/benchmark/src/ground-truth/seed/lodash.gen.ts) is the cleanest example.

PRs adding datasets must include hand-verified ground truth (line numbers checked against the actual files) and a brief writeup of what shape of code the dataset exposes that the existing three don't.

## What we will NOT accept

- **Cherry-picked tasks** designed to make a specific baseline look good. The 120-task suite is fixed; adding tasks means adding 30 at once across all four categories.
- **Synthetic ground truth** generated by an LLM without human verification. The bench's credibility depends on the ground truth being human-validated.
- **Baselines that hard-code task answers.** The harness will fail you if your baseline returns the same locations across different datasets, but please don't try.
- **PRs that change scoring rules.** If you think the scoring is wrong, open an issue first. Drive-by tolerance changes will be reverted.

## Communication

- Issues: [github.com/sverklo/sverklo/issues](https://github.com/sverklo/sverklo/issues) (use the `bench` label)
- For methodology debates that span more than one issue, the tag we use across the GitHub ecosystem is `bench:primitives`

This repo's issues are for bench-specific concerns; product issues live in the main sverklo repo.

## Hall of fame

Past contributors who improved the bench:

- **[@HaleTom](https://github.com/HaleTom)** — opened [#25](https://github.com/sverklo/sverklo/issues/25), the comparison ask that drove the May 2026 5-baseline expansion
- **[@jgravelle](https://github.com/jgravelle)** — jcodemunch-mcp maintainer; shipped v1.80.7/8/9 fixes against bench findings within hours, then helped diagnose the lodash blind spot that exposed sverklo's own brace-counter bug

If you contribute methodology improvements, you'll be added here.
