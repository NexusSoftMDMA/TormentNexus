
import { EventEmitter } from 'events';

export type SystemEventType =
    | 'agent:heartbeat'
    | 'agent:start'
    | 'agent:stop'
    | 'task:update'
    | 'task:complete'
    | 'tool:call'
    | 'memory:prune'
    | 'file:change'
    | 'terminal:error';

export interface SystemEvent {
    type: SystemEventType | string;
    timestamp: number;
    source: string; // Agent ID or Service Name
    payload: unknown;
}

export class EventBus extends EventEmitter {
    private wildcardListeners: Array<{ pattern: RegExp, listener: (event: SystemEvent) => void }> = [];
    private history: SystemEvent[] = [];
    private readonly MAX_HISTORY = 1000;

    constructor() {
        super();
        this.setMaxListeners(50);
    }

    public subscribe(pattern: string, listener: (event: SystemEvent) => void) {
        if (pattern.includes('*')) {
            // Convert "file:*" to /^file:.*$/
            const regex = new RegExp(`^${pattern.replace(/\*/g, '.*')}$`);
            this.wildcardListeners.push({ pattern: regex, listener });
        } else {
            // Standard exact match — EventEmitter expects string|symbol
            super.on(pattern, listener as (...args: unknown[]) => void);
        }
    }

    public emitEvent(type: SystemEventType | string, source: string, payload: unknown) {
        const event: SystemEvent = {
            type,
            timestamp: Date.now(),
            source,
            payload
        };

        // 0. Store in history
        this.history.push(event);
        if (this.history.length > this.MAX_HISTORY) {
            this.history.shift();
        }

        // 1. Emit exact match
        this.emit('system_event', event);
        this.emit(type, event);

        // 2. Check wildcards
        for (const { pattern, listener } of this.wildcardListeners) {
            if (pattern.test(type)) {
                try {
                    listener(event);
                } catch (e) {
                    console.error(`[EventBus] Error in wildcard listener for ${type}:`, e);
                }
            }
        }
    }

    public onEvent(type: SystemEventType | string, listener: (event: SystemEvent) => void) {
        this.subscribe(type, listener);
    }

    public getHistory(limit: number = 100): SystemEvent[] {
        return this.history.slice(-limit);
    }

    public getHistorySince(timestamp: number): SystemEvent[] {
        return this.history.filter(e => e.timestamp > timestamp);
    }

    /**
     * Async iterator for pattern-matched events.
     */
    public async *events(pattern: string): AsyncIterableIterator<[string, any]> {
        const queue: Array<[string, any]> = [];
        let resolve: (() => void) | null = null;

        const listener = (event: SystemEvent) => {
            queue.push([event.type, event.payload]);
            if (resolve) {
                resolve();
                resolve = null;
            }
        };

        this.subscribe(pattern, listener);

        try {
            while (true) {
                if (queue.length === 0) {
                    await new Promise<void>(r => { resolve = r; });
                }
                while (queue.length > 0) {
                    yield queue.shift()!;
                }
            }
        } finally {
            // Cleanup: remove listener would require more implementation details
            // For now, this is a simplified version for the subscription bridge.
        }
    }
}
