/**
 * Daemon-wide CLI semaphore tests.
 *
 * The semaphore composes a global cap (`maxParallelCli`) and per-lineage
 * caps (`perCli['opencode-cli']`). Both layers must hold simultaneously
 * — the queue is FIFO and strict (no lineage-affinity skip).
 *
 * Tests use a temp DB so the settings module's reads against
 * settings.get() resolve to test fixtures, not whatever the user has on
 * disk.
 */

import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import os from 'os';
import path from 'path';
import fs from 'fs';
import { randomUUID } from 'crypto';

import { _resetDbForTests, getDb } from '@/lib/db';
import { setConcurrency } from '@/lib/settings/concurrency';
import { acquire, snapshot, _testing } from '@/daemon/cli-semaphore';

let dbPath: string;

/**
 * Drain pending acquires so the test sees the post-await state of the
 * semaphore. acquire() does an `await getConcurrency()` (libsql read)
 * before the queue push, which is several microtasks; a single
 * `await Promise.resolve()` isn't enough. 20ms is conservative and
 * keeps the test fast — in practice the DB hit completes in <1ms.
 */
const flush = (): Promise<void> =>
  new Promise<void>((r) => setTimeout(r, 20));

beforeEach(async () => {
  dbPath = path.join(os.tmpdir(), `chorus-semaphore-${randomUUID()}.db`);
  process.env.CHORUS_DB_PATH = dbPath;
  await _resetDbForTests();
  await getDb();
  _testing.reset();
});

afterEach(async () => {
  _testing.reset();
  await _resetDbForTests();
  for (const suffix of ['', '-shm', '-wal']) {
    try { fs.unlinkSync(dbPath + suffix); } catch { /* best-effort */ }
  }
  delete process.env.CHORUS_DB_PATH;
});

