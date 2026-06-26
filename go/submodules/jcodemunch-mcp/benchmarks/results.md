# jcodemunch-mcp -- Token Efficiency Benchmark

**Tokenizer:** `cl100k_base` (tiktoken)  
**Workflow:** `search_symbols` (top 5) + `get_symbol_source` x 3
**Baseline:** all source files concatenated (minimum for "open every file" agent)  

## expressjs/express

| Metric | Value |
|--------|-------|
| Files indexed | **165** |
| Symbols extracted | **181** |
| Baseline tokens (all files) | **137,978** |

| Query | Baseline&nbsp;tokens | jMunch&nbsp;tokens | Reduction | Ratio |
|-------|---------------------:|-------------------:|----------:|------:|
| `router route handler` | 137,978 | 886 | **99.4%** | 155.7x |
| `middleware` | 137,978 | 1,008 | **99.3%** | 136.9x |
| `error exception` | 137,978 | 859 | **99.4%** | 160.6x |
| `request response` | 137,978 | 872 | **99.4%** | 158.2x |
| `context bind` | 137,978 | 993 | **99.3%** | 139.0x |
| **Average** | — | — | **99.4%** | **150.1x** |

<details><summary>Query detail (search + fetch tokens, latency)</summary>

| Query | Search&nbsp;tokens | Fetch&nbsp;tokens | Hits&nbsp;fetched | Search&nbsp;ms |
|-------|-----------------:|------------------:|------------------:|---------------:|
| `router route handler` | 381 | 505 | 3 | 6.2 |
| `middleware` | 370 | 638 | 3 | 0.2 |
| `error exception` | 362 | 497 | 3 | 7.1 |
| `request response` | 372 | 500 | 3 | 0.5 |
| `context bind` | 372 | 621 | 3 | 0.3 |

</details>

## fastapi/fastapi

| Metric | Value |
|--------|-------|
| Files indexed | **951** |
| Symbols extracted | **5,325** |
| Baseline tokens (all files) | **699,425** |

| Query | Baseline&nbsp;tokens | jMunch&nbsp;tokens | Reduction | Ratio |
|-------|---------------------:|-------------------:|----------:|------:|
| `router route handler` | 699,425 | 1,199 | **99.8%** | 583.3x |
| `middleware` | 699,425 | 1,643 | **99.8%** | 425.7x |
| `error exception` | 699,425 | 873 | **99.9%** | 801.2x |
| `request response` | 699,425 | 4,439 | **99.4%** | 157.6x |
| `context bind` | 699,425 | 1,016 | **99.9%** | 688.4x |
| **Average** | — | — | **99.8%** | **531.2x** |

<details><summary>Query detail (search + fetch tokens, latency)</summary>

| Query | Search&nbsp;tokens | Fetch&nbsp;tokens | Hits&nbsp;fetched | Search&nbsp;ms |
|-------|-----------------:|------------------:|------------------:|---------------:|
| `router route handler` | 464 | 735 | 3 | 131.4 |
| `middleware` | 460 | 1,183 | 3 | 0.3 |
| `error exception` | 383 | 490 | 3 | 0.7 |
| `request response` | 430 | 4,009 | 3 | 12.0 |
| `context bind` | 402 | 614 | 3 | 0.4 |

</details>

## gin-gonic/gin

| Metric | Value |
|--------|-------|
| Files indexed | **98** |
| Symbols extracted | **1,489** |
| Baseline tokens (all files) | **187,018** |

| Query | Baseline&nbsp;tokens | jMunch&nbsp;tokens | Reduction | Ratio |
|-------|---------------------:|-------------------:|----------:|------:|
| `router route handler` | 187,018 | 1,151 | **99.4%** | 162.5x |
| `middleware` | 187,018 | 1,130 | **99.4%** | 165.5x |
| `error exception` | 187,018 | 818 | **99.6%** | 228.6x |
| `request response` | 187,018 | 1,083 | **99.4%** | 172.7x |
| `context bind` | 187,018 | 1,436 | **99.2%** | 130.2x |
| **Average** | — | — | **99.4%** | **171.9x** |

<details><summary>Query detail (search + fetch tokens, latency)</summary>

| Query | Search&nbsp;tokens | Fetch&nbsp;tokens | Hits&nbsp;fetched | Search&nbsp;ms |
|-------|-----------------:|------------------:|------------------:|---------------:|
| `router route handler` | 440 | 711 | 3 | 43.5 |
| `middleware` | 371 | 759 | 3 | 5.7 |
| `error exception` | 345 | 473 | 3 | 0.8 |
| `request response` | 387 | 696 | 3 | 0.8 |
| `context bind` | 391 | 1,045 | 3 | 7.9 |

</details>

---

## Real-world A/B test: naming audit task (2026-03-18)

50-iteration test by @Mharbulous comparing JCodeMunch vs native tools (Grep/Glob/Read) on a real Vue 3 + Firebase production codebase. Full report: [ab-test-naming-audit-2026-03-18.md](ab-test-naming-audit-2026-03-18.md)

| Metric | Native | JCodeMunch | Delta |
|--------|--------|------------|-------|
| Success rate | 72% | 80% | +8 pp |
| Timeout rate | 40% | 32% | −8 pp |
| Mean cost/iteration | $0.783 | $0.738 | −5.7% |
| Mean cache creation | 104,135 | 93,178 | −10.5% |

Tool-layer savings (isolated from fixed overhead): **15–25%**

---

## Real-world A/B test: dead code detection task (2026-03-18)

50-iteration test by @Mharbulous comparing JCodeMunch vs native tools on the same Vue 3 + Firebase codebase. Designed to isolate pure tool-layer cost with no subagent overhead. Full report: [ab-test-dead-code-2026-03-18.md](ab-test-dead-code-2026-03-18.md)

| Metric | Native | JCodeMunch | Delta |
|--------|--------|------------|-------|
| Success rate | 96% | 92% | −4 pp |
| Mean cost/iteration | $0.4474 | $0.3560 | −20.0% |
| Mean total tokens | 449,356 | 289,275 | −36% |
| Mean duration (s) | 129 | 117 | −9% |
| File-level F1 (dead files) | 95.8% | 95.7% | equivalent |
| File-level F1 (alive files) | 100.0% | 69.6% | gap |
| Export-level F1 | 93.3% | 64.1% | gap |

**Confirmed tool-layer savings: 20%** (statistically significant, Wilcoxon p=0.0074). Dead file detection is equivalent. Accuracy gaps identified on alive-file classification and export-level analysis; three root causes found and addressed (see report).

Raw data: https://gist.github.com/Mharbulous/bb097396fa92ef1d34d03a72b56b2c61

---

## Grand Summary

| | Tokens |
|--|-------:|
| Baseline total (15 task-runs) | 5,122,105 |
| jMunch total | 19,406 |
| **Reduction** | **99.6%** |
| **Ratio** | **263.9x** |

> Measured with tiktoken `cl100k_base`. Baseline = all indexed source files. jMunch = search_symbols (top 5) + get_symbol_source x 3 per query.