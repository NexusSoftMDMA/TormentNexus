# Token Transparency

Every `mnem retrieve` response includes three counters that no other agent-memory system exposes:

| Field | What it tells you |
|---|---|
| `tokens_used` | How many tokens the returned results consume |
| `candidates_seen` | How many nodes were evaluated before the token budget was applied |
| `dropped` | How many candidates were cut at the budget boundary |

## Why this matters

When an agent retrieves memory, it has a token budget - context it can pass to the LLM. Most systems silently truncate at that boundary. You don't know what was cut or whether it was important.

mnem makes the cut explicit. If `dropped > 0`, the agent can widen the budget, tighten the query, or flag the situation to the user. The decision is visible, not hidden.

## Using the counters

In the CLI:
```bash
mnem retrieve "query" --limit 20
# Output includes: tokens_used: 3241 / candidates_seen: 47 / dropped: 12
```

Via MCP, the `_meta` field on every tool response carries `bytes`, `latency_micros`, and `tokens_estimate` so agents can reason about the cost of their own calls.

## See also

- [mnem retrieve flags](../src/cli.md)
- [MCP tool reference](../src/mcp.md)
- [GraphRAG](../../README.md#graphrag)
