/**
 * Daemon-wide semaphore for local-CLI subprocesses.
 *
 * Two layers, both daemon-process-wide (NOT per-chat):
 *
 *   1. Global cap (`maxParallelCli`) — total number of local-CLI shim
 *      processes in flight across all chats.
 *
 *   2. Per-lineage cap (`perCli['opencode-cli']` etc.) — subset cap
 *      per binary family.
 *
 * `acquire(lineage)` blocks until BOTH a global slot AND a lineage slot
 * are available, then returns a `release()` function the caller MUST
 * invoke (use try/finally). Releasing returns the slots and wakes the
 * next waiter for whichever queue had backpressure.
 *
 * Settings are read dynamically per acquire so the user can tune the
 * caps in /settings and have it take effect on the next reviewer that
 * tries to start — no daemon restart needed.
 *
 * HTTP-dispatched shims (openrouter, future API-only shims) DO NOT call
 * acquire — they consume zero local CPU/RAM, just network. The
 * semaphore exists to prevent the user's box melting under N parallel
 * subprocess children.
 */

import {
  getConcurrency,
  resolvePerCliCap,
  type CliLineageKey,
} from '../lib/settings/concurrency.js';

interface Waiter {
  /** Resolves when this waiter has been granted both global + per-CLI slots. */
  resolve: () => void;
  /** Rejects when the caller's AbortSignal fires before grant. Caller is
   *  expected to treat this as "don't run the work" — same shape as a
   *  cancelled fetch. */
  reject: (err: unknown) => void;
  /** The lineage this waiter wants — used to route per-CLI capacity. */
  lineage: CliLineageKey;
  /** AbortSignal teardown so a cancelled chat doesn't leak its listener
   *  on the signal once we've granted the slot. */
  cleanup?: () => void;
}

/**
 * Module-level state. The daemon is a singleton process so this is the
 * authoritative cross-chat counter. Tests use `_resetForTests` to clear
 * between cases.
 */
let globalInFlight = 0;
const perCliInFlight: Map<CliLineageKey, number> = new Map();

/**
 * FIFO queue of waiters. We don't split per lineage because checking
 * whether a waiter at the head can proceed is cheap (two map lookups +
 * a settings read), and a single FIFO preserves fairness — a waiter
 * that came in first won't be starved by a later arrival of a different
 * lineage that happens to have free capacity.
 *
 * Trade-off: under heavy load with one bottlenecked lineage, the queue
 * head might block on perCli cap while later waiters of other lineages
 * could have run immediately. We accept this — fairness > throughput at
 * the scale chorus operates (single-user, ~10 chats max).
 */
const waiters: Waiter[] = [];

function getPerCli(lineage: CliLineageKey): number {
  return perCliInFlight.get(lineage) ?? 0;
}

function incPerCli(lineage: CliLineageKey): void {
  perCliInFlight.set(lineage, getPerCli(lineage) + 1);
}

function decPerCli(lineage: CliLineageKey): void {
  const next = getPerCli(lineage) - 1;
  if (next <= 0) {
    perCliInFlight.delete(lineage);
  } else {
    perCliInFlight.set(lineage, next);
  }
}

/**
 * Mutex flag preventing reentrant `tryGrantHead` execution. The
 * function does an `await getConcurrency()` (DB read) inside its loop,
 * so two concurrent callers (e.g. two near-simultaneous `release()`
 * events) would otherwise both pass the `waiters.length > 0` check,
 * both hit the await, both resume, and both try to shift the same
 * waiter — leading to undefined access and double-grant. With this
 * flag, the second caller bails immediately; the first caller's loop
 * will pick up any newly-released capacity before exiting.
 *
 * `dirty` is the companion: when a reentrant call bails, it sets
 * `dirty = true` to signal "you have new work to do." The live
 * tryGrantHead checks this in `finally` and re-enters if set. In
 * single-threaded JS this is technically belt-and-suspenders — there
 * is no yield between the live's `while`-exit and `granting = false`
 * for an external push to slip in — but it eliminates any ambiguity
 * under a future refactor that might introduce a yield point in that
 * window, and at most costs one extra tryGrantHead call per drain.
 */
let granting = false;
let dirty = false;

/**
 * Try to grant the head waiter (and subsequent waiters whose lineage
 * has free capacity). Called after every release and from `acquire()`.
 *
 * Walks from head: if the head can run, grant it and continue. If not
 * (their per-CLI cap is hit), STOP — strict FIFO. We deliberately do
 * NOT skip the head to grant a later waiter of a different lineage,
 * even if that later one has free capacity. Skipping would convert the
 * queue from "fair FIFO" into "lineage-affinity scheduler" which is
 * surprising under load.
 *
 * Errors (e.g. settings DB unavailable) are swallowed and logged —
 * propagating would land as an unhandled rejection in the
 * fire-and-forget call site in `release()`, crashing the daemon and
 * stranding the queue head forever.
 */
