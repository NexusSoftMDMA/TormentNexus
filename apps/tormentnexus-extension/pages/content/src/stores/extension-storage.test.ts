import { beforeEach, describe, expect, it, vi } from 'vitest';

import {
  getExtensionStorageValue,
  removeExtensionStorageValue,
  setExtensionStorageValue,
} from './extension-storage';

describe('extension-storage fallback', () => {
  beforeEach(() => {
    const storage = new Map<string, string>();
    Object.defineProperty(globalThis, 'localStorage', {
      value: {
        getItem: (key: string) => storage.get(key) ?? null,
        setItem: (key: string, value: string) => {
          storage.set(key, value);
        },
        removeItem: (key: string) => {
          storage.delete(key);
        },
        clear: () => {
          storage.clear();
        },
      },
      configurable: true,
      writable: true,
    });

    localStorage.clear();
    vi.restoreAllMocks();
    vi.spyOn(console, 'warn').mockImplementation(() => {});

    (globalThis as typeof globalThis & { chrome?: unknown }).chrome = {
      storage: {
        local: {
          get: vi.fn(async () => {
            throw new Error('Access to storage is not allowed from this context.');
          }),
          set: vi.fn(async () => {
            throw new Error('Access to storage is not allowed from this context.');
          }),
          remove: vi.fn(async () => {
            throw new Error('Access to storage is not allowed from this context.');
          }),
        },
      },
    };
  });

  it('does not fall back to page localStorage when extension storage access is denied', async () => {
    localStorage.setItem('tormentnexus-key', 'legacy-value');

    await expect(getExtensionStorageValue('tormentnexus-key')).resolves.toBeNull();

    await setExtensionStorageValue('tormentnexus-key', 'next-value');
    expect(localStorage.getItem('tormentnexus-key')).toBe('legacy-value');

    await removeExtensionStorageValue('tormentnexus-key');
    expect(localStorage.getItem('tormentnexus-key')).toBe('legacy-value');
    localStorage.setItem('tormentnexus-key', 'legacy-value');

    await expect(getExtensionStorageValue('tormentnexus-key')).resolves.toBeNull();

    await setExtensionStorageValue('tormentnexus-key', 'next-value');
    expect(localStorage.getItem('tormentnexus-key')).toBe('legacy-value');

    await removeExtensionStorageValue('tormentnexus-key');
    expect(localStorage.getItem('tormentnexus-key')).toBe('legacy-value');

    expect(console.warn).toHaveBeenCalledTimes(1);
  });

  it('falls back to localStorage when extension storage is unavailable entirely', async () => {
    delete (globalThis as typeof globalThis & { chrome?: unknown }).chrome;
    localStorage.setItem('tormentnexus-key', 'legacy-value');

    await expect(getExtensionStorageValue('tormentnexus-key')).resolves.toBe('legacy-value');

    await setExtensionStorageValue('tormentnexus-key', 'next-value');
    expect(localStorage.getItem('tormentnexus-key')).toBe('next-value');

    await removeExtensionStorageValue('tormentnexus-key');
    expect(localStorage.getItem('tormentnexus-key')).toBeNull();
    localStorage.setItem('tormentnexus-key', 'legacy-value');

    await expect(getExtensionStorageValue('tormentnexus-key')).resolves.toBe('legacy-value');

    await setExtensionStorageValue('tormentnexus-key', 'next-value');
    expect(localStorage.getItem('tormentnexus-key')).toBe('next-value');

    await removeExtensionStorageValue('tormentnexus-key');
    expect(localStorage.getItem('tormentnexus-key')).toBeNull();

    expect(console.warn).not.toHaveBeenCalled();
  });
});
