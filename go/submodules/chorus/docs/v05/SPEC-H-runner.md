# Agent H — Phase Runner + Output Watcher

You own: `src/daemon/runner.ts` (NEW), `src/daemon/output-watcher.ts` (NEW), and edits to `src/daemon/index.ts`'s `/chats/:id/stream` SSE handler ONLY (do not touch other routes).

**Read first:** `docs/v05/SPEC-llm-shared.md`, then memory files there.

## Build

The runner is the conductor. Given a chat row, it walks phases in order, picks agents, asks the tmux manager for a session, drops a prompt file in the chat dir, watches for the answer, and emits events. Everything streams to the SSE handler.

### Filesystem layout (you own this)

```
~/.chorus/chats/<chatId>/
  meta.json                   # { chatId, work, templateId, createdAt }
  round-<N>/
    <role>-<agentName>/       # role = doer | reviewer
      ask.md                  # daemon writes
      answer.md               # CLI writes (chokidar watches)
      done                    # optional sentinel touched by CLI
      stderr.log              # capture-pane snapshots
```

When the chat starts, daemon writes `meta.json` and the round-1 `ask.md` for each role.

### `ask.md` structure

```markdown
# Chorus task — round <N>, phase <id>

## Your role
<role>: <doer | reviewer>

## What to do
<phase.title>
<phase.description>

## The user's request
<chat.work>

## Inputs (from prior phases)
<each include — ref to other phase's answer.md if include[] non-empty>

## Excluded (do NOT read)
<each exclude — explicit asymmetry, e.g. tests excluded from implement phase>

## How to respond
Write your full answer to: <absolute path to answer.md>
End with the sentinel: ## DONE
```

Keep it boring. The runner concatenates these sections from the phase + chat.

### `runner.ts`

```ts
export interface RunnerEvent {
  chatId: string;
  type: 'phase_start' | 'phase_progress' | 'phase_done' | 'phase_failed' | 'cli_error' | 'chat_done';
  payload: Record<string, unknown>;
  ts: number;
}

export interface PhaseRunnerOptions {
  chatId: string;
  template: Template;          // from src/lib/template-schema.ts
  work: string;
  onEvent: (e: RunnerEvent) => void;
  abortSignal: AbortSignal;
}

export async function runChat(opts: PhaseRunnerOptions): Promise<void>
```

Loop:
```
for phase of template.phases:
  for round of 1..phase.iterate.maxRounds:
    spawn doer:
      session = await tmux.acquire({ chatId, phaseId: phase.id, role: 'doer', round, share*, spawn: { lineage: phase.doer.lineage, ... } })
      shim = registry.pickShim(phase.doer.lineage)
      shim.preNudge?(...) → mgr.sendKeys(...)  // dismiss overlays, /clear opencode, etc.
      write ask.md
      mgr.pasteBuffer(session, shim.formatPrompt({ promptFile, answerFile, task: phase.title, expectDoneSentinel: true }))
      mgr.sendKeys(session, ['Enter'])
      onEvent({ type: 'phase_start', payload: { phase, role: 'doer', round, agent: shim.name } })
    wait for answer.md (chokidar in output-watcher) OR cli_error OR timeout (5 min default)
    onEvent({ type: 'phase_progress', payload: {...} })
    if doer succeeded: optionally fan out reviewers per phase.reviewer.candidates (parallel)
    consensus check: if all reviewers agree (>= phase.reviewer.require) → break round loop
    else: feed disagreement summary back into next round's ask.md
  onEvent({ type: 'phase_done' or 'phase_failed', payload: ... })
onEvent({ type: 'chat_done', payload: { verdict } })
```

For v0.5: parallel reviewer fan-out is allowed but optional. Sequential is fine if simpler. Document which you chose.

Cancellation: `abortSignal` aborts current waits, sends a polite Escape to the active session, then flips chat status to `cancelled` in the DB.

### `output-watcher.ts`

Use `chokidar`. Watch `~/.chorus/chats/<chatId>/**/answer.md` (and `done` sentinel files).

```ts
export function waitForAnswer(answerFile: string, opts: { timeoutMs: number; doneSentinel?: string }): Promise<{ content: string; full: boolean }>
```

Resolves when:
- `done` sentinel file appears (preferred), OR
- `answer.md` ends with the configured sentinel string (default `## DONE`), OR
- `answer.md` has been quiet for 90s after first write (matches openbridge folder transport timeout)

Rejects on `timeoutMs` reached (default 5 min — use `phase.iterate.maxRounds * 90s` if shorter).

### `index.ts` SSE handler — REPLACE the stub

Find the existing `fastify.get<{...}>('/chats/:id/stream', ...)` route. Replace its body:

```ts
fastify.get('/chats/:id/stream', async (request, reply) => {
  const chat = chats.getById(request.params.id);
  if (!chat) { reply.code(404).send({ ok: false, error: { code: 'not_found', message: 'chat not found' } }); return; }
  reply.raw.writeHead(200, {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache',
    Connection: 'keep-alive',
  });
  const ac = new AbortController();
  request.raw.on('close', () => ac.abort());
  // Look up template
  const tmplRow = templates.getById(chat.template_id);
  if (!tmplRow) { reply.raw.write(`data: ${JSON.stringify({ type: 'error', message: 'template missing' })}\n\n`); reply.raw.end(); return; }
  const template = TemplateSchema.parse(yaml.parse(tmplRow.yaml));
  await runChat({
    chatId: chat.id,
    template,
    work: chat.work,
    abortSignal: ac.signal,
    onEvent: (e) => {
      reply.raw.write(`data: ${JSON.stringify(e)}\n\n`);
      // also persist phase_events to DB so getChat() returns a populated history
      if (e.type === 'phase_start' || e.type === 'phase_done' || e.type === 'phase_failed') {
        phaseEvents.create({ chat_id: chat.id, ... });
      }
    },
  });
  reply.raw.end();
});
```

(If your existing tmux-based stub already populates phase_events, mimic that pattern.)

## Don't touch

- `src/daemon/tmux.ts` — Agent F
- `src/daemon/agents/` — Agent G
- `src/daemon/error-detector.ts` — Agent I
- Any other route handlers in `src/daemon/index.ts`
- UI files

## Setup

```bash
cd <your worktree>
pnpm install --prefer-offline
# chokidar should already be installable via pnpm; if not in deps, add it:
pnpm add chokidar
```

## Acceptance

```bash
pnpm typecheck                       # green
pnpm lint                            # green for your files
```

Manual smoke (you won't have F's tmux mgr or G's shims yet — stub them locally for your test, but DON'T commit those stubs; they're for your local validation only).

**Branch / commit:**
```
feat(runner): phase runner + output watcher + real SSE wiring (Agent H)
```