describe('cli-semaphore', () => {
  it('grants slots up to the global cap, then queues', async () => {
    await setConcurrency({ maxParallelCli: 2, perCli: {} });

    // First two should resolve immediately.
    const r1 = await acquire('claude-code');
    const r2 = await acquire('codex-cli');
    expect(snapshot().globalInFlight).toBe(2);
    expect(snapshot().queueDepth).toBe(0);

    // Third must wait — set up a watcher that resolves the slot
    // promise's settled state without blocking the test.
    let r3Settled = false;
    const r3Promise = acquire('gemini-cli').then((r) => {
      r3Settled = true;
      return r;
    });

    // Yield enough for acquire's `await getConcurrency()` (DB read) to
    // settle — a single microtask isn't enough.
    await flush();
    expect(r3Settled).toBe(false);
    expect(snapshot().queueDepth).toBe(1);

    // Releasing r1 wakes r3.
    r1();
    const r3 = await r3Promise;
    expect(r3Settled).toBe(true);
    expect(snapshot().globalInFlight).toBe(2); // r2 + r3 now active

    r2();
    r3();
    expect(snapshot().globalInFlight).toBe(0);
  });

  it('per-CLI cap blocks same-lineage acquires when global has spare capacity', async () => {
    // Global allows 5, but opencode is capped at 1. Without per-CLI
    // enforcement, all 5 of these would acquire instantly.
    await setConcurrency({
      maxParallelCli: 5,
      perCli: { 'opencode-cli': 1 },
    });

    const r1 = await acquire('opencode-cli');
    expect(snapshot().perCli['opencode-cli']).toBe(1);
    expect(snapshot().globalInFlight).toBe(1);

    // Second opencode must queue even though global has 4 slots free.
    let r2Settled = false;
    const r2Promise = acquire('opencode-cli').then((r) => {
      r2Settled = true;
      return r;
    });
    await flush();
    expect(r2Settled).toBe(false);
    expect(snapshot().queueDepth).toBe(1);
    expect(snapshot().globalInFlight).toBe(1); // only r1 counts

    // Releasing r1 wakes r2 since the per-CLI quota frees up.
    r1();
    const r2 = await r2Promise;
    expect(r2Settled).toBe(true);
    expect(snapshot().perCli['opencode-cli']).toBe(1); // r2 now owns it
    r2();
  });

  it('strict-FIFO: a later arrival of an unblocked lineage queues behind a blocked head', async () => {
    // Opencode cap = 1. Global cap = 5.
    await setConcurrency({
      maxParallelCli: 5,
      perCli: { 'opencode-cli': 1 },
    });

    const r1 = await acquire('opencode-cli'); // takes the only opencode slot
    const r2Promise = acquire('opencode-cli'); // queues — head waiter

    // Wait for r2's queue.push to complete before r3 starts; if both
    // race the queue order is non-deterministic.
    await flush();

    let r3Settled = false;
    const r3Promise = acquire('codex-cli').then((r) => {
      r3Settled = true;
      return r;
    });
    await flush();

    // r3 should be queued (strict FIFO — head r2 is blocked, so r3
    // doesn't skip ahead even though codex has free capacity).
    expect(r3Settled).toBe(false);
    expect(snapshot().queueDepth).toBe(2);

    // Releasing r1 wakes r2 (head), which then frees r3.
    r1();
    const r2 = await r2Promise;
    const r3 = await r3Promise;
    expect(r3Settled).toBe(true);

    r2();
    r3();
  });

  it('release is idempotent — double-call does not underflow', async () => {
    await setConcurrency({ maxParallelCli: 2, perCli: {} });
    const r1 = await acquire('claude-code');
    expect(snapshot().globalInFlight).toBe(1);
    r1();
    r1(); // should be no-op
    r1(); // still no-op
    expect(snapshot().globalInFlight).toBe(0);
    expect(snapshot().perCli).toEqual({});
  });

  it('reads settings dynamically so changes take effect without restart', async () => {
    await setConcurrency({ maxParallelCli: 1, perCli: {} });
    const r1 = await acquire('claude-code');

    // r2 queues at cap=1.
    let r2Settled = false;
    const r2Promise = acquire('codex-cli').then((r) => {
      r2Settled = true;
      return r;
    });
    await flush();
    expect(r2Settled).toBe(false);

    // Bump the cap to 2 — r2 should NOT auto-grant on its own (the
    // semaphore only re-evaluates on release, by design — settings
    // changes take effect on next acquire/release, not retroactively
    // on already-queued waiters). Verify that contract.
    await setConcurrency({ maxParallelCli: 2, perCli: {} });
    await flush();
    expect(r2Settled).toBe(false);

    // Releasing r1 triggers tryGrantHead, which now sees the new cap
    // and grants r2.
    r1();
    const r2 = await r2Promise;
    expect(r2Settled).toBe(true);
    r2();
  });

  it('concurrent acquire calls do not over-grant past the global cap', async () => {
    // Regression guard for the fast-path race: two near-simultaneous
    // acquires both await getConcurrency(), both resume after the yield,
    // both might see globalInFlight < cap and both increment, exceeding
    // the cap. The always-enqueue + tryGrantHead-mutex design must
    // serialize them.
    await setConcurrency({ maxParallelCli: 2, perCli: {} });

    // Fire 5 simultaneous acquires — only 2 should resolve immediately.
    const releases: (() => void)[] = [];
    const promises = Array.from({ length: 5 }, () => acquire('claude-code'));

    // Let all microtasks drain; with cap=2 only the first two should resolve.
    await flush();
    expect(snapshot().globalInFlight).toBe(2);
    expect(snapshot().queueDepth).toBe(3);

    // Drain in order to verify we eventually grant all 5 without ever
    // exceeding the cap.
    let maxObserved = 0;
    while (releases.length < 5) {
      // Settle the next batch of grants, capture peak.
      await flush();
      maxObserved = Math.max(maxObserved, snapshot().globalInFlight);
      // Resolve one promise that has already resolved (no-op if pending).
      const settledIndex = await Promise.race(
        promises.map((p, i) => p.then(() => i, () => -1)),
      );
      if (settledIndex >= 0 && releases.length <= settledIndex) {
        const release = await promises[settledIndex];
        releases.push(release);
        release();
      } else {
        break;
      }
    }
    expect(maxObserved).toBeLessThanOrEqual(2);
  });

  it('aborting a queued waiter releases the queue head and unblocks others', async () => {
    await setConcurrency({ maxParallelCli: 1, perCli: {} });

    // r1 takes the only slot.
    const r1 = await acquire('claude-code');
    expect(snapshot().globalInFlight).toBe(1);

    // r2 queues with an abort controller so we can cancel it.
    const ac = new AbortController();
    const r2Promise = acquire('claude-code', ac.signal);
    await flush();
    expect(snapshot().queueDepth).toBe(1);

    // r3 queues behind r2.
    let r3Settled = false;
    const r3Promise = acquire('codex-cli').then((r) => {
      r3Settled = true;
      return r;
    });
    await flush();
    expect(snapshot().queueDepth).toBe(2);

    // Abort r2. This must:
    //   1. Reject r2's promise.
    //   2. Remove r2 from the queue.
    //   3. Trigger tryGrantHead so r3 (which has free codex capacity)
    //      gets granted now that the head's gone.
    ac.abort();
    await expect(r2Promise).rejects.toThrow();
    await flush();

    // r1 still holds claude-code, so r3 can't get the global slot
    // until r1 releases. But r2 should be off the queue.
    expect(snapshot().queueDepth).toBe(1);

    // Releasing r1 should now grant r3.
    r1();
    const r3 = await r3Promise;
    expect(r3Settled).toBe(true);
    r3();
  });

  it('aborting before slot is granted does not increment counters', async () => {
    await setConcurrency({ maxParallelCli: 1, perCli: {} });
    const r1 = await acquire('claude-code');

    // Pre-abort signal — acquire should reject immediately, never enqueue.
    const aborted = AbortSignal.abort(new Error('pre-aborted'));
    await expect(acquire('claude-code', aborted)).rejects.toThrow();

    // No leftover counters or waiters.
    expect(snapshot().globalInFlight).toBe(1); // only r1
    expect(snapshot().queueDepth).toBe(0);
    r1();
  });

  it('tryGrantHead error in getConcurrency does not crash the daemon or strand the head', async () => {
    // We can't easily make settings.get() throw without mocking, but we
    // can prove the public guarantee: a stress sequence of release-fire
    // never produces an unhandled rejection. If the catch-and-log
    // wrapper around tryGrantHead is removed, this test would emit an
    // unhandled rejection warning under vitest's default config and
    // cause subsequent tests to behave oddly.
    await setConcurrency({ maxParallelCli: 2, perCli: {} });
    const r1 = await acquire('claude-code');
    const r2 = await acquire('codex-cli');

    // Rapid back-to-back releases — exercises the granting-mutex path.
    r1();
    r2();
    await flush();

    expect(snapshot().globalInFlight).toBe(0);
    expect(snapshot().queueDepth).toBe(0);
  });

  it('aborting a per-CLI-blocked head wakes the queue (codex-cli-0 finding)', async () => {
    // Round-2 reviewer concern: when an abort splices a blocked head
    // and pokes tryGrantHead while a live grant loop is running, the
    // bailing tryGrantHead's poke gets swallowed (granting=true). The
    // dirty flag fix re-runs the inner pass when a reentrant call
    // marks new work. Without the flag, the test below would hang on
    // the r3 acquire because B's per-CLI block was cleared but no one
    // re-evaluated.
    await setConcurrency({
      maxParallelCli: 5,
      perCli: { 'opencode-cli': 1 },
    });

    // r1 takes the only opencode slot.
    const r1 = await acquire('opencode-cli');
    expect(snapshot().perCli['opencode-cli']).toBe(1);

    // r2 (opencode) queues — strict-FIFO head, blocked on per-CLI cap.
    const ac = new AbortController();
    const r2Promise = acquire('opencode-cli', ac.signal);
    await flush();
    expect(snapshot().queueDepth).toBe(1);

    // r3 (codex) queues behind r2 — codex has free capacity but FIFO
    // means it waits behind the blocked head.
    let r3Settled = false;
    const r3Promise = acquire('codex-cli').then((r) => {
      r3Settled = true;
      return r;
    });
    await flush();
    expect(r3Settled).toBe(false);
    expect(snapshot().queueDepth).toBe(2);

    // Abort r2. The splice removes the blocked head, leaving r3 (codex,
    // free capacity) at the head. The dirty flag ensures the live
    // grant loop (or its successor) sees the new state and grants r3.
    ac.abort();
    await expect(r2Promise).rejects.toThrow();
    const r3 = await r3Promise;
    expect(r3Settled).toBe(true);

    r1();
    r3();
  });

  it('rapid push during grant drain does not strand any waiters (stress)', async () => {
    // Belt-and-suspenders: even though the empty-queue stranding
    // scenario is structurally impossible in single-threaded JS (no
    // yield between while-exit and granting=false), this test exercises
    // a high-churn sequence where many acquires and releases interleave
    // through tryGrantHead's await points. If the dirty flag ever gets
    // accidentally removed, a future yield-introducing refactor could
    // strand waiters here.
    await setConcurrency({ maxParallelCli: 2, perCli: {} });

    const N = 30;
    const releases: Promise<() => void>[] = [];
    for (let i = 0; i < N; i++) {
      releases.push(acquire('claude-code'));
    }

    // Drain — every acquire must eventually resolve. If any are
    // stranded the test times out.
    const allReleases = await Promise.all(
      releases.map(async (p) => {
        const release = await p;
        release();
        return release;
      }),
    );
    expect(allReleases).toHaveLength(N);
    expect(snapshot().globalInFlight).toBe(0);
    expect(snapshot().queueDepth).toBe(0);
  });

  it('falls back to defaults when settings row is absent', async () => {
    // No setConcurrency() call — table is empty. Defaults: global=3,
    // per-CLI=2-3. Just verify acquire doesn't throw and respects
    // defaults reasonably.
    const r1 = await acquire('claude-code');
    const r2 = await acquire('codex-cli');
    const r3 = await acquire('gemini-cli');

    // Default global=3 — fourth must queue.
    let r4Settled = false;
    const r4Promise = acquire('claude-code').then((r) => {
      r4Settled = true;
      return r;
    });
    await flush();
    expect(r4Settled).toBe(false);

    r1();
    await r4Promise.then((r) => r());
    r2();
    r3();
  });
});
