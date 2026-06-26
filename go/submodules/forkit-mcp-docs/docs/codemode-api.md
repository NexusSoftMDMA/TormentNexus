# Codemode API Reference

All functions are available inside `execute_code` as `await codemode.<function>()`. Every call is automatically scoped to your workspace — there is no workspace ID parameter.

Get the TypeScript definitions at runtime by calling `search_api` once per session.

---

## Tasks

### `create_task(input)`

```typescript
create_task(input: {
  title: string;                  // max 500 chars
  description?: string;           // max 10,000 chars
  project_id?: string;
  sprint_id?: string;
  priority?: "low" | "medium" | "high" | "urgent";
  assignee?: string;              // agent identifier
  estimate?: number;              // story points or hours
  parent_id?: string | null;      // parent task ID for sub-tasks
  due_date?: string | null;       // ISO date string
  metadata?: Record<string, unknown>;
}): Promise<Task>
```

### `get_task(task_id)`

```typescript
get_task(task_id: string): Promise<Task | null>
```

### `list_tasks(filter?)`

```typescript
list_tasks(filter?: {
  project_id?: string;
  sprint_id?: string;
  status?: "pending" | "in_progress" | "done" | "cancelled";
  assignee?: string;
  priority?: "low" | "medium" | "high" | "urgent";
  limit?: number;   // default 50, max 500
  offset?: number;
}): Promise<Task[]>
```

### `update_task(task_id, updates)`

```typescript
update_task(
  task_id: string,
  updates: {
    title?: string;
    description?: string;
    status?: "pending" | "in_progress" | "done" | "cancelled";
    priority?: "low" | "medium" | "high" | "urgent";
    assignee?: string;
    estimate?: number;
    sprint_id?: string;
    project_id?: string;
    parent_id?: string | null;
    due_date?: string | null;
    metadata?: Record<string, unknown>;
  }
): Promise<Task>
```

### `delete_task(task_id)`

```typescript
delete_task(task_id: string): Promise<{ deleted: boolean }>
```

### `ready_tasks(filter?)`

```typescript
ready_tasks(filter?: {
  assignee?: string;
  limit?: number;  // default 20
}): Promise<Task[]>
```

Returns tasks where status is `pending` and all dependency blockers are `done` or `cancelled`. Topologically sorted — no need to walk the dependency graph.

### `claim_task(task_id, agent_id)`

```typescript
claim_task(
  task_id: string,
  agent_id: string
): Promise<{ task: Task; claimed: boolean }>
```

Atomic conditional write. `claimed: false` means another agent won the race — retry with a different task.

### `add_dependency(input)`

```typescript
add_dependency(input: {
  blocker_id: string;   // task that must finish first
  blocked_id: string;   // task that cannot start yet
  dependency_type?: "blocks" | "relates_to";
}): Promise<{ blocker_id: string; blocked_id: string }>
```

Both tasks must belong to the same workspace. Creating circular dependencies is rejected.

---

## Projects

### `create_project(input)`

```typescript
create_project(input: {
  name: string;         // max 200 chars
  description?: string; // max 10,000 chars
  icon?: string;
  kind?: string;        // free-form classification, e.g. "backend", "research", "infra"
}): Promise<Project>
```

### `list_projects(filters?)`

```typescript
list_projects(filters?: {
  status?: "active" | "archived";
  kind?: string;        // filter to a specific classification
}): Promise<Project[]>
```

### `get_project(project_id)`

```typescript
get_project(project_id: string): Promise<Project | null>
```

### `update_project(project_id, updates)`

```typescript
update_project(
  project_id: string,
  updates: {
    name?: string;
    description?: string;
    status?: "active" | "archived";
    icon?: string;
    kind?: string;
  }
): Promise<Project>
```

---

## Sprints

### `create_sprint(input)`

```typescript
create_sprint(input: {
  name: string;         // max 200 chars
  project_id: string;
  goal?: string;        // max 2,000 chars
  start_date?: string;  // ISO date string
  end_date?: string;    // ISO date string
}): Promise<Sprint>
```

### `list_sprints(project_id?)`

```typescript
list_sprints(project_id?: string): Promise<Sprint[]>
```

### `get_sprint(sprint_id)`

```typescript
get_sprint(sprint_id: string): Promise<Sprint | null>
```

### `get_active_sprint(project_id)`

```typescript
get_active_sprint(project_id: string): Promise<Sprint | null>
```

### `update_sprint(sprint_id, updates)`

```typescript
update_sprint(
  sprint_id: string,
  updates: {
    name?: string;
    goal?: string;
    start_date?: string;
    end_date?: string;
    status?: "planning" | "active" | "completed";
  }
): Promise<Sprint>
```

---

## Labels

### `create_label(input)`

```typescript
create_label(input: {
  name: string;   // max 100 chars
  color: string;  // hex format: "#ff6b35"
}): Promise<Label>
```

### `list_labels()`

```typescript
list_labels(): Promise<Label[]>
```

### `add_label_to_task(task_id, label_id)`

```typescript
add_label_to_task(task_id: string, label_id: string): Promise<void>
```

### `remove_label_from_task(task_id, label_id)`

