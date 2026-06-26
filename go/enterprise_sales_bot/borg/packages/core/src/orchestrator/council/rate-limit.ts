import type { Context, Next } from 'hono';

interface RateLimitConfig {
  windowMs: number;
  maxRequests: number;
}

interface RateLimitEntry {
  count: number;
  resetAt: number;
}

const rateLimitStore = new Map<string, RateLimitEntry>();

const DEFAULT_CONFIG: RateLimitConfig = {
  windowMs: 60_000,
  maxRequests: 100,
};

function getClientId(c: Context): string {
  return c.req.header('x-forwarded-for') || 
         c.req.header('x-real-ip') || 
         'unknown';
}

function cleanupExpired(): void {
  const now = Date.now();
  for (const [key, entry] of rateLimitStore) {
    if (entry.resetAt <= now) {
      rateLimitStore.delete(key);
    }
  }
}

export function rateLimit(config: Partial<RateLimitConfig> = {}) {
  const { windowMs, maxRequests } = { ...DEFAULT_CONFIG, ...config };
  
  setInterval(cleanupExpired, windowMs);
  
  return async (c: Context, next: Next) => {
    const clientId = getClientId(c);
    const key = `${clientId}:${c.req.path}`;
    const now = Date.now();
    
    let entry = rateLimitStore.get(key);
    
    if (!entry || entry.resetAt <= now) {
      entry = { count: 0, resetAt: now + windowMs };
      rateLimitStore.set(key, entry);
    }
    
    entry.count++;
    
    const remaining = Math.max(0, maxRequests - entry.count);
    const resetIn = Math.ceil((entry.resetAt - now) / 1000);
    
    c.header('X-RateLimit-Limit', String(maxRequests));
    c.header('X-RateLimit-Remaining', String(remaining));
    c.header('X-RateLimit-Reset', String(resetIn));
    
    if (entry.count > maxRequests) {
      return c.json(
        { 
          error: 'Too many requests', 
          retryAfter: resetIn 
        }, 
        429
      );
    }
    
    await next();
  };
}

export function debateRateLimit() {
  return rateLimit({ windowMs: 60_000, maxRequests: 10 });
}

export function apiRateLimit() {
  return rateLimit({ windowMs: 60_000, maxRequests: 100 });
}