async function tryGrantHead(): Promise<void> {
  if (granting) {
    // A live tryGrantHead is mid-flight; mark dirty so it knows new
    // work arrived and re-runs before clearing the mutex.
    dirty = true;
    return;
  }
  granting = true;
  dirty = false;
  try {
    // Outer drain loop: re-runs the inner grant pass while any reentrant
    // call set `dirty = true`. Without this, an aborted-waiter splice
    // that pokes us mid-grant could be swallowed (the abort's
    // tryGrantHead call sees granting=true, bails, and we'd exit
    // without re-checking the queue).
    do {
      dirty = false;
      while (waiters.length > 0) {
        const config = await getConcurrency();
        // Defensive recheck — another release may have fired and emptied
        // the queue while we were awaiting the DB read.
        if (waiters.length === 0) break;
        if (globalInFlight >= config.maxParallelCli) break;

        const head = waiters[0];
        const perCliCap = resolvePerCliCap(config, head.lineage);
        if (getPerCli(head.lineage) >= perCliCap) break;

        waiters.shift();
        globalInFlight++;
        incPerCli(head.lineage);
        head.cleanup?.();
        head.resolve();
      }
      // If a reentrant call marked dirty (e.g. an abort spliced the
      // blocked head, freeing the queue), loop again so the new state
      // gets a chance to grant. Without this we'd exit while leaving
      // an eligible waiter unattended until the next external event.
    } while (dirty);
  } catch (err) {
    // Settings DB unreachable, corrupt, or any other unexpected failure.
    // Logged but never re-thrown — the void-call in release() can't
    // catch it, and an unhandled rejection here would crash the daemon
    // AND strand the queue head forever (subsequent releases would
    // re-trigger and silently fail again, creating a permanent
    // semaphore deadlock).
    console.error('[chorus] cli-semaphore tryGrantHead failed:', err);
  } finally {
    granting = false;
  }
}

/**
 * Acquire global + per-lineage slots. Blocks until both are free.
 * Returns a release function the caller MUST invoke (typically in a
 * `finally` block) to free the slots and wake the next waiter.
 *
 * Implementation note: every call enqueues, even when there's no
 * contention. Without this, two concurrent callers could both pass the
 * `waiters.length === 0 && globalInFlight < cap` check (separated by
 * the `await getConcurrency()` yield) and both increment, exceeding
 * the cap. Always-enqueue routes everything through `tryGrantHead`'s
 * mutex so increments are atomic w.r.t. the cap check. The extra
 * microtask cost is negligible at chorus's scale.
 *
 * Idempotent on release — calling release() twice is a no-op so a
 * caller that double-frees on an error path doesn't underflow the
 * counters.
 *
 * Cancellation: the optional `signal` lets a caller bail out of a
 * queued wait without leaving a stale waiter blocking the head. On
 * abort, the waiter is removed from the queue and the returned promise
 * rejects. If the slot was already granted before abort fires, the
 * caller is responsible for invoking the release function as usual —
 * we don't auto-release because the caller may already have spawned
 * the subprocess.
 */
export async function acquire(
  lineage: CliLineageKey,
  signal?: AbortSignal,
): Promise<() => void> {
  return new Promise<() => void>((resolve, reject) => {
    if (signal?.aborted) {
      reject(signal.reason ?? new Error('aborted'));
      return;
    }

    const waiter: Waiter = {
      lineage,
      resolve: () => resolve(makeRelease(lineage)),
      reject,
    };

    if (signal) {
      const onAbort = (): void => {
        const i = waiters.indexOf(waiter);
        if (i >= 0) waiters.splice(i, 1);
        // Also re-poke the queue: the aborted waiter may have been
        // blocking later waiters of a different lineage.
        void tryGrantHead();
        reject(signal.reason ?? new Error('aborted'));
      };
      signal.addEventListener('abort', onAbort, { once: true });
      waiter.cleanup = (): void => signal.removeEventListener('abort', onAbort);
    }

    waiters.push(waiter);
    void tryGrantHead();
  });
}

function makeRelease(lineage: CliLineageKey): () => void {
  let released = false;
  return () => {
    if (released) return;
    released = true;
    globalInFlight = Math.max(0, globalInFlight - 1);
    decPerCli(lineage);
    // Don't await — the caller doesn't care, and awaiting would couple
    // every release to the settings read inside tryGrantHead.
    // tryGrantHead has its own try/catch so a DB failure can't surface
    // as an unhandled rejection here.
    void tryGrantHead();
  };
}

/**
 * Diagnostic snapshot. Used by /api/v1/diagnostics for the health page.
 */
export function snapshot(): {
  globalInFlight: number;
  perCli: Record<string, number>;
  queueDepth: number;
} {
  return {
    globalInFlight,
    perCli: Object.fromEntries(perCliInFlight),
    queueDepth: waiters.length,
  };
}

/**
 * @internal — for tests only. Resets all in-flight counters and clears
 * the wait queue. Tests use this between cases since the module-level
 * state is otherwise sticky across the whole vitest worker.
 *
 * Pending waiters get rejected with a 'reset' error so a previous
 * test's queued acquire() doesn't leak into the next test as a hung
 * promise. Without this, awaiting a never-rejected acquire from a
 * prior test would silently hang the next test until vitest's timeout.
 */
export const _testing = {
  reset: (): void => {
    globalInFlight = 0;
    perCliInFlight.clear();
    granting = false;
    dirty = false;
    while (waiters.length > 0) {
      const w = waiters.shift();
      if (!w) break;
      w.cleanup?.();
      try {
        w.reject(new Error('cli-semaphore reset (tests)'));
      } catch {
        /* defensive — never let one waiter's reject handler block another */
      }
    }
  },
};
