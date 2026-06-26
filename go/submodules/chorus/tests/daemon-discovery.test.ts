/**
 * Coverage for the v0.8 port-discovery substrate (~/.chorus/daemon.json).
 *
 * The tests pin behaviour at the boundaries the rest of the codebase
 * depends on — read order, fall-through to env / default, port-walk
 * exhaustion, schema rejection — without booting a real daemon.
 *
 * HOME is rewritten per-test to a fresh tempdir so daemon.json never
 * pollutes the developer's actual ~/.chorus.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

import {
  DEFAULT_COCKPIT_URL,
  DEFAULT_DAEMON_URL,
  daemonInfoPath,
  isPidAlive,
  pickFreePort,
  readDaemonInfo,
  readLiveDaemonInfo,
  resolveCockpitUrl,
  resolveDaemonUrl,
  writeDaemonInfo,
  type DaemonInfo,
} from '@/lib/daemon-discovery';

let homeDir: string;
let originalHome: string | undefined;

beforeEach(() => {
  homeDir = fs.mkdtempSync(path.join(os.tmpdir(), 'chorus-discovery-'));
  originalHome = process.env.HOME;
  process.env.HOME = homeDir;
  delete process.env.CHORUS_DAEMON_URL;
  delete process.env.CHORUS_COCKPIT_URL;
});

afterEach(() => {
  if (originalHome === undefined) delete process.env.HOME;
  else process.env.HOME = originalHome;
  fs.rmSync(homeDir, { recursive: true, force: true });
});

const sampleInfo = (overrides: Partial<DaemonInfo> = {}): DaemonInfo => ({
  schemaVersion: 1,
  daemonPort: 7707,
  cockpitPort: 5050,
  daemonPid: process.pid,
  cockpitPid: process.pid,
  startedAt: new Date().toISOString(),
  version: '0.8.0',
  ...overrides,
});

describe('daemonInfoPath', () => {
  it('points at $HOME/.chorus/daemon.json', () => {
    expect(daemonInfoPath()).toBe(path.join(homeDir, '.chorus', 'daemon.json'));
  });
});

describe('writeDaemonInfo + readDaemonInfo round-trip', () => {
  it('writes and reads back a valid record', () => {
    const info = sampleInfo({ daemonPort: 7708, cockpitPort: 5051 });
    writeDaemonInfo(info);

    const read = readDaemonInfo();
    expect(read).not.toBeNull();
    expect(read!.daemonPort).toBe(7708);
    expect(read!.cockpitPort).toBe(5051);
    expect(read!.schemaVersion).toBe(1);
  });

  it('returns null when the file does not exist', () => {
    expect(readDaemonInfo()).toBeNull();
  });

  it('returns null on corrupted JSON', () => {
    const target = daemonInfoPath();
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, '{ not json');
    expect(readDaemonInfo()).toBeNull();
  });

  it('returns null on schemaVersion mismatch', () => {
    const target = daemonInfoPath();
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(
      target,
      JSON.stringify({ ...sampleInfo(), schemaVersion: 2 }),
    );
    expect(readDaemonInfo()).toBeNull();
  });

  it('returns null when required ports are missing', () => {
    const target = daemonInfoPath();
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(
      target,
      JSON.stringify({ schemaVersion: 1, daemonPid: 1 }),
    );
    expect(readDaemonInfo()).toBeNull();
  });
});

describe('isPidAlive', () => {
  it('returns true for the current process', () => {
    expect(isPidAlive(process.pid)).toBe(true);
  });

  it('returns false for invalid PIDs', () => {
    expect(isPidAlive(0)).toBe(false);
    expect(isPidAlive(-1)).toBe(false);
    expect(isPidAlive(NaN)).toBe(false);
  });

  it('returns false for a clearly-dead PID', () => {
    // PID 999999 is unlikely to exist on any test runner.
    expect(isPidAlive(999_999)).toBe(false);
  });
});

describe('readLiveDaemonInfo', () => {
  it('returns null when no daemon.json exists', async () => {
    expect(await readLiveDaemonInfo()).toBeNull();
  });

  it('returns null when PID is dead even if file is valid', async () => {
    writeDaemonInfo(sampleInfo({ daemonPid: 999_999 }));
    expect(await readLiveDaemonInfo()).toBeNull();
  });

  it('returns null when health probe fails (no listener)', async () => {
    // Use an unlikely high port so the health probe gets ECONNREFUSED.
    writeDaemonInfo(sampleInfo({ daemonPort: 64321 }));
    expect(await readLiveDaemonInfo({ healthTimeoutMs: 200 })).toBeNull();
  });
});

describe('resolveDaemonUrl', () => {
  it('returns the default when no daemon.json and no env var', async () => {
    expect(await resolveDaemonUrl()).toBe(DEFAULT_DAEMON_URL);
  });

  it('uses CHORUS_DAEMON_URL when no live daemon', async () => {
    process.env.CHORUS_DAEMON_URL = 'http://example.invalid:9999';
    expect(await resolveDaemonUrl()).toBe('http://example.invalid:9999');
  });

  it('falls back to default if env var is empty string', async () => {
    process.env.CHORUS_DAEMON_URL = '';
    expect(await resolveDaemonUrl()).toBe(DEFAULT_DAEMON_URL);
  });

  it('still falls back to env var when daemon.json points at a dead PID', async () => {
    writeDaemonInfo(sampleInfo({ daemonPid: 999_999 }));
    process.env.CHORUS_DAEMON_URL = 'http://example.invalid:9999';
    expect(await resolveDaemonUrl()).toBe('http://example.invalid:9999');
  });
});

describe('resolveCockpitUrl', () => {
  it('returns the default when no daemon.json and no env var', async () => {
    expect(await resolveCockpitUrl()).toBe(DEFAULT_COCKPIT_URL);
  });

  it('uses CHORUS_COCKPIT_URL override', async () => {
    process.env.CHORUS_COCKPIT_URL = 'http://elsewhere:6060';
    expect(await resolveCockpitUrl()).toBe('http://elsewhere:6060');
  });
});

describe('pickFreePort', () => {
  it('returns the preferred port when free', async () => {
    const isInUse = vi.fn().mockResolvedValue(false);
    expect(await pickFreePort(7707, 14, isInUse)).toBe(7707);
    expect(isInUse).toHaveBeenCalledWith(7707);
  });

  it('walks past taken ports', async () => {
    const taken = new Set([7707, 7708]);
    const isInUse = (p: number): Promise<boolean> => Promise.resolve(taken.has(p));
    expect(await pickFreePort(7707, 14, isInUse)).toBe(7709);
  });

  it('returns null when the entire range is taken', async () => {
    const isInUse = (): Promise<boolean> => Promise.resolve(true);
    expect(await pickFreePort(7707, 14, isInUse)).toBeNull();
  });

  it('respects the range boundary', async () => {
    // Range of 3: 7707, 7708, 7709 — last in range is 7709.
    const taken = new Set([7707, 7708, 7709]);
    const isInUse = (p: number): Promise<boolean> => Promise.resolve(taken.has(p));
    expect(await pickFreePort(7707, 3, isInUse)).toBeNull();
  });
});
