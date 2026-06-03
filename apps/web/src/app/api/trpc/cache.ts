/**
 * @file cache.ts
 * @module apps/web/src/app/api/trpc
 *
 * Lightweight in-memory cache for tRPC proxy responses.
 * Reduces load on the TS Core by caching frequently-polled
 * procedure responses for a short TTL.
 *
 * This is intentionally simple — no LRU, no persistence,
 * just a Map with TTL-based expiration checked on access.
 */

interface CacheEntry {
  data: unknown;
  status: number;
  headers: Record<string, string>;
  createdAt: number;
  ttl: number;
}

const cache = new Map<string, CacheEntry>();

/**
 * Procedures that are polled frequently by the dashboard
 * and can benefit from short-lived response caching.
 *
 * TTL values match the dashboard's polling interval for each procedure.
 */
const CACHEABLE_PROCEDURES: Record<string, number> = {
  // Polling every 3-5 seconds
  'startupStatus': 10000,  // 10s TTL so cache stays warm across 5s poll cycles
  'mcp.traffic': 3000,
  'session.list': 3000,

  // Polling every 10 seconds
  'billing.getCostHistory': 10000,
  'billing.getModelPricing': 10000,
  'billing.getTaskRoutingRules': 10000,

  // Infrequently changing data
  'mcp.getWorkingSet': 10000,
  'mcp.getJsoncEditor': 10000,
  'mcp.searchTools': 5000,
  'serverHealth.check': 5000,
};

export function getCacheTTL(procedurePath: string): number | null {
  // For batch requests, check the first procedure
  const firstProc = procedurePath.split(',')[0]?.trim() ?? '';
  return CACHEABLE_PROCEDURES[firstProc] ?? null;
}

export function getCached(procedurePath: string, input: string): CacheEntry | null {
  const key = `${procedurePath}:${input}`;
  const entry = cache.get(key);
  if (!entry) return null;

  // Check expiration
  if (Date.now() - entry.createdAt > entry.ttl) {
    cache.delete(key);
    return null;
  }

  return entry;
}

export function setCached(
  procedurePath: string,
  input: string,
  data: unknown,
  status: number,
  headers: Record<string, string>,
  ttl: number,
): void {
  const key = `${procedurePath}:${input}`;
  cache.set(key, {
    data,
    status,
    headers,
    createdAt: Date.now(),
    ttl,
  });

  // Prune expired entries periodically (keep cache size manageable)
  if (cache.size > 100) {
    const now = Date.now();
    for (const [k, v] of cache) {
      if (now - v.createdAt > v.ttl) {
        cache.delete(k);
      }
    }
  }
}
