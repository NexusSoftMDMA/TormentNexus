# Agent F — Tmux Manager + Reaper

You own: `src/daemon/tmux.ts` (REPLACE existing stub), `src/daemon/reaper.ts` (NEW).

**Read first:** `docs/v05/SPEC-llm-shared.md`, then the listed memory files there.

## Build

Implement the `TmuxManager` interface from `src/daemon/tmux-types.ts`. The existing `src/daemon/tmux.ts` is a stub from Agent A — replace it (keep the file path, the daemon imports from `./tmux`).

### Session naming

```
chorus-<chatId>-<phaseId>-<role>-<agentName>
```

Tmux session names allow [a-zA-Z0-9_-]. Sanitize chatId/phaseId/agentName to that charset before substitution.

### acquire() — the heart

Decision tree:

```
key = chatId + phaseId + role + agentName

if shareSessionAcrossRounds && session exists for (chatId, phaseId, role, agentName):
    update lastActivityAt → return existing handle
elif shareSessionAcrossPhases && session exists for (chatId, role, agentName) on a previous phase:
    rename it to current phaseId, update lastActivityAt → return existing handle
else:
    spawn fresh — `tmux new-session -d -s <name> "<launch>"`
    wait up to 8s for cold-start (poll `tmux has-session`)
    return new handle
```

Track sessions in an in-memory `Map<string, SessionHandle>` PLUS reconcile from `tmux ls` on startup (so daemon restarts don't lose sessions).

### Spawn flow

1. Build launch via `shim.buildLaunchCommand(spawn)` — caller passes the right shim.
2. **%q-quote any user/template values** before substitution. Validate names against `^[A-Za-z0-9_-]+$` upfront.
3. `tmux new-session -d -s <name> "<launch>"` — exec via Node `child_process.spawnSync('tmux', [...args])`. Pass args as an array, NOT one shell string, except for the launch command itself which IS a shell string by tmux design.
4. After spawn, poll `tmux has-session -t <name>` every 200ms up to 8s. Return when session is up.
5. On failure: throw `TmuxSpawnError` with code `cold_start_timeout` or `tmux_unavailable`.

### sendKeys / pasteBuffer

```ts
sendKeys(name: string, keys: string[]): void
  // tmux send-keys -t <name> <key1> <key2> ... (no Enter unless in keys[])

pasteBuffer(name: string, content: string): void
  // 1. tmux load-buffer -b chorus-<sessionName>-<pid> - <<< content
  // 2. tmux paste-buffer -b chorus-<sessionName>-<pid> -t <name>
  // Per-session buffer name avoids races between parallel chats
  // (lesson from openbridge_security_model.md)
```

### capturePane

```ts
capturePane(name: string): string
  // tmux capture-pane -t <name> -p -S -200
  // Returns last 200 lines of pane content as one string.
  // Used by the failure detector (Agent I).
```

### Reaper (`src/daemon/reaper.ts`)

```ts
export interface ReaperConfig {
  intervalMs: number;          // default 5 * 60 * 1000
  idleDestroyMinutes: number;  // default 30
}

export function startReaper(mgr: TmuxManager, getActiveChats: () => Map<string, string>, cfg: ReaperConfig): () => void {
  // Returns a stop() function for graceful shutdown.
  // setInterval(..., cfg.intervalMs):
  //   1. mgr.list() — all chorus-* sessions
  //   2. for each: if chatId NOT in activeChats (or status terminal) → mgr.kill()
  //   3. for each: if state==awaiting_user && lastActivityAt > idleDestroyMinutes ago → mgr.kill()
  //   4. log a one-liner summary { killed: [names] }
}
```

Wire from `src/daemon/index.ts` — call `startReaper(...)` after startup; `stop()` on SIGTERM.

## Don't touch

- `src/daemon/agents/` — Agent G
- `src/daemon/runner.ts`, `src/daemon/output-watcher.ts` — Agent H
- `src/daemon/error-detector.ts` — Agent I
- Any UI files

## Setup

```bash
cd <your worktree>
pnpm install --prefer-offline
```

## Acceptance

```bash
pnpm typecheck                       # green
pnpm lint                            # green for your files
```

Manual: spawn a real session in a unit test or smoke harness:
```bash
node --import tsx -e "
  import { TmuxManagerImpl } from './src/daemon/tmux.ts';
  const mgr = new TmuxManagerImpl();
  // ... acquire a fake session, capturePane, kill, etc.
"
```

Document any test you wrote in your final commit message.

**Branch / commit:**
```
feat(tmux): real tmux manager + reaper with share-session policy (Agent F)
```
Push and report the branch name + file count.
