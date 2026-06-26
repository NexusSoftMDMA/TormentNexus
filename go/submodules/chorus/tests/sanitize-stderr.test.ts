import { describe, it, expect } from 'vitest';
import { sanitizeStderr } from '../src/daemon/ship';

describe('sanitizeStderr', () => {
  it('returns empty string for empty input', () => {
    expect(sanitizeStderr('')).toBe('');
  });

  it('passes plain text through trimmed', () => {
    expect(sanitizeStderr('  hello world  ')).toBe('hello world');
  });

  it('replaces unix /home/<user> paths with ~', () => {
    expect(sanitizeStderr('error in /home/alice/repo/file.ts')).toBe('error in ~/repo/file.ts');
  });

  it('replaces /Users/<name> macOS paths with ~', () => {
    expect(sanitizeStderr('open /Users/bob/.ssh/config failed')).toBe('open ~/.ssh/config failed');
  });

  it('replaces Windows C:\\Users\\<name> paths with ~', () => {
    expect(sanitizeStderr('cannot read C:\\Users\\Alice\\repo\\thing')).toContain('~\\repo\\thing');
  });

  it('drops lines mentioning id_rsa or id_ed25519', () => {
    const raw = 'normal line\nfailed reading id_rsa\nanother normal line';
    const out = sanitizeStderr(raw);
    expect(out).toContain('normal line');
    expect(out).toContain('another normal line');
    expect(out).not.toContain('id_rsa');
  });

  it('caps output at 600 chars with truncation marker', () => {
    const raw = 'x'.repeat(2000);
    const out = sanitizeStderr(raw);
    expect(out.length).toBeLessThanOrEqual(600 + '… [truncated]'.length + 5);
    expect(out).toContain('[truncated]');
  });

  it('handles 100KB adversarial input in under 100ms', () => {
    const raw = '/home/' + 'a'.repeat(100_000);
    const start = Date.now();
    sanitizeStderr(raw);
    expect(Date.now() - start).toBeLessThan(100);
  });
});