```typescript
remove_label_from_task(task_id: string, label_id: string): Promise<void>
```

### `list_labels_for_task(task_id)`

```typescript
list_labels_for_task(task_id: string): Promise<Label[]>
```

---

## Comments

### `add_comment(input)`

```typescript
add_comment(input: {
  task_id: string;
  body: string;       // max 10,000 chars
  author: string;     // agent identifier
  author_type?: "agent" | "human";
}): Promise<Comment>
```

### `list_comments(task_id)`

```typescript
list_comments(task_id: string): Promise<Comment[]>
```

---

## Payments

### `list_payments(filter?)`

Read-only. Returns your workspace's x402 payment history.

```typescript
list_payments(filter?: {
  since?: number;   // Unix timestamp (seconds)
  limit?: number;   // default 100, max 500
}): Promise<Payment[]>
```

```typescript
type Payment = {
  id: string;
  tool: string;                   // which MCP tool triggered the payment
  amount: string;                 // USDC wei (6 decimals): "10000" = $0.01
  settlement_status: "pending" | "settled" | "failed" | "expired";
  on_chain_tx_hash?: string;      // Base mainnet tx once settled
  created_at: string;
};
```

---

## Trajectories

Git-like versioning of agent reasoning. Fork before a risky operation, revert if it fails, merge if it succeeds.

### `fork_trajectory(input)`

```typescript
fork_trajectory(input: {
  name: string;               // branch name (e.g. "refactor-attempt-1")
  from_execution_id?: string; // fork from a specific execution (defaults to HEAD)
}): Promise<Trajectory>
```

### `switch_trajectory(input)`

```typescript
switch_trajectory(input: {
  id: string;  // trajectory ID to switch HEAD to
}): Promise<{ ok: boolean }>
```

If `ok` is `false`, HEAD was changed by another agent — call `get_current_trajectory()` and retry.

### `list_trajectories()`

```typescript
list_trajectories(): Promise<Trajectory[]>
```

### `get_current_trajectory()`

```typescript
get_current_trajectory(): Promise<Trajectory | null>
```

```typescript
type Trajectory = {
  id: string;
  name: string;
  parent_id?: string;
  is_head: boolean;
  created_at: string;
};
```

---

## Sessions

Track agent work sessions with structured activity logs.

### `start_session(input)`

```typescript
start_session(input: {
  task_id: string;
  agent_id: string;
  notes?: string;
}): Promise<Session>
```

### `log_activity(input)`

```typescript
log_activity(input: {
  session_id: string;
  type: string;           // e.g. "tool_call", "decision", "error"
  body: string;
  execution_id?: string;
}): Promise<void>
```

### `complete_session(input)`

```typescript
complete_session(input: {
  session_id: string;
  status: string;         // e.g. "completed", "failed", "abandoned"
  notes?: string;
}): Promise<Session>
```

### `list_sessions(filter?)`

```typescript
list_sessions(filter?: {
  task_id?: string;
  agent_id?: string;
  status?: string;
  limit?: number;
}): Promise<Session[]>
```

```typescript
type Session = {
  id: string;
  task_id: string;
  agent_id: string;
  status: string;
  notes?: string;
  created_at: string;
  completed_at?: string;
};
```

---

## Webhooks

Register outbound webhooks to get notified when task events happen. Deliveries are signed with HMAC-SHA256.

### `register_webhook(input)`

```typescript
register_webhook(input: {
  url: string;
  events: string[];   // e.g. ["task.created", "task.updated", "task.done"]
  secret: string;     // used to sign deliveries via X-MCP-Signature header
}): Promise<Webhook>
```

### `list_webhooks()`

```typescript
list_webhooks(): Promise<Webhook[]>
```

### `delete_webhook(id)`

```typescript
delete_webhook(id: string): Promise<void>
```

```typescript
type Webhook = {
  id: string;
  url: string;
  events: string[];
  created_at: string;
};
```

Delivery payload: `{ event: string, task: Task, workspace_id: string, timestamp: string }`

Verify delivery: `X-MCP-Signature: sha256=<hmac>` computed over the raw request body with your secret.

---

## Type Reference

```typescript
type Task = {
  id: string;                     // "tk-a1b2c3"
  workspace_id: string;
  title: string;
  description?: string;
  status: "pending" | "in_progress" | "done" | "cancelled";
  priority?: "low" | "medium" | "high" | "urgent";
  assignee?: string;
  claimed_by?: string;
  estimate?: number;
  parent_id?: string;
  due_date?: string;
  project_id?: string;
  sprint_id?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

type Project = {
  id: string;
  name: string;
  description?: string;
  status: "active" | "archived";
  icon?: string;
  kind?: string;
  created_at: string;
  updated_at: string;
};

type Sprint = {
  id: string;
  project_id: string;
  name: string;
  goal?: string;
  start_date?: string;
  end_date?: string;
  status: "planning" | "active" | "completed";
  created_at: string;
};

type Label = {
  id: string;
  name: string;
  color: string;  // "#rrggbb"
  created_at: string;
};

type Comment = {
  id: string;
  task_id: string;
  author: string;
  author_type: "agent" | "human";
  body: string;
  created_at: string;
};
```
