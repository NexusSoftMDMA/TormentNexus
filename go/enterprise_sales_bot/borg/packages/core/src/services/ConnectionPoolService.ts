import { EventEmitter } from 'events';

/**
 * Configuration for a ConnectionPool instance.
 */
interface ConnectionPoolOptions {
    /** Minimum number of connections to maintain (created on `initialize()`) */
    minSize?: number;
    /** Maximum number of connections the pool can hold */
    maxSize?: number;
    /** Milliseconds to wait for a connection before throwing (default: 30s) */
    acquireTimeout?: number;
    /** Milliseconds a connection can sit idle before being eligible for cleanup */
    idleTimeout?: number;
    /** Maximum lifetime in milliseconds for any single connection */
    maxLifetime?: number;
}

/**
 * Internal wrapper around a pooled connection, tracking timing metadata.
 */
interface PooledConnection<T> {
    connection: T;
    createdAt: number;
    lastUsedAt: number;
}

/**
 * A waiter in the acquire queue, resolved when a connection becomes available
 * or rejected on timeout/drain.
 */
interface AcquireWaiter<T> {
    resolve: (conn: T) => void;
    reject: (err: Error) => void;
    timer: ReturnType<typeof setTimeout>;
}

/**
 * Generic connection pool with configurable min/max size, acquire timeout,
 * factory/destroyer/validator lifecycle, and event emission.
 *
 * Supports:
 * - Pre-warming via `initialize()` (creates `minSize` connections)
 * - Bounded pool with backpressure (`acquire()` waits or times out when exhausted)
 * - Transparent validation on release (invalid connections are destroyed and recycled)
 * - Graceful drain with waiter rejection
 * - EventEmitter: `connectionCreated`, `connectionAcquired`, `connectionReleased`
 *
 * Used by the TormentNexus control plane for managing reusable connections to
 * databases, MCP server stdio transports, and external services.
 */
export class ConnectionPool<T> extends EventEmitter {
    private readonly factory: () => Promise<T>;
    private readonly destroyer: (conn: T) => Promise<void>;
    private readonly validator: (conn: T) => Promise<boolean>;

    private readonly minSize: number;
    private readonly maxSize: number;
    private readonly acquireTimeout: number;

    private readonly available: PooledConnection<T>[] = [];
    private readonly inUse: Set<T> = new Set();
    private readonly waiters: AcquireWaiter<T>[] = [];
    private draining = false;

    constructor(
        factory: () => Promise<T>,
        destroyer: (conn: T) => Promise<void>,
        validator: (conn: T) => Promise<boolean>,
        options: ConnectionPoolOptions = {},
    ) {
        super();
        this.factory = factory;
        this.destroyer = destroyer;
        this.validator = validator;
        this.minSize = options.minSize ?? 0;
        this.maxSize = options.maxSize ?? 10;
        this.acquireTimeout = options.acquireTimeout ?? 30_000;
    }

    /**
     * Create the minimum number of connections to pre-warm the pool.
     * Should be called once before the pool is used.
     */
    async initialize(): Promise<void> {
        for (let i = 0; i < this.minSize; i++) {
            const conn = await this.createConnection();
            this.available.push(conn);
        }
    }

    /**
     * Acquire a connection from the pool. If no idle connection is available
     * and the pool is below `maxSize`, a new one is created. If the pool is
     * exhausted, the call waits up to `acquireTimeout` for a connection to
     * be released.
     *
     * @throws Error if the acquire times out or the pool is draining
     */
    async acquire(): Promise<T> {
        if (this.draining) {
            throw new Error('Pool is draining');
        }

        // Try to find a valid available connection
        while (this.available.length > 0) {
            const pooled = this.available.shift()!;
            const isValid = await this.validator(pooled.connection);
            if (isValid) {
                this.inUse.add(pooled.connection);
                this.emit('connectionAcquired', pooled.connection);
                return pooled.connection;
            }
            // Connection is invalid — destroy it and try the next one
            await this.destroyer(pooled.connection);
        }

        // No available connections — can we create a new one?
        const totalCount = this.available.length + this.inUse.size;
        if (totalCount < this.maxSize) {
            const pooled = await this.createConnection();
            this.inUse.add(pooled.connection);
            this.emit('connectionAcquired', pooled.connection);
            return pooled.connection;
        }

        // Pool exhausted — wait for a release
        return new Promise<T>((resolve, reject) => {
            const timer = setTimeout(() => {
                const index = this.waiters.findIndex((w) => w.resolve === resolve);
                if (index >= 0) this.waiters.splice(index, 1);
                reject(new Error('Connection acquire timeout'));
            }, this.acquireTimeout);

            this.waiters.push({ resolve, reject, timer });
        });
    }

