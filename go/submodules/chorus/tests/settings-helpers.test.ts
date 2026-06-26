/**
 * Round-trip coverage for the typed settings helpers — the highest-risk
 * callsite class for the libsql migration (per cdx-1 review).
 *
 * Why this exists: each helper does
 *   const raw = await settings.get(KEY);
 *   const parsed = Schema.safeParse(raw);
 *   return parsed.success ? parsed.data : DEFAULT;
 *
 * If a caller forgets to `await` the helper, the typed return is a
 * Promise that gets coerced to defaults downstream — silently dropping
 * the user's stored value. These round-trip tests detect that bug: persist
 * a NON-default value, then read it back via the helper and assert. Any
 * silent fallback to the default will fail the test.
 */

import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import os from 'os';
import path from 'path';
import fs from 'fs';
import { randomUUID } from 'crypto';

import { _resetDbForTests } from '@/lib/db';
import { getTransport, setTransport, DEFAULT_TRANSPORT } from '@/lib/settings/transport';
import {
  getPermissions,
  setPermissions,
  DEFAULT_PERMISSIONS,
} from '@/lib/settings/permissions';
import { getBillingMode, setBillingMode, DEFAULT_BILLING_MODE } from '@/lib/settings/billing';

let dbPath: string;

beforeEach(async () => {
  dbPath = path.join(os.tmpdir(), `chorus-settings-${randomUUID()}.db`);
  process.env.CHORUS_DB_PATH = dbPath;
  delete process.env.CHORUS_TRANSPORT; // env override would mask the bug
  await _resetDbForTests();
});

afterEach(async () => {
  await _resetDbForTests();
  for (const suffix of ['', '-shm', '-wal']) {
    try { fs.unlinkSync(dbPath + suffix); } catch { /* best-effort */ }
  }
  delete process.env.CHORUS_DB_PATH;
});

describe('transport', () => {
  it('defaults to headless when nothing is stored', async () => {
    expect(await getTransport()).toBe(DEFAULT_TRANSPORT);
  });

  it('round-trips a non-default value (tmux)', async () => {
    await setTransport('tmux');
    expect(await getTransport()).toBe('tmux');
  });

  it('CHORUS_TRANSPORT env override wins over stored value', async () => {
    await setTransport('tmux');
    process.env.CHORUS_TRANSPORT = 'headless';
    expect(await getTransport()).toBe('headless');
  });
});

describe('permissions', () => {
  it('defaults match DEFAULT_PERMISSIONS when nothing is stored', async () => {
    expect(await getPermissions()).toEqual(DEFAULT_PERMISSIONS);
  });

  it('round-trips sandbox profile', async () => {
    await setPermissions({ sandboxProfile: 'strict' });
    expect((await getPermissions()).sandboxProfile).toBe('strict');
  });

  it('round-trips autoApprovePrompts (the boolean trap field)', async () => {
    // Default is true, so persist false to ensure a real round-trip is observed.
    await setPermissions({ autoApprovePrompts: false });
    expect((await getPermissions()).autoApprovePrompts).toBe(false);
  });

  it('round-trips networkAccess (boolean)', async () => {
    await setPermissions({ networkAccess: true });
    expect((await getPermissions()).networkAccess).toBe(true);
  });

  it('partial set merges with stored values', async () => {
    await setPermissions({ sandboxProfile: 'full', autoApprovePrompts: false });
    await setPermissions({ networkAccess: true }); // partial — should not reset others
    const got = await getPermissions();
    expect(got.sandboxProfile).toBe('full');
    expect(got.autoApprovePrompts).toBe(false);
    expect(got.networkAccess).toBe(true);
  });
});

describe('billing mode', () => {
  it('defaults to DEFAULT_BILLING_MODE when nothing is stored', async () => {
    expect(await getBillingMode()).toBe(DEFAULT_BILLING_MODE);
  });

  it('round-trips subscription', async () => {
    await setBillingMode('subscription');
    expect(await getBillingMode()).toBe('subscription');
  });

  it('round-trips mixed', async () => {
    await setBillingMode('mixed');
    expect(await getBillingMode()).toBe('mixed');
  });
});
