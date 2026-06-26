import { describe, it, expect, vi } from 'vitest';
import { runWithChainFallback, type ChainEntry } from '../src/daemon/runner/run-with-fallback';

describe('runWithChainFallback', () => {
  it('returns the result on first attempt without invoking onFallback', async () => {
    const onFallback = vi.fn();
    const chain: ChainEntry[] = [
      { lineage: 'openai', model: 'gpt-5.5' },
      { lineage: 'openai', model: 'gpt-5.4' },
    ];
    const result = await runWithChainFallback(
      chain,
      async (entry) => ({ ok: true, used: entry }),
      onFallback,
    );
    expect(result).toEqual({ ok: true, used: { lineage: 'openai', model: 'gpt-5.5' } });
    expect(onFallback).not.toHaveBeenCalled();
  });

  it('falls through same-lineage entries on null', async () => {
    const onFallback = vi.fn();
    const calls: ChainEntry[] = [];
    const chain: ChainEntry[] = [
      { lineage: 'openai', model: 'gpt-5.5' },
      { lineage: 'openai', model: 'gpt-5.4' },
    ];
    const result = await runWithChainFallback(
      chain,
      async (entry) => {
        calls.push(entry);
        return entry.model === 'gpt-5.4' ? { ok: true } : null;
      },
      onFallback,
    );
    expect(result).toEqual({ ok: true });
    expect(calls).toEqual(chain);
    expect(onFallback).toHaveBeenCalledTimes(1);
    expect(onFallback).toHaveBeenCalledWith(chain[0], chain[1], 0);
  });

  it('falls through cross-lineage when same-lineage chain exhausts', async () => {
    // The headline v0.8 case: codex slot with two same-lineage gpt-5.x
    // models, then a claude cross-lineage fallback. First two return null
    // (e.g. quota exhausted), claude succeeds.
    const onFallback = vi.fn();
    const calls: ChainEntry[] = [];
    const chain: ChainEntry[] = [
      { lineage: 'openai', model: 'gpt-5.5' },
      { lineage: 'openai', model: 'gpt-5.4' },
      { lineage: 'anthropic', model: 'claude-opus-4-7' },
    ];
    const result = await runWithChainFallback(
      chain,
      async (entry) => {
        calls.push(entry);
        return entry.lineage === 'anthropic' ? { ok: true, by: entry } : null;
      },
      onFallback,
    );
    expect(result).toEqual({
      ok: true,
      by: { lineage: 'anthropic', model: 'claude-opus-4-7' },
    });
    expect(calls).toHaveLength(3);
    // Two transitions: one same-lineage, one cross-lineage.
    expect(onFallback).toHaveBeenCalledTimes(2);
    expect(onFallback).toHaveBeenNthCalledWith(1, chain[0], chain[1], 0);
    expect(onFallback).toHaveBeenNthCalledWith(2, chain[1], chain[2], 1);
  });

  it('returns null when every entry returns null', async () => {
    const onFallback = vi.fn();
    const chain: ChainEntry[] = [
      { lineage: 'openai', model: 'gpt-5.5' },
      { lineage: 'anthropic', model: 'claude-opus-4-7' },
    ];
    const result = await runWithChainFallback(chain, async () => null, onFallback);
    expect(result).toBeNull();
    expect(onFallback).toHaveBeenCalledTimes(1);
  });

  it('returns null on empty chain without invoking attempt', async () => {
    const attempt = vi.fn();
    const onFallback = vi.fn();
    const result = await runWithChainFallback([], attempt, onFallback);
    expect(result).toBeNull();
    expect(attempt).not.toHaveBeenCalled();
    expect(onFallback).not.toHaveBeenCalled();
  });

  it('propagates thrown errors instead of swallowing them', async () => {
    const onFallback = vi.fn();
    const chain: ChainEntry[] = [
      { lineage: 'openai', model: 'gpt-5.5' },
      { lineage: 'anthropic', model: 'claude-opus-4-7' },
    ];
    await expect(
      runWithChainFallback(chain, async () => { throw new Error('boom'); }, onFallback),
    ).rejects.toThrow('boom');
    expect(onFallback).not.toHaveBeenCalled();
  });

  it('handles undefined model (lineage default) entries', async () => {
    const onFallback = vi.fn();
    const seen: ChainEntry[] = [];
    const chain: ChainEntry[] = [
      { lineage: 'anthropic', model: undefined },
      { lineage: 'openai', model: undefined },
    ];
    await runWithChainFallback(
      chain,
      async (entry) => {
        seen.push(entry);
        return null;
      },
      onFallback,
    );
    expect(seen).toEqual(chain);
  });
});
