/**
 * AuthMiddleware.ts
 * 
 * Secures tRPC procedures using bearer tokens or API keys.
 * Integrates into the central policy engine for high-tier validations.
 */

import { TRPCError } from '@trpc/server';
import crypto from 'crypto';

export class AuthMiddleware {
    private validKeys: Set<string>;
    private useStrictTokenHash: boolean;

    /**
     * @param allowedKeys List of valid static API keys
     * @param requireHash If true, compares timing-safe hashes
     */
    constructor(allowedKeys: string[] = [], requireHash: boolean = true) {
        this.validKeys = new Set(allowedKeys);
        this.useStrictTokenHash = requireHash;
    }

    /**
     * Dynamically update valid keys (e.g. from DB)
     */
    public addKey(key: string) {
        this.validKeys.add(key);
    }

    public removeKey(key: string) {
        this.validKeys.delete(key);
    }

    /**
     * The tRPC generic authorization middleware
     */
    public create() {
        return async ({ ctx, next, path }: any) => {
            // Extract token from context (usually passed in from Express/Fastify headers)
            // Example: Authorization: Bearer <token>
            const authHeader = ctx?.req?.headers?.authorization;

            if (!authHeader || typeof authHeader !== 'string') {
                throw new TRPCError({
                    code: 'UNAUTHORIZED',
                    message: `Missing authorization header for protected route: ${path}`,
                });
            }

            const token = authHeader.replace(/^Bearer\s+/i, '').trim();

            if (!this.isValid(token)) {
                throw new TRPCError({
                    code: 'UNAUTHORIZED',
                    message: `Invalid or revoked API key.`,
                });
            }

            // Inject the validated identity context downstream
            ctx.user = { isAuthenticated: true, keyHint: token.slice(-4) };

            return next({ ctx });
        };
    }

    private isValid(token: string): boolean {
        if (!this.useStrictTokenHash) {
            return this.validKeys.has(token);
        }

        // Timing-safe comparison to prevent timing attacks on keys
        for (const storedKey of this.validKeys) {
            if (storedKey.length === token.length) {
                if (crypto.timingSafeEqual(Buffer.from(storedKey), Buffer.from(token))) {
                    return true;
                }
            }
        }
        return false;
    }
}
