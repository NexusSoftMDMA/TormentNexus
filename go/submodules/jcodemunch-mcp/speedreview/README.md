# speedreview

AI code review in under 5 seconds. Powered by [jCodeMunch](https://github.com/jgravelle/jcodemunch-mcp) + [Groq](https://groq.com).

> Your code reviewer is slower than your linter.

Uses jCodeMunch's token-efficient code retrieval (AST parsing, symbol analysis, blast radius) combined with Groq's ultra-fast inference (280-1000 tok/s) to post a structured review as a PR comment — before your CI checks even start.

## Example Output

```markdown
## speedreview (3.2s)

### Summary
This PR adds retry logic to the HTTP client with exponential backoff.

### Issues Found
- **[High]** `retry_count` has no upper bound — infinite loop risk (src/client.py:42)
- **[Medium]** Missing timeout on the retry delay — could block event loop (src/client.py:58)

### Impact Analysis
3 downstream callers affected: `fetch_user()`, `fetch_repo()`, `sync_data()`.
No breaking signature changes detected.

---
*Powered by [jCodeMunch](https://github.com/jgravelle/jcodemunch-mcp) + [Groq](https://groq.com) | Review completed in 3.2s*
```

## Setup

### 1. Get a Groq API key

Sign up at [console.groq.com](https://console.groq.com) and create an API key.

### 2. Add the secret to your repo

Go to **Settings > Secrets and variables > Actions** and add `GROQ_API_KEY`.

### 3. Create the workflow

```yaml
# .github/workflows/speedreview.yml
name: speedreview
on: [pull_request]

permissions:
  pull-requests: write
  contents: read

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: jgravelle/jcodemunch-mcp/speedreview@v1.108.22
        with:
          groq_api_key: ${{ secrets.GROQ_API_KEY }}
```

### Pinning for production

The example above pins the action to a specific package release tag (`@v1.108.22`),
which is the recommended baseline. For workflows under stricter supply-chain
review, pin to the commit SHA the tag points to instead:

```yaml
      - uses: jgravelle/jcodemunch-mcp/speedreview@<full-40-char-sha>
        with:
          groq_api_key: ${{ secrets.GROQ_API_KEY }}
```

Resolve the SHA with `git ls-remote https://github.com/jgravelle/jcodemunch-mcp refs/tags/v1.108.22`.

SHA pinning makes the consumed action immutable: if a tag is ever moved or a
breaking change ships on `main`, the workflow keeps running the same code
until you intentionally bump the SHA. The `@main` form (older docs) is no
longer recommended for production workflows.

The action also pins its installed Python packages by default
(`jcodemunch-mcp==1.108.20`, `openai>=1.50,<2`). Override those defaults via
the `jcodemunch_version` and `openai_version` inputs if your environment
requires a different pin.

## Configuration

| Input | Default | Description |
|-------|---------|-------------|
| `groq_api_key` | **(required)** | Groq API key |
| `model` | `llama-3.3-70b-versatile` | Groq model for generating the review |
| `severity_threshold` | `low` | Minimum severity to include: `low`, `medium`, `high` |
| `max_comment_length` | `4000` | Maximum PR comment length in characters |
| `token_budget` | `8000` | Token budget for jCodeMunch context retrieval |
| `base_ref` | *(auto-detect)* | Base ref to diff against |
| `jcodemunch_version` | `==1.108.22` | PyPI version specifier for jcodemunch-mcp |
| `openai_version` | `>=1.50,<2` | PyPI version specifier for the openai SDK |

### Model options

| Model | Speed | Best for |
|-------|-------|----------|
| `llama-3.3-70b-versatile` | 280 tps | Best review quality (default) |
| `openai/gpt-oss-120b` | 500 tps | Complex reasoning |
| `openai/gpt-oss-20b` | 1000 tps | Maximum speed |
| `llama-3.1-8b-instant` | 560 tps | Rate-limit friendly |

## How It Works

```
PR opened/updated
     |
     v
1. git diff -> changed files + hunks
2. jCodeMunch indexes the repo (cached between runs)
3. get_changed_symbols -> what changed at the symbol level
4. get_blast_radius -> downstream impact analysis
5. get_ranked_context -> relevant surrounding code (token-budgeted)
6. Groq inference -> structured review (sub-2s)
7. Post as PR comment
     |
     v
Review appears in < 5 seconds (cache hit)
```

- **jCodeMunch runs locally** in the Action — no network hop, works on private repos
- **Index is cached** via GitHub Actions cache — subsequent PRs skip indexing
- **Groq is the only external call** — one API request, sub-2s inference
- **Updates existing comment** on force-push — no comment spam

## What It Reviews

Focuses on: bugs, security vulnerabilities, performance problems, missing error handling, breaking API changes.

Ignores: style, formatting, naming, docs, test coverage.

## Cost

- **Groq API**: ~$0.001-0.01 per review depending on diff size and model
- **GitHub Actions**: Standard runner minutes (typically 10-30s per review)
