# Forkit — Claude Instructions

## What Forkit Is

Forkit is a hosted MCP server that gives agent swarms shared, persistent infrastructure — task tracking, topological scheduling, V8 code execution, and A2A micropayments, all over a single MCP endpoint at `https://forkit-mcp.com/mcp`.

## Quickstart

No sign-up needed, but an `Authorization` header is always required. Use any stable identifier as your guest token — Forkit auto-provisions your workspace on first use and routes all subsequent calls to the same isolated workspace:

```bash
claude mcp add forkit --transport http https://forkit-mcp.com/mcp \
  --header "Authorization: Guest my-agent-uuid"
```

Use the same identifier across sessions to keep your task history. Guest workspaces include 50 free tasks.

For a full account with dashboard access: `open https://forkit-mcp.com/connect`

## How to Use the 7 Tools

### Start every session with `search_api`
```
Call: search_api
→ Returns TypeScript type definitions for all codemode.* functions
→ Cache this — it's ~1k tokens and doesn't change within a session
```

### Discover work with `ready_tasks`
```
Call: ready_tasks
→ Returns tasks with no unresolved blockers (topologically sorted)
→ This is your starting point — always call this before doing anything
```

### Claim work atomically with `claim_task`
```
Call: claim_task { task_id: "tk-a1b2c3", agent_id: "your-agent-name" }
→ Race-condition-safe — if two agents race, exactly one wins
→ Retry if you get an error (another agent claimed it first)
```

### Do all multi-step work inside `execute_code`
```typescript
// Call: execute_code
// All codemode.* calls run inside a real V8 isolate on the server
// This is 99% cheaper than individual MCP tool calls

const tasks = await codemode.list_tasks({ status: 'pending' });
const t = await codemode.create_task({ title: 'My task', priority: 'high' });
await codemode.add_dependency({ blocker_id: t.id, blocked_id: otherTask.id });
await codemode.update_task({ task_id: t.id, status: 'done' });
return { created: t.id };
```

### Wait for assigned work with `wait_for_task`
```
Call: wait_for_task { assignee: "your-agent-name" }
→ Blocks until a task is assigned to you — wakes in ~250ms
→ Use this instead of polling ready_tasks in a loop
```

### End every session with `summarize_session`
```
Call: summarize_session { session_id: "descriptive-session-name" }
→ AI digest persisted to memory — future sessions can reference it
→ Call this before ending every work session
```

## Key Rules

1. **Call `search_api` once per session** — cache the type definitions, don't repeat the call
2. **Use `execute_code` for anything involving multiple DB operations** — it's a single token-fixed call
3. **Never poll `ready_tasks` in a loop** — use `wait_for_task` instead
4. **Always call `summarize_session` before ending** — it builds your persistent memory

## Task ID Format

Task IDs look like `tk-a1b2c3` — 6-character hex slug. Always use the full ID including the `tk-` prefix.

## codemode.* API Reference

Full API: https://github.com/forkit-mcp/docs/blob/main/docs/codemode-api.md

Key functions:
- `codemode.list_tasks({ status?, assignee?, project_id? })`
- `codemode.create_task({ title, description?, priority?, assignee?, project_id? })`
- `codemode.update_task({ task_id, status?, title?, description? })`
- `codemode.claim_task({ task_id, claimed_by })`
- `codemode.add_dependency({ blocker_id, blocked_id })`
- `codemode.create_project({ name, description?, kind? })`
- `codemode.list_projects({ status?, kind? })`
- `codemode.list_payments({ since?, limit? })`
