import { describe, it, expect, vi } from 'vitest';
import { runWithModelFallback } from '../src/daemon/runner/run-with-fallback';

describe('runWithModelFallback', () => {
  it('returns the result on first attempt without invoking onFallback', async () => {
    const onFallback = vi.fn();
    const result = await runWithModelFallback(
      ['m1', 'm2'],
      async (model) => ({ ok: true, model }),
      onFallback,
    );
    expect(result).toEqual({ ok: true, model: 'm1' });
    expect(onFallback).not.toHaveBeenCalled();
  });

  it('falls through when first attempt returns null and second succeeds', async () => {
    const onFallback = vi.fn();
    const calls: (string | undefined)[] = [];
    const result = await runWithModelFallback(
      ['m1', 'm2', 'm3'],
      async (model) => {
        calls.push(model);
        return model === 'm2' ? { ok: true, model } : null;
      },
      onFallback,
    );
    expect(result).toEqual({ ok: true, model: 'm2' });
    expect(calls).toEqual(['m1', 'm2']);
    expect(onFallback).toHaveBeenCalledTimes(1);
    expect(onFallback).toHaveBeenCalledWith('m1', 'm2', 0);
  });

  it('returns null when every model returns null and emits onFallback between each pair', async () => {
    const onFallback = vi.fn();
    const result = await runWithModelFallback(
      ['m1', 'm2', 'm3'],
      async () => null,
      onFallback,
    );
    expect(result).toBeNull();
    // Three models = two transitions (m1→m2, m2→m3). No transition after the last.
    expect(onFallback).toHaveBeenCalledTimes(2);
    expect(onFallback).toHaveBeenNthCalledWith(1, 'm1', 'm2', 0);
    expect(onFallback).toHaveBeenNthCalledWith(2, 'm2', 'm3', 1);
  });

  it('treats empty/missing models as a single undefined attempt (lineage default)', async () => {
    const onFallback = vi.fn();
    const calls: (string | undefined)[] = [];
    const resultUndefined = await runWithModelFallback(
      undefined,
      async (model) => {
        calls.push(model);
        return { ok: true, model };
      },
      onFallback,
    );
    expect(resultUndefined).toEqual({ ok: true, model: undefined });
    expect(calls).toEqual([undefined]);
    expect(onFallback).not.toHaveBeenCalled();

    calls.length = 0;
    const resultEmpty = await runWithModelFallback(
      [],
      async (model) => {
        calls.push(model);
        return null;
      },
      onFallback,
    );
    expect(resultEmpty).toBeNull();
    expect(calls).toEqual([undefined]);
    // Single attempt → no transitions.
    expect(onFallback).not.toHaveBeenCalled();
  });

  it('propagates thrown errors instead of swallowing them', async () => {
    const onFallback = vi.fn();
    await expect(
      runWithModelFallback(
        ['m1', 'm2'],
        async () => {
          throw new Error('boom');
        },
        onFallback,
      ),
    ).rejects.toThrow('boom');
    // Throw is fail-fast — no fallback is engaged.
    expect(onFallback).not.toHaveBeenCalled();
  });

  it('serialises attempts (later attempts only run after earlier ones resolve)', async () => {
    const onFallback = vi.fn();
    const order: string[] = [];
    await runWithModelFallback(
      ['a', 'b', 'c'],
      async (model) => {
        order.push(`start:${model}`);
        await new Promise((r) => setTimeout(r, 5));
        order.push(`end:${model}`);
        return model === 'c' ? { ok: true, model } : null;
      },
      onFallback,
    );
    expect(order).toEqual([
      'start:a',
      'end:a',
      'start:b',
      'end:b',
      'start:c',
      'end:c',
    ]);
  });
});
