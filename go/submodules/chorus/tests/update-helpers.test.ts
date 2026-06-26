/**
 * Coverage for the v0.8.4 self-update helpers.
 *
 * detectNpmPrefix is the load-bearing piece — it determines where
 * `chorus update` writes the new version. A wrong prefix means the
 * update lands somewhere PATH never sees, leaving the user stuck on
 * the old version (the original bug we're trying to fix).
 *
 * versionGreater is the gate that decides whether to nudge the user
 * about an update. False negatives are fine (no nudge); false
 * positives spam the start banner.
 */
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { checkPrefixUsable, versionGreater } from '@/cli/commands/update';

describe('versionGreater', () => {
  it('returns true when latest is strictly greater', () => {
    expect(versionGreater('0.8.4', '0.8.3')).toBe(true);
    expect(versionGreater('0.9.0', '0.8.99')).toBe(true);
    expect(versionGreater('1.0.0', '0.99.99')).toBe(true);
  });

  it('returns false when latest equals current', () => {
    expect(versionGreater('0.8.3', '0.8.3')).toBe(false);
    expect(versionGreater('1.0.0', '1.0.0')).toBe(false);
  });

  it('returns false when latest is less than current', () => {
    // Defensive: a stale registry response shouldn't trigger an update
    // nudge. False here = "no update available".
    expect(versionGreater('0.8.3', '0.8.4')).toBe(false);
    expect(versionGreater('0.8.0', '0.9.0')).toBe(false);
  });

  it('handles minor and patch differences correctly', () => {
    expect(versionGreater('0.8.10', '0.8.9')).toBe(true);
    expect(versionGreater('0.8.9', '0.8.10')).toBe(false);
  });

  it('treats missing segments as 0', () => {
    expect(versionGreater('1.0', '1.0.0')).toBe(false);
    expect(versionGreater('1.0.1', '1.0')).toBe(true);
  });

  it('returns false on garbage input rather than throwing', () => {
    // parseInt('foo') === NaN → coerced to 0 in the helper.
    expect(versionGreater('foo', 'bar')).toBe(false);
    expect(versionGreater('', '0.8.3')).toBe(false);
  });
});

describe('checkPrefixUsable', () => {
  let tmpPrefix: string;

  beforeEach(() => {
    tmpPrefix = fs.mkdtempSync(path.join(os.tmpdir(), 'chorus-prefix-test-'));
  });

  afterEach(() => {
    try {
      fs.rmSync(tmpPrefix, { recursive: true, force: true });
    } catch {
      /* ignore */
    }
  });

  it('returns ok for a writable prefix', () => {
    const result = checkPrefixUsable(tmpPrefix);
    expect(result.ok).toBe(true);
  });

  it('flags Windows-mounted /mnt/c paths', () => {
    const result = checkPrefixUsable('/mnt/c/Users/dev/.nvm/versions/node/v22');
    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.reason).toContain('Windows-mounted drive');
    }
  });

  it('flags Windows-mounted /mnt/e paths (non-c drives)', () => {
    const result = checkPrefixUsable(
      '/mnt/e/openclaw/.nvm/versions/node/v22.22.1',
    );
    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.reason).toContain('Windows-mounted drive');
    }
  });

  it('does NOT flag /mnt/wsl (WSL-internal mount, not a Windows drive)', () => {
    // /mnt/wsl is the WSL2 internal namespace; not subject to the
    // Windows-handle issue. Single-letter regex anchors prevent the
    // false positive.
    const result = checkPrefixUsable('/mnt/wsl/something');
    // Either ok (writable) or fails with the writability probe — but
    // never with the Windows-mount reason.
    if (!result.ok) {
      expect(result.reason).not.toContain('Windows-mounted drive');
    }
  });

  it('does NOT flag normal Linux paths', () => {
    const result = checkPrefixUsable('/usr/local');
    // /usr/local exists; ok-ness depends on test runner's perms.
    // What matters: not flagged as Windows-mounted.
    if (!result.ok) {
      expect(result.reason).not.toContain('Windows-mounted drive');
    }
  });

  it('flags a read-only prefix with EACCES', () => {
    // Skip when running as root — root bypasses chmod restrictions.
    if (typeof process.getuid === 'function' && process.getuid() === 0) {
      return;
    }
    const lockedDir = path.join(tmpPrefix, 'locked');
    fs.mkdirSync(lockedDir);
    fs.chmodSync(lockedDir, 0o500); // read+exec, no write
    const result = checkPrefixUsable(lockedDir);
    expect(result.ok).toBe(false);
    if (!result.ok) {
      expect(result.reason).toContain("isn't writable");
    }
    // Restore perms so afterEach cleanup works.
    fs.chmodSync(lockedDir, 0o700);
  });
});
