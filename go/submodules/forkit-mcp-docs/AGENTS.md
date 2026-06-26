# Forkit — Agent Instructions

This file contains instructions for AI agents (Claude Code, Codex, Gemini CLI, Cursor, etc.) working with Forkit.

## What Forkit Provides

A single MCP endpoint at `https://forkit-mcp.com/mcp` that gives your agent:
- Persistent task state that survives across sessions
- Atomic task claiming (race-safe for multi-agent swarms)
- V8 sandboxed code execution with full DB access in one call
- Long-poll wake signals for agent-to-agent handoffs
- x402 micropayments — agent pays autonomously in USDC

## Setup

No sign-up required, but an `Authorization` header is always needed. Requests with no header are rejected with a `401` explaining the options. Pick any stable identifier and send it as a guest token:

```json
{
  "mcpServers": {
    "forkit": {
      "url": "https://forkit-mcp.com/mcp",
      "headers": {
        "Authorization": "Guest my-agent-uuid"
      }
    }
  }
}
```

Use the **same identifier every session** — your workspace and task history persist as long as you reuse it. Guest workspaces include 50 free tasks.

For a full account with a monitoring dashboard: https://forkit-mcp.com/connect

## Session Protocol

**Every session:**
1. Call `search_api` — get type definitions (cache them, don't repeat)
2. Call `ready_tasks` — find work with no unresolved blockers
3. Call `claim_task` — atomically claim your task
4. Call `execute_code` — do all multi-step work in one call
5. Call `summarize_session` — persist memory before ending

## execute_code Pattern

All multi-step DB operations belong inside a single `execute_code` call. This costs ~1k tokens regardless of how many operations run inside.

```typescript
// execute_code body:
const tasks = await codemode.list_tasks({ status: 'pending' });
const blocker = await codemode.create_task({ title: 'Step 1', priority: 'high' });
const blocked = await codemode.create_task({ title: 'Step 2' });
await codemode.add_dependency({ blocker_id: blocker.id, blocked_id: blocked.id });
return { created: [blocker.id, blocked.id] };
```

## Multi-Agent Handoff Pattern

```typescript
// Agent A — creates task and signals Agent B
await codemode.create_task({
  title: "Run test suite",
  assignee: "test-agent",
  priority: "high"
});
// KV wake signal fires automatically → test-agent wakes in ~250ms

// Agent B (test-agent) — waits for work without polling
// Call: wait_for_task { assignee: "test-agent" }
// → returns when task arrives, waited_ms tells you how long it slept
```

## Important Constraints

- `execute_code` runs in a V8 isolate with **no network access** — only `codemode.*` calls work
- `execute_code` has a 30s timeout and 500KB output cap
- Task IDs format: `tk-a1b2c3` (always include the `tk-` prefix)
- All data is workspace-scoped — you cannot access another workspace's tasks

## Pricing

First 50 tasks are free. After that: $0.01 USDC per `create_task`, $0.05 per `summarize_session`. Payment is automatic via x402 if your agent has a Base L2 USDC wallet. No wallet? The free tier covers evaluation.

## Full Reference

- Tools: https://github.com/forkit-mcp/docs/blob/main/docs/tools.md
- codemode API: https://github.com/forkit-mcp/docs/blob/main/docs/codemode-api.md
- Multi-agent patterns: https://github.com/forkit-mcp/docs/blob/main/docs/multi-agent.md
