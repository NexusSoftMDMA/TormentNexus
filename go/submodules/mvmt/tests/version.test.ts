import { describe, expect, it, vi } from 'vitest';
import {
  checkForUpdate,
  formatUpdateNotice,
  isVersionGreater,
  maybePrintUpdateNotice,
  shouldSkipUpdateCheck,
} from '../src/utils/version.js';

describe('version utilities', () => {
  it('compares semver versions without a dependency', () => {
    expect(isVersionGreater('1.2.4', '1.2.3')).toBe(true);
    expect(isVersionGreater('1.3.0', '1.2.9')).toBe(true);
    expect(isVersionGreater('2.0.0', '1.99.99')).toBe(true);
    expect(isVersionGreater('1.2.3', '1.2.3')).toBe(false);
    expect(isVersionGreater('1.2.2', '1.2.3')).toBe(false);
    expect(isVersionGreater('1.0.0', '1.0.0-beta.1')).toBe(true);
  });

  it('skips update checks when disabled by env', () => {
    expect(shouldSkipUpdateCheck({ MVMT_NO_UPDATE_CHECK: '1' })).toBe(true);
    expect(shouldSkipUpdateCheck({ MVMT_NO_UPDATE_CHECK: 'true' })).toBe(true);
    expect(shouldSkipUpdateCheck({ CI: 'true' })).toBe(true);
    expect(shouldSkipUpdateCheck({})).toBe(false);
  });

  it('reports available updates', async () => {
    const fetchImpl = vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({ version: '1.1.0' }),
    })) as unknown as typeof fetch;

    const result = await checkForUpdate({
      packageInfo: { name: 'mvmt', version: '1.0.0', packagePath: '/tmp/package.json' },
      fetchImpl,
      env: {},
    });

    expect(result).toMatchObject({
      status: 'available',
      packageName: 'mvmt',
      currentVersion: '1.0.0',
      latestVersion: '1.1.0',
    });
  });

  it('formats and prints update notices only when a newer version exists', async () => {
    expect(
      formatUpdateNotice({
        status: 'available',
        packageName: 'mvmt',
        currentVersion: '1.0.0',
        latestVersion: '1.1.0',
      }),
    ).toContain('Update available');

    expect(
      formatUpdateNotice({
        status: 'current',
        packageName: 'mvmt',
        currentVersion: '1.0.0',
        latestVersion: '1.0.0',
      }),
    ).toBeUndefined();

    let output = '';
    const fetchImpl = vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({ version: '1.1.0' }),
    })) as unknown as typeof fetch;

    await maybePrintUpdateNotice({
      packageInfo: { name: 'mvmt', version: '1.0.0', packagePath: '/tmp/package.json' },
      fetchImpl,
      env: {},
      stream: { write: (chunk: string) => { output += chunk; return true; } } as NodeJS.WriteStream,
    });

    expect(output).toContain('mvmt 1.0.0 -> 1.1.0');
  });
});