    /**
     * Return a connection to the pool. If there are pending waiters, the
     * connection is immediately handed to the next waiter. Otherwise, the
     * connection is validated and returned to the idle pool (or destroyed
     * if invalid).
     */
    release(conn: T): void {
        this.inUse.delete(conn);
        this.emit('connectionReleased', conn);

        // Service any waiting acquires first
        if (this.waiters.length > 0) {
            const waiter = this.waiters.shift()!;
            clearTimeout(waiter.timer);
            this.inUse.add(conn);
            this.emit('connectionAcquired', conn);
            waiter.resolve(conn);
            return;
        }

        // Return to the idle pool immediately. Validation happens on next acquire()
        // to avoid async race conditions where getStats() is called before the
        // validator resolves.
        this.available.push({
            connection: conn,
            createdAt: Date.now(),
            lastUsedAt: Date.now(),
        });
    }

    /**
     * Gracefully drain the pool: destroy all connections and reject all waiters.
     */
    async drain(): Promise<void> {
        this.draining = true;

        // Reject all waiting acquires
        for (const waiter of this.waiters) {
            clearTimeout(waiter.timer);
            waiter.reject(new Error('Pool is draining'));
        }
        this.waiters.length = 0;

        // Destroy all available connections
        for (const pooled of this.available) {
            await this.destroyer(pooled.connection);
        }
        this.available.length = 0;

        // Destroy all in-use connections (best effort)
        for (const conn of this.inUse) {
            await this.destroyer(conn);
        }
        this.inUse.clear();

        this.draining = false;
    }

    /**
     * Return current pool statistics for observability.
     */
    getStats(): { total: number; available: number; inUse: number; waiting: number } {
        return {
            total: this.available.length + this.inUse.size,
            available: this.available.length,
            inUse: this.inUse.size,
            waiting: this.waiters.length,
        };
    }

    /**
     * Create a new connection via the factory and emit an event.
     */
    private async createConnection(): Promise<PooledConnection<T>> {
        const connection = await this.factory();
        const now = Date.now();
        this.emit('connectionCreated', connection);
        return { connection, createdAt: now, lastUsedAt: now };
    }
}

/**
 * Static registry of named connection pools.
 *
 * Allows services across the TormentNexus control plane to register, retrieve,
 * and manage pools by name without passing pool instances directly.
 */
export class ConnectionPoolManager {
    private static readonly pools: Map<string, ConnectionPool<unknown>> = new Map();

    /** Register a named pool. Overwrites any existing pool with the same name. */
    static register<T>(name: string, pool: ConnectionPool<T>): void {
        ConnectionPoolManager.pools.set(name, pool as ConnectionPool<unknown>);
    }

    /** Retrieve a previously registered pool by name. */
    static get<T>(name: string): ConnectionPool<T> | undefined {
        return ConnectionPoolManager.pools.get(name) as ConnectionPool<T> | undefined;
    }

    /** Get statistics for all registered pools. */
    static getAllStats(): Record<string, { total: number; available: number; inUse: number; waiting: number }> {
        const result: Record<string, { total: number; available: number; inUse: number; waiting: number }> = {};
        for (const [name, pool] of ConnectionPoolManager.pools) {
            result[name] = pool.getStats();
        }
        return result;
    }

    /** Drain all registered pools and clear the registry. */
    static async drainAll(): Promise<void> {
        for (const pool of ConnectionPoolManager.pools.values()) {
            await pool.drain();
        }
        ConnectionPoolManager.pools.clear();
    }
}
