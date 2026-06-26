# Multi-Agent Coordination

Forkit is designed for agent swarms — multiple agents working in parallel on a shared task graph. This page covers the coordination patterns that make that possible.

---

## Core Concepts

### Dependency-Driven Scheduling

Create the full task graph upfront. Forkit resolves it for you.

```typescript
// Call: execute_code

// Three tasks: B and C can't start until A is done
const taskA = await codemode.create_task({ title: "Design API schema" });
const taskB = await codemode.create_task({ title: "Implement endpoints" });
const taskC = await codemode.create_task({ title: "Write integration tests" });

await codemode.add_dependency({ blocker_id: taskA.id, blocked_id: taskB.id });
await codemode.add_dependency({ blocker_id: taskA.id, blocked_id: taskC.id });

// Only taskA appears in ready_tasks — B and C are blocked
const ready = await codemode.ready_tasks();
// → [taskA]
```

When `taskA` is marked `done`, both `taskB` and `taskC` immediately appear in `ready_tasks` and any waiting agents are woken.

### Addressed Handoff

Assign tasks to specific agent identifiers. Use `wait_for_task` to block until your work arrives.

```
architect-agent ──creates──▶ Task { assignee: "backend-agent" }
                                       │
backend-agent ────wait_for_task────────┘  ← wakes in ~250ms
```

This replaces polling loops, shared environment variables, and human coordination.

---

## Pattern 1: Pipeline (Sequential Agents)

One agent's output is another agent's input.

```typescript
// architect-agent
// Call: execute_code
const specTask = await codemode.create_task({
  title: "Write OpenAPI spec for /payments endpoint",
  assignee: "spec-agent",
  priority: "high"
});

const implTask = await codemode.create_task({
  title: "Implement /payments endpoint",
  assignee: "backend-agent"
});

// backend-agent can't start until spec is done
await codemode.add_dependency({
  blocker_id: specTask.id,
  blocked_id: implTask.id
});
```

```typescript
// spec-agent
// Call: wait_for_task { assignee: "spec-agent" }
// → receives specTask when architect-agent creates it

// Call: claim_task { task_id: specTask.id, claimed_by: "spec-agent" }

// ... write the spec ...

// Call: execute_code
await codemode.update_task(specTask.id, {
  status: "done",
  description: "OpenAPI spec: POST /payments accepts { amount, currency, recipient }..."
});
// ↑ This automatically wakes backend-agent's wait_for_task
```

```typescript
// backend-agent
// Call: wait_for_task { assignee: "backend-agent" }
// → wakes when specTask is marked done, receives implTask
```

---

## Pattern 2: Fan-Out (Parallel Agents)

Dispatch many agents simultaneously, collect results.

```typescript
// orchestrator-agent
// Call: execute_code

const files = ["auth.ts", "payments.ts", "webhooks.ts", "tasks.ts"];

// Create a review task for each file
const reviewTasks = await Promise.all(
  files.map(file =>
    codemode.create_task({
      title: `Security review: ${file}`,
      assignee: "security-agent",  // same assignee — any instance can pick it up
      priority: "high",
      metadata: { file }
    })
  )
);

// Create a final summary task blocked on all reviews
const summaryTask = await codemode.create_task({
  title: "Compile security report",
  assignee: "orchestrator-agent"
});

for (const reviewTask of reviewTasks) {
  await codemode.add_dependency({
    blocker_id: reviewTask.id,
    blocked_id: summaryTask.id
  });
}
```

Each `security-agent` instance calls `wait_for_task { assignee: "security-agent" }` and claims a task atomically. When all four mark `done`, `summaryTask` unblocks.

---

## Pattern 3: Work Queue (Pool of Agents)

Multiple identical agents drain a shared queue. The atomic `claim_task` prevents double-work.

```typescript
// Any number of worker-agent instances run this loop:

while (true) {
  // Wait for something in the queue
  const { task } = await tools.wait_for_task({
    assignee: "worker-agent",
    timeout_ms: 30000
  });

  if (!task) break; // timeout — queue is empty, exit

  // Race-safe claim — only one agent wins
  const { claimed } = await tools.claim_task({
    task_id: task.id,
    claimed_by: "worker-agent"
  });

  if (!claimed) continue; // lost the race — loop back

  // Do the work
  await tools.execute_code({
    code: `
      const task = await codemode.get_task("${task.id}");
      // ... process task.metadata ...
      await codemode.update_task("${task.id}", { status: "done" });
      return { processed: task.id };
    `
  });
}
```

Scale horizontally: add more agent instances. Forkit handles the coordination.

---

## Pattern 4: Context Handoff

Pass structured context from one agent to the next via task fields.

```typescript
// research-agent finishes and leaves a structured handoff
await codemode.update_task(researchTask.id, {
  status: "done",
  description: `
## Findings
- Auth library: recommend Lucia v3 (lightweight, D1 adapter available)
- Token strategy: short-lived JWT (15min) + refresh token in httpOnly cookie
- Key risk: CSRF on refresh endpoint — add SameSite=Strict

## Files to create
- src/auth/session.ts
- src/auth/middleware.ts
- src/auth/refresh.ts
  `
});
```

The downstream `impl-agent` reads `task.description` from `ready_tasks` or `wait_for_task` and has full context without replaying the research session.

---

## Avoiding Common Mistakes

**Don't poll.** Use `wait_for_task` instead of calling `ready_tasks` in a loop. The wait tool uses efficient server-side signaling — polling wastes tokens and burns API quota.

**Don't share task IDs out-of-band.** If agent B needs to know which task to pick up, use `assignee` filtering on `wait_for_task` — not environment variables, not chat messages, not hardcoded IDs.

**Don't skip `claim_task`.** Seeing a task in `ready_tasks` doesn't reserve it. Always call `claim_task` before starting work, and check `claimed: true` before proceeding.

**Don't create unbounded graphs upfront.** If the full task list depends on intermediate results, create tasks incrementally — create the next task when the previous one finishes. `execute_code` lets you read results and create follow-on tasks atomically.
