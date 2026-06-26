# RFC-001: Code-Intel Bracket Spec for the MCP Server Arena

**Status:** Open for comment. Spec freezes after at least one external baseline lands through auto-bench CI.

**Genesis:** [@jmunchLlc on r/mcp](https://www.reddit.com/r/mcp/comments/1t4n9un/) (Jake Gravelle, jcodemunch-mcp maintainer) proposed an "MCP Server Arena" pattern in the public discussion around the original sverklo-bench publication. The bigger opportunity, in his framing: a daily go-to for MCP developers, segmented by category, with measured comparisons across same-substrate tools.

This RFC takes the segmentation problem as load-bearing and proposes the **code-intel bracket** as v1. If the spec lands well, the same shape extends to browser-automation, data-extraction, and shell-execution brackets later — each with its own metric set. The decision to start narrow is deliberate: trying to spec four brackets at once is the wrong scope.

Tracking issue: [#5](https://github.com/sverklo/sverklo-bench/issues/5). Comment there; this doc captures the current state of the spec.

---

## What "code-intel" includes

A baseline qualifies for the code-intel bracket if it:

- Exposes an MCP-shaped interface (stdio, JSON-RPC, the standard `tools/call` surface).
- Operates over a developer's source code as the primary input.
- Returns retrieval-shaped output: file paths, line numbers, symbol names, or natural-language explanations grounded in code.
- Runs on a developer's machine OR returns results without round-tripping the user's code through a third-party LLM that wasn't invoked by the user themselves.

What this excludes (and why):

- **PR review bots** (Greptile, CodeRabbit). They consume diffs, not arbitrary user queries; different evaluation contract.
- **IDE-bound retrieval** (Cursor's `@codebase`, Cody's symbol search). They're embedded in a host editor; the unit of evaluation isn't a standalone MCP call.
- **Agent frameworks** (Aider, Continue, Codex CLI). They orchestrate other tools; benchmarking them confuses the orchestration with the retrieval.
- **Shell-execution wrappers**. Different bracket entirely.

This is the cleanest line I can draw. Disagreements welcome — flag specific tools that should or shouldn't qualify.

## The metric set (proposed v1)

The current `bench:primitives` task suite already exercises four primitives: P1 (definition lookup), P2 (reference finding), P4 (file dependencies), P5 (dead-code detection). I propose carrying these forward, with one addition:

| Metric | Why it's load-bearing |
|---|---|
| **F1 (overall)** | The standard. Reported per-category to avoid hiding losses in averages. |
| **F1 (per category)** | Different baselines win different categories. Smart-grep wins P2; sverklo wins P4; jcodemunch ties P1. The arena framing fails if the leaderboard collapses tradeoffs. |
| **Input tokens / task** | Load-bearing for AI agent integration. The wedge between sverklo (469) and naive grep (20,278) is the part most retrieval evaluations omit. |
| **Tool calls / task** | Same. A baseline that wins F1 by making 12 calls is a different product than one that wins F1 in 1 call. |
| **Cold-start ms** | Reported separately, not amortized. Engineers deciding whether to install need to see "first call latency" separately from "warm-call latency." |
| **License** | Not a bench number, but listed alongside. PolyForm-Noncommercial and MIT belong on different lines of the table for any commercial-deployment evaluation. |

What I'm explicitly **not** proposing as a bracket-level metric: cosine-similarity-against-a-curated-set, ROUGE/BLEU, or any LLM-as-judge scoring. The bench-loop only fires on metrics where contributors can verify the response themselves, and LLM-as-judge has too much sampling variance to be a credible eval surface. Open to revisiting if someone has a defensible scoring shape.

## The contract (what a baseline implements)

The current [`Baseline` interface](https://github.com/sverklo/sverklo/blob/main/benchmark/src/baselines/base.ts) is the proposed contract:

```typescript
interface Baseline {
  name: string;
  setupForDataset(dataset: { name: string; rootPath: string }): Promise<void>;
  teardownForDataset?(): Promise<void>;
  run(task: Task): Promise<BaselineOutput>;
}
```

`BaselineOutput` carries `{ prediction, rawPayload, toolCalls, wallTimeMs, coldStartMs, warmCallMs, notes? }`. The `rawPayload` field is the LLM-visible string used for honest token counting.

This is a good contract for the code-intel bracket because it forces baselines to declare:

1. What they'd return to an agent (`prediction`)
2. What it would cost (`rawPayload`, measured in tokens by tiktoken at scoring time)
3. How they handled cold-start (separately reported, not amortized)

Open to changes. Specifically: the contract assumes per-task isolation (no cross-task state). For baselines that want to amortize learning across tasks within a session, we'd need a different shape.

## Datasets

Currently three: `express` (CommonJS modular JS), `lodash` (single-file IIFE JS), `sverklo` (TS monorepo). 30 tasks each, 90 total. Documented at [github.com/sverklo/sverklo-bench](https://github.com/sverklo/sverklo-bench).

Open issues:

- [#1](https://github.com/sverklo/sverklo-bench/issues/1): Python codebase. Real gap; the bracket can't seriously claim to evaluate Python code-intel without a Python dataset.
- Rust, Java, Go: not on the immediate roadmap. PRs welcome.

The bracket spec doesn't lock the dataset list — datasets can grow under the spec. What's locked is the task-category structure (P1/P2/P4/P5) and the per-dataset task budget (10/10/5/5 = 30).

## Candidate baselines for v1

| Tool | Status |
|---|---|
| naive-grep | Already on bench. The floor. |
| smart-grep | Already on bench. Tuned ripgrep with language filters and definition-shaped patterns. |
| sverklo | Already on bench (the host project). MIT, npm. |
| jcodemunch-mcp | Already on bench. MIT, uvx. Active maintenance ([@jgravelle](https://github.com/jgravelle)). |
| GitNexus | Already on bench. PolyForm Noncommercial 1.0, npm + native KuzuDB. |
| Serena | **Open invitation.** MIT, LSP-backed. [Filed as part of vs/serena](https://sverklo.com/vs/serena/). PR welcome. |
| codebase-memory-mcp | **Open invitation.** Single C binary, 66 languages tree-sitter. [Filed as part of vs/codebase-memory-mcp](https://sverklo.com/vs/codebase-memory-mcp/). PR welcome. |

Auto-bench CI runs on every baseline-touching PR. New entrants don't need to wait for maintainer ceremony.

## What's deliberately out of scope for v1

- **Multi-language cross-file reference resolution.** The bench is currently single-language per dataset.
- **LLM-augmented baselines** (a baseline that wraps an LLM around `sverklo_search` and counts the wrapped output). Different evaluation contract; would need its own bracket.
- **Cost/$.** API pricing changes weekly; out-of-band.
- **Quality-of-life features** (UI, dashboards, IDE integration). Not retrieval; not bench-shaped.

## Open questions

- **Is the per-category task budget (10/10/5/5) the right shape?** P5 dead-code currently has the smallest budget; should it grow given how much P5 has driven cross-iteration (the jcodemunch lodash arc was P5-heavy)?
- **Does anyone want a P3 (semantic neighborhood / "what's this related to")?** Currently there isn't one. Trade-off: the more categories, the more brittle the average; the fewer, the less the bracket measures.
- **Tolerance defaults.** P1 ±3 lines, P2 ±2 lines, P4/P5 set membership. Any baseline maintainer with a position on this — speak up.
- **Naming.** "Code-intel bracket" is a working name. Better suggestions welcome.

## Process

This RFC stays open for ~2 weeks of comment. Spec freezes after at least one external baseline lands through [auto-bench CI](https://github.com/sverklo/sverklo/blob/main/.github/workflows/auto-bench.yml) — the friction surfaced by an actual external submission should inform the final contract.

For the maintainers of any tool listed above (or any tool that thinks it should be listed): the explicit ask is to weigh in on metrics, contract, or category boundary you'd want changed. The bench-loop ethic applies — if your tool would lose under this spec on a specific axis, name the axis publicly so the spec can either accommodate it or honestly reject it.

