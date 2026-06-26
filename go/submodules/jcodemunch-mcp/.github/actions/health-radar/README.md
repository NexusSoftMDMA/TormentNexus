# jCodeMunch Health Radar — GitHub Action

A composite action that computes a six-axis code-health radar on the PR
branch and the base branch, diffs them, and posts the result as a
**sticky PR comment**. Suggestion-style — never blocks merges.

## What it does

1. Indexes the PR branch with `jcodemunch-mcp index .`
2. Runs `jcodemunch-mcp health . --radar-only` → PR radar JSON
3. Checks out the base branch, re-indexes, runs the same command → base
   radar JSON
4. Computes the diff via the same pure helper exposed as the
   `diff_health_radar` MCP tool
5. Renders a markdown table + axis movements + regression / improvement
   bullets
6. Finds an existing sticky comment by HTML marker — `PATCH`es it on
   re-runs, `POST`s a new one on the first run.

## Why a comment, not a status check

Status checks block merges. A heuristic that blocks merges gets the
Action disabled by the first frustrated maintainer. The radar comment
is **explanatory, not gating** — reviewers see the deltas, decide for
themselves.

## Usage

```yaml
# .github/workflows/health-radar.yml
name: Health Radar
on:
  pull_request:
    types: [opened, synchronize, reopened]

permissions:
  pull-requests: write   # for sticky comment
  contents: read         # for git checkout

jobs:
  radar:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0    # need full history for index + base checkout
      - uses: jgravelle/jcodemunch-mcp/.github/actions/health-radar@v1.88.0
```

That's the whole setup. The action handles install, index, base/branch
toggling, and comment posting itself.

## Inputs

| Input | Default | Description |
|---|---|---|
| `python-version` | `3.11` | Python version on the runner. |
| `jcodemunch-version` | `latest` | Pin a specific package version, or `latest`. |
| `base-ref` | _(PR's base branch)_ | Override the comparison ref. |
| `github-token` | `${{ github.token }}` | Token used to post/edit the comment. |

## Output shape

The sticky comment looks like:

```markdown
<!-- jcm-health-radar -->
## jCodeMunch Health Radar

🟡 **Composite:** B → C (-7.5 pts)
🔴 **Verdict:** REGRESSION on 2 axis/axes (composite -7.5)

| Axis | Baseline | PR | Δ |
|---|---:|---:|---:|
| `complexity`     | 88 | 64 | **-24.0** ↓ |
| `dead_code`      | 82 | 79 | -3.0 ↓ |
| `cycles`         | 100 | 100 | +0.0 · |
...

### Regressions
- `complexity`: raw 4.5 → 11.0
- `dead_code`: raw 4.5 → 5.7
```

## Methodology

Per-axis scoring rules and rationale: see
[`tools/health_radar.py`](https://github.com/jgravelle/jcodemunch-mcp/blob/main/src/jcodemunch_mcp/tools/health_radar.py).

The composite is the arithmetic mean of every scored axis;
`omitted_axes` lists axes whose underlying signals weren't available
(e.g. `test_gap` if `get_untested_symbols` couldn't run).

## Known caveats

- **Re-indexes on both branches** — adds runtime in CI, scaling roughly
  with codebase size. For very large repos, the baseline could be
  cached by base SHA in a follow-up release.
- **Heuristic, not coverage data** — `test_gap` is import-graph
  reachability + name matching. It catches "this function isn't
  referenced by anything in `tests/`," not runtime line coverage.
- **`coupling` axis penalises high import fan-out**, which can be
  legitimate in framework-style codebases. Treat the absolute number
  as suggestive; the *delta* is what matters at PR time.

## Disabling

Just remove the workflow file. The action stores no state outside the
runner; no cleanup is required.
