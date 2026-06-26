/**
 * RateLimiterMiddleware.ts
 * 
 * Simple in-memory sliding window rate limiter for tRPC routes.
 * Protects against DDoS or runaway loop scripts in production.
 */

import { TRPCError } from '@trpc/server';

interface RateLimitTracker {
    count: number;
    resetAt: number;
}

export class RateLimiterMiddleware {
    private requests: Map<string, RateLimitTracker> = new Map();
    private windowMs: number;
    private maxRequests: number;

    constructor(windowMs: number = 60000, maxRequests: number = 100) {
        this.windowMs = windowMs;
        this.maxRequests = maxRequests;

        // Cleanup stale entries occasionally
        setInterval(() => this.cleanup(), this.windowMs * 2).unref();
    }

    /**
     * The tRPC middleware function
     */
    public create() {
        return async ({ ctx, next, path }: any) => {
            // Identifier could be IP, User ID, or a custom token
            // Fallback to path if no identifiable user info is present
            const identifier = ctx?.user?.id || ctx?.req?.ip || 'anonymous';
            const key = `${identifier}:${path}`;

            const now = Date.now();
            let tracker = this.requests.get(key);

            if (!tracker || now > tracker.resetAt) {
                tracker = { count: 0, resetAt: now + this.windowMs };
                this.requests.set(key, tracker);
            }

            tracker.count++;

            if (tracker.count > this.maxRequests) {
                // Return a clear 429 Too Many Requests error
                throw new TRPCError({
                    code: 'TOO_MANY_REQUESTS',
                    message: `Rate limit exceeded for ${path}. Try again later.`,
                });
            }

            return next();
        };
    }

    private cleanup() {
        const now = Date.now();
        for (const [key, tracker] of this.requests.entries()) {
            if (now > tracker.resetAt) {
                this.requests.delete(key);
            }
        }
    }
}
