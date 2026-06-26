# Core Concepts

---

## The Reasoning Layer

Standard AI agents are stateless — each session starts from scratch. Persistent memory tools (Claude Projects, ChatGPT memory) solve single-agent context. They don't solve coordination.

When multiple agents work in parallel — or when an agent needs to resume a task from a previous session — you need something different: shared, structured state that all agents can read and write atomically, with dependency tracking and work assignment built in.

Forkit is that layer. It sits between your agents and your codebase.

```
Claude Code ─────────────────────────┐
Codex        ──── POST /mcp ─────────┼──▶ Forkit ──▶ D1 (tasks, projects, history)
Custom agent ─────────────────────────┘              R2 (session memory)
```

---

## execute_code-First Design

Most MCP task trackers expose one tool per operation: `create_task`, `update_task`, `list_tasks`, etc. Each tool call costs tokens — tool definition, input, output. A 10-step workflow costs 10 round-trips.

Forkit takes a different approach. The entire DB surface is accessible inside a single `execute_code` call via `codemode.*` async functions. The LLM fetches type definitions once (~1k tokens via `search_api`), then expresses arbitrarily complex workflows in a single tool call.

| Approach | Tokens per 10-step workflow |
|----------|-----------------------------|
| Standard MCP (one tool per op) | ~150,000 |
| Forkit execute_code | ~1,000 |

The sandbox runs in a real V8 isolate — not `eval`, not a string interpreter. It has access to `async/await`, `Promise.all`, and the full codemode API. Network access is disabled inside the sandbox (all I/O goes through `codemode.*`), so there's no way for agent code to exfiltrate data or make unauthorized calls.

---

## Atomic Task Claiming

When multiple agents compete for the same task, exactly one should win. Forkit's `claim_task` uses a conditional database write:

```sql
UPDATE tasks
SET status = 'in_progress', claimed_by = ?
WHERE id = ? AND status = 'pending' AND claimed_by IS NULL
```

If the row was already claimed, `changes === 0` and the caller gets `claimed: false`. No locks, no queues — a single atomic write resolves the race.

---

## Topological Scheduling

`ready_tasks` returns only tasks whose dependency blockers are all resolved:

```sql
SELECT * FROM tasks
WHERE workspace_id = ?
  AND status = 'pending'
  AND id NOT IN (
    SELECT blocked_id FROM task_dependencies
    JOIN tasks AS blocker ON blocker.id = task_dependencies.blocker_id
    WHERE blocker.status NOT IN ('done', 'cancelled')
  )
```

You never walk the dependency graph manually. Create the full task graph, call `ready_tasks`, and let the scheduler hand out work as it becomes available.

---

## Addressed Wake Signals

`wait_for_task` blocks efficiently using KV-signaled polling. When a task is created, claimed, updated, or when a dependency resolves, Forkit writes a wake signal to:

- `relay:wake:{workspaceId}:*` — catch-all (all waiting agents)
- `relay:wake:{workspaceId}:{assignee}` — addressed to a specific agent

A waiting `wait_for_task` call polls every 500ms, checking its assignee key first. Median wake latency is ~250ms. No busy-loops, no webhooks needed.

External systems can also trigger a wake via:

```
POST https://forkit-mcp.com/relay/wake
Authorization: Bearer <api-key>
Content-Type: application/json

{ "assignee": "my-agent" }
```

---

## x402 Agent-Native Payments

Forkit uses the [x402 payment protocol](https://x402.org) for metered tool calls. Your agent's crypto wallet is its identity — no credit card, no subscription.

When a gated tool is called (e.g., `create_task` after your free 50 tasks), Forkit returns an HTTP 402 with a payment quote. A compatible x402 agent wallet pays the quote in USDC on Base L2. Forkit verifies the EIP-3009 `TransferWithAuthorization` signature locally (no external dependency) and proceeds.

Payments are settled on-chain automatically every 10 minutes via a Cloudflare Cron trigger. Every payment is recorded in your workspace's `payments` table, readable via `codemode.list_payments()`.

**If you don't have an x402 wallet yet:** the first 50 tasks per workspace are free. Most evaluation workflows stay within that limit.

---

## Workspace Isolation

Every piece of data — tasks, projects, sprints, labels, comments, executions, payments — is scoped to a workspace. Every DB query includes `WHERE workspace_id = ?`. There is no admin view, no cross-tenant query, no shared namespace.

Your API key maps to exactly one workspace. Workspaces are created automatically — either when you sign in with GitHub, or instantly when you use `Authorization: Guest <identifier>` (no sign-up required).
