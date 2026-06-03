import { EventEmitter } from 'events';

/**
 * CacheEntry stores a value along with timing metadata for TTL and LRU eviction.
 * - `createdAt`: when the entry was first set (used for TTL checks)
 * - `accessedAt`: last time the entry was read (used for LRU ordering)
 * - `ttl`: time-to-live in milliseconds; entries expire after `createdAt + ttl`
 */
interface CacheEntry<T> {
    value: T;
    createdAt: number;
    accessedAt: number;
    ttl: number;
}

/**
 * Options for constructing a CacheService instance.
 */
interface CacheServiceOptions {
    /** Maximum number of entries before LRU eviction kicks in */
    maxSize?: number;
    /** Default TTL in milliseconds for entries that don't specify one */
    defaultTTL?: number;
    /** Interval in milliseconds for automatic expired-entry cleanup. Set to 0 to disable. */
    cleanupInterval?: number;
}

/**
 * Generic in-memory cache with LRU eviction and TTL expiration.
 *
 * Supports:
 * - TTL-based auto-expiration (entries are lazily pruned on access AND periodically via cleanup timer)
 * - LRU eviction when the cache reaches `maxSize` (least recently accessed entry is evicted)
 * - Event emission: `set`, `hit`, `miss`, `delete`
 * - Singleton pattern via `CacheService.getInstance(namespace)` for shared caches
 * - `cached()` helper for wrapping async functions with transparent caching
 *
 * Used by the TormentNexus control plane for metadata caching, tool inventory caching,
 * and provider response caching to reduce redundant IO and computation.
 */
export class CacheService<T> extends EventEmitter {
    private readonly entries: Map<string, CacheEntry<T>> = new Map();
    private readonly maxSize: number;
    private readonly defaultTTL: number;
    private cleanupTimer: ReturnType<typeof setInterval> | undefined;

    /** Singleton registry keyed by namespace */
    private static readonly instances: Map<string, CacheService<unknown>> = new Map();

    constructor(options: CacheServiceOptions = {}) {
        super();
        this.maxSize = options.maxSize ?? 1000;
        this.defaultTTL = options.defaultTTL ?? 60_000;

        if (options.cleanupInterval && options.cleanupInterval > 0) {
            this.cleanupTimer = setInterval(() => this.cleanup(), options.cleanupInterval);
            // Allow the process to exit without waiting for the timer
            if (typeof this.cleanupTimer === 'object' && 'unref' in this.cleanupTimer) {
                this.cleanupTimer.unref();
            }
        }
    }

    /**
     * Get a singleton CacheService instance for the given namespace.
     * Returns the same instance for repeated calls with the same namespace.
     */
    static getInstance<V = unknown>(namespace: string): CacheService<V> {
        let instance = CacheService.instances.get(namespace);
        if (!instance) {
            instance = new CacheService<unknown>();
            CacheService.instances.set(namespace, instance);
        }
        return instance as CacheService<V>;
    }

    /**
     * Store a value in the cache.
     * @param key    Unique key for the entry
     * @param value  Value to cache
     * @param ttl    Optional TTL in ms (overrides defaultTTL)
     */
    set(key: string, value: T, ttl?: number): void {
        const effectiveTTL = ttl ?? this.defaultTTL;

        // Evict LRU if at capacity (and the key isn't already present)
        if (!this.entries.has(key) && this.entries.size >= this.maxSize) {
            this.evictLRU();
        }

        const now = Date.now();
        this.entries.set(key, {
            value,
            createdAt: now,
            accessedAt: now,
            ttl: effectiveTTL,
        });

        this.emit('set', { key, ttl: effectiveTTL });
    }

    /**
     * Retrieve a cached value. Returns `undefined` if the key is missing or expired.
     * On a hit, the entry's `accessedAt` timestamp is updated for LRU ordering.
     */
    get(key: string): T | undefined {
        const entry = this.entries.get(key);

        if (!entry) {
            this.emit('miss', { key });
            return undefined;
        }

        if (this.isExpired(entry)) {
            this.entries.delete(key);
            this.emit('miss', { key });
            return undefined;
        }

        // Update LRU timestamp
        entry.accessedAt = Date.now();
        this.emit('hit', { key });
        return entry.value;
    }

    /**
     * Check if a key exists and is not expired.
     */
    has(key: string): boolean {
        const entry = this.entries.get(key);
        if (!entry) return false;
        if (this.isExpired(entry)) {
            this.entries.delete(key);
            return false;
        }
        return true;
    }

    /**
     * Remove a specific entry from the cache.
     * @returns `true` if the key existed and was removed, `false` otherwise
     */
    delete(key: string): boolean {
        const existed = this.entries.delete(key);
        if (existed) {
            this.emit('delete', { key });
        }
        return existed;
    }

    /**
     * Remove all entries from the cache.
     */
    clear(): void {
        this.entries.clear();
    }

    /**
     * Return current cache statistics.
     */
    getStats(): { size: number; maxSize: number } {
        return {
            size: this.entries.size,
            maxSize: this.maxSize,
        };
    }

    /**
     * Stop the cleanup timer and release resources.
     * Should be called when the cache is no longer needed (e.g., in test teardown).
     */
    dispose(): void {
        if (this.cleanupTimer) {
            clearInterval(this.cleanupTimer);
            this.cleanupTimer = undefined;
        }
    }

    /**
     * Remove all expired entries. Called automatically by the cleanup timer
     * and can also be invoked manually.
     */
    private cleanup(): void {
        for (const [key, entry] of this.entries) {
            if (this.isExpired(entry)) {
                this.entries.delete(key);
            }
        }
    }

    /**
     * Check if an entry has exceeded its TTL.
     */
    private isExpired(entry: CacheEntry<T>): boolean {
        return Date.now() - entry.createdAt > entry.ttl;
    }

    /**
     * Evict the least recently accessed entry to make room for a new one.
     * Finds the entry with the oldest `accessedAt` timestamp and removes it.
     */
    private evictLRU(): void {
        let oldestKey: string | undefined;
        let oldestAccess = Infinity;

        for (const [key, entry] of this.entries) {
            if (entry.accessedAt < oldestAccess) {
                oldestAccess = entry.accessedAt;
                oldestKey = key;
            }
        }

        if (oldestKey !== undefined) {
            this.entries.delete(oldestKey);
        }
    }
}

/**
 * Transparent caching helper for async functions.
 *
 * If the key exists in the cache and is not expired, returns the cached value.
 * Otherwise, calls `fn()`, caches the result, and returns it.
 *
 * @example
 * ```ts
 * const result = await cached(cache, 'expensive-key', async () => {
 *     return await computeExpensiveValue();
 * }, 5000);
 * ```
 */
export async function cached<T>(
    cache: CacheService<T>,
    key: string,
    fn: () => Promise<T>,
    ttl?: number,
): Promise<T> {
    const existing = cache.get(key);
    if (existing !== undefined) {
        return existing;
    }

    const value = await fn();
    cache.set(key, value, ttl);
    return value;
}
