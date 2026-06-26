/**
 * Unit tests for the sidebar-collapsed storage helpers.
 *
 * Covers the SSR/storage-unavailable paths and the boolean round-trip,
 * including the silent-swallow contract on quota / private-mode errors.
 */
import { describe, expect, it } from 'vitest';
import {
  SIDEBAR_COLLAPSED_KEY,
  readCollapsed,
  writeCollapsed,
} from '../src/lib/sidebar-collapsed-storage.js';

class MemoryStorage {
  private map = new Map<string, string>();
  getItem(key: string): string | null {
    return this.map.get(key) ?? null;
  }
  setItem(key: string, value: string): void {
    this.map.set(key, value);
  }
}

describe('readCollapsed', () => {
  it('returns false when storage is null (SSR)', () => {
    expect(readCollapsed(null)).toBe(false);
  });

  it('returns false when storage is undefined', () => {
    expect(readCollapsed(undefined)).toBe(false);
  });

  it('returns false when the key is missing', () => {
    expect(readCollapsed(new MemoryStorage())).toBe(false);
  });

  it('returns true when the stored value is exactly "1"', () => {
    const s = new MemoryStorage();
    s.setItem(SIDEBAR_COLLAPSED_KEY, '1');
    expect(readCollapsed(s)).toBe(true);
  });

  it('returns false for any other stored value (defensive)', () => {
    const s = new MemoryStorage();
    for (const v of ['true', 'yes', '0', '01', '11', '']) {
      s.setItem(SIDEBAR_COLLAPSED_KEY, v);
      expect(readCollapsed(s)).toBe(false);
    }
  });

  it('returns false when getItem throws (private mode, quota error)', () => {
    const throwing = {
      getItem: () => {
        throw new Error('private mode');
      },
    };
    expect(readCollapsed(throwing)).toBe(false);
  });
});

describe('writeCollapsed', () => {
  it('persists "1" for true and "0" for false', () => {
    const s = new MemoryStorage();
    writeCollapsed(s, true);
    expect(s.getItem(SIDEBAR_COLLAPSED_KEY)).toBe('1');
    writeCollapsed(s, false);
    expect(s.getItem(SIDEBAR_COLLAPSED_KEY)).toBe('0');
  });

  it('is a no-op when storage is null', () => {
    expect(() => writeCollapsed(null, true)).not.toThrow();
  });

  it('swallows setItem errors silently (quota/private mode)', () => {
    const throwing = {
      setItem: () => {
        throw new Error('QuotaExceeded');
      },
    };
    expect(() => writeCollapsed(throwing, true)).not.toThrow();
  });

  it('round-trips through read/write', () => {
    const s = new MemoryStorage();
    writeCollapsed(s, true);
    expect(readCollapsed(s)).toBe(true);
    writeCollapsed(s, false);
    expect(readCollapsed(s)).toBe(false);
  });
});
