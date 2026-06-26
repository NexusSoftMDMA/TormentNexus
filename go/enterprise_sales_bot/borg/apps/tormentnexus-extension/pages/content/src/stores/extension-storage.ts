import type { StateStorage } from 'zustand/middleware';

interface ExtensionStorageArea {
  get(key: string): Promise<unknown>;
  set(items: Record<string, string>): Promise<void>;
  remove(key: string): Promise<void>;
}

interface BrowserStorageArea {
  get(key: string): Promise<Record<string, unknown>>;
  set(items: Record<string, string>): Promise<void>;
  remove(key: string): Promise<void>;
}

let hasLoggedExtensionStorageFallbackWarning = false;

function logExtensionStorageFallback(error: unknown): void {
  if (hasLoggedExtensionStorageFallbackWarning) {
    return;
  }

  hasLoggedExtensionStorageFallbackWarning = true;
  const message = error instanceof Error ? error.message : String(error);
  console.warn(`[extension-storage] Falling back to localStorage because extension storage is unavailable: ${message}`);
}

function hasLocalStorage(): boolean {
  try {
    return typeof localStorage !== 'undefined';
  } catch {
    return false;
  }
}

function getLocalStorageValue(key: string): string | null {
  if (!hasLocalStorage()) {
    return null;
  }

  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}

function setLocalStorageValue(key: string, value: string): void {
  if (!hasLocalStorage()) {
    return;
  }

  try {
    localStorage.setItem(key, value);
  } catch {
    // Ignore storage quota/private mode errors and let extension storage be the source of truth.
  }
}

function removeLocalStorageValue(key: string): void {
  if (!hasLocalStorage()) {
    return;
  }

  try {
    localStorage.removeItem(key);
  } catch {
    // Ignore cleanup failures.
  }
}

export function safeLocalStorageGetItem(key: string): string | null {
  return getLocalStorageValue(key);
}

export function safeLocalStorageSetItem(key: string, value: string): void {
  setLocalStorageValue(key, value);
}

export function safeLocalStorageRemoveItem(key: string): void {
  removeLocalStorageValue(key);
}

function createChromeStorageArea(): ExtensionStorageArea | null {
  if (typeof chrome === 'undefined' || !chrome.storage?.local) {
    return null;
  }

  return {
    async get(key: string) {
      const maybePromise = chrome.storage.local.get(key);
      if (maybePromise && typeof (maybePromise as Promise<Record<string, unknown>>).then === 'function') {
        const result = await (maybePromise as Promise<Record<string, unknown>>);
        return result?.[key];
      }

      return await new Promise<unknown>((resolve, reject) => {
        chrome.storage.local.get(key, result => {
          const error = chrome.runtime?.lastError;
          if (error) {
            reject(new Error(error.message));
            return;
          }

          resolve(result?.[key]);
        });
      });
    },
    async set(items: Record<string, string>) {
      const maybePromise = chrome.storage.local.set(items);
      if (maybePromise && typeof (maybePromise as Promise<void>).then === 'function') {
        await maybePromise;
        return;
      }

      await new Promise<void>((resolve, reject) => {
        chrome.storage.local.set(items, () => {
          const error = chrome.runtime?.lastError;
          if (error) {
            reject(new Error(error.message));
            return;
          }

          resolve();
        });
      });
    },
    async remove(key: string) {
      const maybePromise = chrome.storage.local.remove(key);
      if (maybePromise && typeof (maybePromise as Promise<void>).then === 'function') {
        await maybePromise;
        return;
      }

      await new Promise<void>((resolve, reject) => {
        chrome.storage.local.remove(key, () => {
          const error = chrome.runtime?.lastError;
          if (error) {
            reject(new Error(error.message));
            return;
          }

          resolve();
        });
      });
    },
  };
}

function createBrowserStorageArea(): ExtensionStorageArea | null {
  const extensionBrowser = (globalThis as typeof globalThis & {
    browser?: { storage?: { local?: BrowserStorageArea } };
  }).browser;

  const storageArea = extensionBrowser?.storage?.local;

  if (!storageArea) {
    return null;
  }

  return {
    async get(key: string) {
      const result = await storageArea.get(key);
      return result?.[key];
    },
    async set(items: Record<string, string>) {
      await storageArea.set(items);
    },
    async remove(key: string) {
      await storageArea.remove(key);
    },
  };
}

function getExtensionStorageArea(): ExtensionStorageArea | null {
  return createBrowserStorageArea() ?? createChromeStorageArea();
}

export async function getExtensionStorageValue(name: string): Promise<string | null> {
  const extensionStorage = getExtensionStorageArea();

  if (!extensionStorage) {
    return getLocalStorageValue(name);
  }

  const legacyValue = getLocalStorageValue(name);

  try {
    const extensionValue = await extensionStorage.get(name);
    if (typeof extensionValue === 'string') {
      return extensionValue;
    }

    if (legacyValue !== null) {
      await extensionStorage.set({ [name]: legacyValue });
      return legacyValue;
    }
  } catch (error) {
    logExtensionStorageFallback(error);
    return null;
  }

  return null;
}

export async function setExtensionStorageValue(name: string, value: string): Promise<void> {
  const extensionStorage = getExtensionStorageArea();
  if (extensionStorage) {
    try {
      await extensionStorage.set({ [name]: value });
      return;
    } catch (error) {
      logExtensionStorageFallback(error);
      return;
    }
  }

  setLocalStorageValue(name, value);
}

export async function removeExtensionStorageValue(name: string): Promise<void> {
  const extensionStorage = getExtensionStorageArea();
  if (extensionStorage) {
    try {
      await extensionStorage.remove(name);
      return;
    } catch (error) {
      logExtensionStorageFallback(error);
      return;
    }
  }

  removeLocalStorageValue(name);
}

export async function getExtensionStorageJson<T>(name: string, fallback: T): Promise<T> {
  const rawValue = await getExtensionStorageValue(name);
  if (rawValue === null) {
    return fallback;
  }

  try {
    return JSON.parse(rawValue) as T;
  } catch {
    return fallback;
  }
}

export async function setExtensionStorageJson(name: string, value: unknown): Promise<void> {
  await setExtensionStorageValue(name, JSON.stringify(value));
}

/**
 * Persist state in extension-scoped storage when available.
 *
 * Content-script `localStorage` is page-origin scoped, which fragments saved data by site.
 * This adapter promotes the source of truth to extension storage and only falls back to
 * localStorage in tests or non-extension environments where extension storage does not exist.
 */
export function createExtensionStateStorage(): StateStorage {
  return {
    async getItem(name: string) {
      return await getExtensionStorageValue(name);
    },
    async setItem(name: string, value: string) {
      await setExtensionStorageValue(name, value);
    },
    async removeItem(name: string) {
      await removeExtensionStorageValue(name);
    },
  };
}
