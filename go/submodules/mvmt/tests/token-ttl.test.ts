import { describe, expect, it } from 'vitest';
import { formatTokenExpiry, isExpired, parseTokenTtl } from '../src/cli/token-ttl.js';

describe('parseTokenTtl', () => {
  it('parses minute, hour, and day durations into expiration timestamps', () => {
    const now = Date.parse('2026-04-29T12:00:00.000Z');

    expect(parseTokenTtl('30m', now)).toEqual({
      label: '30m',
      expiresAt: '2026-04-29T12:30:00.000Z',
    });
    expect(parseTokenTtl('2h', now)).toEqual({
      label: '2h',
      expiresAt: '2026-04-29T14:00:00.000Z',
    });
    expect(parseTokenTtl('7d', now)).toEqual({
      label: '7d',
      expiresAt: '2026-05-06T12:00:00.000Z',
    });
  });

  it('uses never as a non-expiring token', () => {
    expect(parseTokenTtl('never', 1_000_000)).toEqual({
      label: 'never',
      expiresAt: undefined,
    });
  });

  it('defaults to 30 days when no TTL is provided', () => {
    const now = Date.parse('2026-04-29T12:00:00.000Z');

    expect(parseTokenTtl(undefined, now)).toEqual({
      label: '30d',
      expiresAt: '2026-05-29T12:00:00.000Z',
    });
  });

  it('rejects invalid or unsafe TTL values', () => {
    expect(() => parseTokenTtl('', 1_000_000)).toThrow('TTL must look like 30m, 7d, or never');
    expect(() => parseTokenTtl('0m', 1_000_000)).toThrow('TTL must be greater than zero');
    expect(() => parseTokenTtl('-1d', 1_000_000)).toThrow('TTL must look like 30m, 7d, or never');
    expect(() => parseTokenTtl('7w', 1_000_000)).toThrow('TTL must look like 30m, 7d, or never');
  });

  it('formats and evaluates optional expiration timestamps', () => {
    const now = Date.parse('2026-04-29T12:00:00.000Z');

    expect(formatTokenExpiry(undefined)).toBe('never');
    expect(formatTokenExpiry('2026-05-06T12:00:00.000Z')).toBe('2026-05-06T12:00:00.000Z');
    expect(isExpired(undefined, now)).toBe(false);
    expect(isExpired('2026-04-29T11:59:59.000Z', now)).toBe(true);
    expect(isExpired('2026-04-29T12:00:01.000Z', now)).toBe(false);
    expect(isExpired('not-a-date', now)).toBe(true);
  });
});
