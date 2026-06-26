/**
 * Concurrency settings persistence + parse semantics.
 *
 * Mostly guards: defaults when row absent, safeParse fallback when row
 * is corrupt, partial perCli merge keeps unaffected lineages alone.
 */

import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import os from 'os';
import path from 'path';
import fs from 'fs';
import { randomUUID } from 'crypto';

import { _resetDbForTests, getDb, settings } from '@/lib/db';
import {
  getConcurrency,
  setConcurrency,
  resolvePerCliCap,
  _defaults,
} from '@/lib/settings/concurrency';

let dbPath: string;

beforeEach(async () => {
  dbPath = path.join(os.tmpdir(), `chorus-concurrency-${randomUUID()}.db`);
  process.env.CHORUS_DB_PATH = dbPath;
  await _resetDbForTests();
  await getDb();
});

afterEach(async () => {
  await _resetDbForTests();
  for (const suffix of ['', '-shm', '-wal']) {
    try { fs.unlinkSync(dbPath + suffix); } catch { /* best-effort */ }
  }
  delete process.env.CHORUS_DB_PATH;
});

describe('concurrency settings', () => {
  it('returns defaults when no row exists', async () => {
    const config = await getConcurrency();
    expect(config.maxParallelCli).toBe(_defaults.maxParallelCli);
    expect(config.perCli).toEqual({});
  });

  it('round-trips a full config', async () => {
    await setConcurrency({
      maxParallelCli: 5,
      perCli: { 'opencode-cli': 1, 'gemini-cli': 3 },
    });
    const fetched = await getConcurrency();
    expect(fetched.maxParallelCli).toBe(5);
    expect(fetched.perCli['opencode-cli']).toBe(1);
    expect(fetched.perCli['gemini-cli']).toBe(3);
  });

  it('safeParse fallback: corrupt JSON in DB → defaults, no throw', async () => {
    // Plant a structurally-bad value directly via the underlying
    // settings store. Without safeParse, getConcurrency would throw on
    // schema validation; with it, defaults take over silently.
    await settings.set('concurrency', { maxParallelCli: 'not-a-number' });
    const config = await getConcurrency();
    expect(config.maxParallelCli).toBe(_defaults.maxParallelCli);
  });

  it('rejects out-of-range values on write', async () => {
    await expect(
      setConcurrency({ maxParallelCli: 99, perCli: {} }),
    ).rejects.toThrow();
    await expect(
      setConcurrency({ maxParallelCli: 0, perCli: {} }),
    ).rejects.toThrow();
    await expect(
      setConcurrency({
        maxParallelCli: 3,
        perCli: { 'opencode-cli': 10 }, // perCli max is 5
      }),
    ).rejects.toThrow();
  });

  it('resolvePerCliCap falls back to defaults for unset lineages', async () => {
    const config = await getConcurrency();
    // Default for opencode-cli is 2; verify the helper returns it
    // when the row is empty.
    expect(resolvePerCliCap(config, 'opencode-cli')).toBe(2);
    expect(resolvePerCliCap(config, 'claude-code')).toBe(3);

    await setConcurrency({
      maxParallelCli: 5,
      perCli: { 'opencode-cli': 1 },
    });
    const updated = await getConcurrency();
    expect(resolvePerCliCap(updated, 'opencode-cli')).toBe(1); // override
    expect(resolvePerCliCap(updated, 'gemini-cli')).toBe(2); // still default
  });
});
