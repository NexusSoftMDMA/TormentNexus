/**
 * RateLimiter.ts
 * 
 * A token bucket rate limiter to prevent the swarm from flooding AI providers
 * and triggering HTTP 429 (Too Many Requests) errors.
 * Tracks both Requests Per Minute (RPM) and Tokens Per Minute (TPM).
 */

export class RateLimiter {
    private rpmLimit: number;
    private tpmLimit: number;

    private requestsAvailable: number;
    private tokensAvailable: number;

    private lastRefill: number;
    private isBackingOff: boolean = false;
    private backoffUntil: number = 0;

    constructor(rpmLimit: number = 50, tpmLimit: number = 40000) {
        this.rpmLimit = rpmLimit;
        this.tpmLimit = tpmLimit;
        this.requestsAvailable = rpmLimit;
        this.tokensAvailable = tpmLimit;
        this.lastRefill = Date.now();
    }

    private refill() {
        const now = Date.now();
        const elapsedMinutes = (now - this.lastRefill) / 60000;

        if (elapsedMinutes > 0) {
            this.requestsAvailable = Math.min(this.rpmLimit, this.requestsAvailable + (this.rpmLimit * elapsedMinutes));
            this.tokensAvailable = Math.min(this.tpmLimit, this.tokensAvailable + (this.tpmLimit * elapsedMinutes));
            this.lastRefill = now;
        }
    }

    /**
     * Attempts to consume capacity. Returns true if successful, false if throttled.
     */
    public tryAcquire(estimatedTokens: number = 1000): boolean {
        if (this.isBackingOff && Date.now() < this.backoffUntil) {
            return false;
        } else if (this.isBackingOff) {
            this.isBackingOff = false; // Backoff period ended
        }

        this.refill();

        if (this.requestsAvailable >= 1 && this.tokensAvailable >= estimatedTokens) {
            this.requestsAvailable -= 1;
            this.tokensAvailable -= estimatedTokens;
            return true;
        }

        return false;
    }

    /**
     * Called when a 429 error is detected. Triggers an exponential/fixed backoff.
     */
    public triggerBackoff(seconds: number = 30) {
        this.isBackingOff = true;
        this.backoffUntil = Date.now() + (seconds * 1000);

        // Severely drain current buckets to prevent immediate burst after backoff
        this.requestsAvailable = Math.max(0, this.requestsAvailable / 2);
        this.tokensAvailable = Math.max(0, this.tokensAvailable / 2);
    }

    public isThrottled(): boolean {
        this.refill();
        if (this.isBackingOff && Date.now() < this.backoffUntil) return true;
        return this.requestsAvailable < 1 || this.tokensAvailable < 100; // Arbitrary low threshold
    }

    public getBackoffRemainingMs(): number {
        if (!this.isBackingOff) return 0;
        return Math.max(0, this.backoffUntil - Date.now());
    }
}
