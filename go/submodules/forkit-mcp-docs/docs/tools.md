# MCP Tools Reference

Forkit exposes 7 tools over a standard MCP connection. All write operations go through `execute_code` — the other tools handle coordination and discovery.

---

## execute_code

**The primary tool.** Runs async JavaScript in a real V8 isolate (Cloudflare Worker sandbox). All database operations are available as `await codemode.*` calls inside the script.

**Why one tool instead of many?** Token cost is fixed regardless of how many DB operations your agent performs. Call `search_api` once to get type definitions (~1k tokens), then express entire multi-step workflows in a single `execute_code` call instead of ~150k tokens of individual tool round-trips.

### Input

| Field | Type | Description |
|-------|------|-------------|
| `code` | `string` | Async JavaScript to execute. Must be a valid async function body. |

### Output

```typescript
{
  result: unknown;      // Return value of your script
  logs: string[];       // console.log() output captured during execution
  duration_ms: number;  // Wall-clock execution time
}
```

### Constraints

- **No network access** inside the sandbox — all external I/O goes through `codemode.*`
- **500KB output cap** — result + logs serialized; exceeding this returns an error
- All `codemode.*` calls are workspace-scoped automatically — no cross-tenant access possible

### Example

```typescript
// Create a project, add tasks with dependencies, return the full graph
const project = await codemode.create_project({
  name: "API Refactor",
  description: "Break monolith into services"
});

const designTask = await codemode.create_task({
  title: "Design service boundaries",
  project_id: project.id,
  priority: "high",
  assignee: "architect-agent"
});

const implTask = await codemode.create_task({
  title: "Implement auth service",
  project_id: project.id,
  priority: "high",
  assignee: "backend-agent"
});

// implTask can't start until designTask is done
await codemode.add_dependency({
  blocker_id: designTask.id,
  blocked_id: implTask.id
});

return { project, tasks: [designTask, implTask] };
```

---

## search_api

Returns the full TypeScript type definitions for the `execute_code` codemode API. Call this once at the start of a session and cache the result — the definitions don't change between calls.

### Input

None.

### Output

A string containing TypeScript interface and function signature definitions for every `codemode.*` function available inside `execute_code`.

### Usage pattern

```
1. Call search_api → get type definitions
2. Parse / cache the definitions
3. Use codemode.* inside every subsequent execute_code call
```

---

## ready_tasks

Returns the set of tasks that are ready to start: `pending` status, all dependency blockers resolved (`done` or `cancelled`), and optionally filtered to a specific assignee.

Uses a topological sort — you never need to walk the dependency graph yourself. Call `ready_tasks` to get your work queue.

### Input

| Field | Type | Description |
|-------|------|-------------|
| `assignee` | `string` (optional) | Filter to tasks assigned to a specific agent |
| `limit` | `number` (optional) | Max tasks to return (default 20) |

### Output

```typescript
Task[]  // Array of tasks with no unresolved blockers
```

### Pricing

Free for the first 50 tasks in your workspace. $0.01 USDC/call after that (x402).

---

## claim_task

Atomically assigns a task to an agent. Race-condition-safe — if two agents call `claim_task` for the same task simultaneously, exactly one succeeds and the other receives an error.

### Input

| Field | Type | Description |
|-------|------|-------------|
| `task_id` | `string` | ID of the task to claim (format: `tk-a1b2c3`) |
| `agent_id` | `string` | Identifier for the claiming agent (e.g. `"claude-code"` or a UUID) |
| `enforce_assignee` | `boolean` | Optional. When `true`, rejects claim if task assignee doesn't match `agent_id` |

### Output

```typescript
{
  task: Task;      // The claimed task with updated status
  claimed: boolean; // true if this agent won the race
}
```

### Coordination pattern

```typescript
// Get ready tasks
const ready = await tools.ready_tasks({ assignee: "my-agent" });

// Claim the first one atomically
const { task, claimed } = await tools.claim_task({
  task_id: ready[0].id,
  agent_id: "my-agent"
});

if (!claimed) {
  // Another agent got there first — retry with next task
}
```

---

## wait_for_task

Long-polls until a task matching the given `assignee` becomes available (pending + no blockers). Blocks for up to the specified timeout, waking within ~250ms when a matching task appears.

Use this for agent-to-agent handoffs instead of polling loops.

### Input

| Field | Type | Description |
|-------|------|-------------|
| `assignee` | `string` | Agent identifier to wait for work on |
| `timeout_ms` | `number` (optional) | Max wait in milliseconds (default 30000, max 120000) |

### Output

```typescript
{
  task: Task | null;   // The available task, or null if timeout elapsed
  waited_ms: number;   // How long the call actually waited
}
```

### Wake signals

`wait_for_task` wakes automatically when:
- A new task is created with a matching `assignee`
- An existing task is updated to match (status change, assignee change)
- A blocker task is marked `done` or `cancelled`, unblocking a downstream task

External systems can also trigger a wake via `POST /relay/wake` with `{ assignee: "agent-name" }`.

### Example

```typescript
// test-agent waits for work — median wake latency ~250ms
const { task, waited_ms } = await tools.wait_for_task({
  assignee: "test-agent",
  timeout_ms: 60000
});

if (task) {
  await tools.claim_task({ task_id: task.id, claimed_by: "test-agent" });
  // ... do the work
}
```

---

## summarize_session

Generates an AI digest of the current session (recent tasks, executions, decisions) and writes it to persistent R2 storage as a Markdown file. Future sessions can read this file to resume context without replaying the full execution history.

### Input

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | `string` | Identifier for this session (used as the memory file key) |
| `notes` | `string` (optional) | Additional context to include in the summary |

### Output

```typescript
{
  summary: string;    // The generated Markdown summary
  memory_key: string; // R2 key where the summary was written
}
```

### Pricing

$0.05 USDC/call (x402). Powered by llama-3.1-8b-instruct via Cloudflare AI.

---

## list_executions

Returns recent `execute_code` audit records for your workspace. Every execution — success or error — is logged automatically. Use this to debug failed runs, track agent activity, or build observability tooling on top.

### Input

| Field | Type | Description |
|-------|------|-------------|
| `limit` | `number` (optional) | Max records to return (default 20, max 100) |
| `status` | `"success" \| "error"` (optional) | Filter by outcome |

### Output

```typescript
{
  id: string;
  status: "success" | "error";
  duration_ms: number;
  error?: string;          // Error message if status === "error"
  created_at: string;      // ISO timestamp
}[]
```

### Example

```typescript
// Find recent failures
const failures = await tools.list_executions({ status: "error", limit: 10 });
// Each record includes the error message — no need to dig through logs manually
```
