import { EventEmitter } from 'events';
import crypto from 'crypto';

/**
 * Represents a single span event (annotation) attached to a span.
 */
interface SpanEvent {
    name: string;
    timestamp: number;
    attributes: Record<string, string | number | boolean>;
}

/**
 * Represents a single distributed tracing span.
 *
 * Follows OpenTelemetry semantic conventions:
 * - `traceId` groups all spans in a single trace
 * - `spanId` uniquely identifies this span
 * - `parentSpanId` links to the parent span (undefined for root spans)
 * - `status` is 'unset' until explicitly ended with 'ok' or 'error'
 */
interface Span {
    traceId: string;
    spanId: string;
    parentSpanId?: string;
    name: string;
    status: 'unset' | 'ok' | 'error';
    startTime: number;
    endTime?: number;
    attributes: Record<string, string | number | boolean>;
    events: SpanEvent[];
}

/**
 * Context propagation object used to create child spans.
 */
interface SpanContext {
    traceId: string;
    spanId: string;
}

/**
 * In-process distributed tracing service compatible with W3C Trace Context.
 *
 * Provides:
 * - Span lifecycle: `startSpan()`, `endSpan()`, `addSpanAttribute()`, `addSpanEvent()`
 * - `trace()` async wrapper that auto-manages span lifecycle with error capture
 * - Parent-child context propagation via `getContext()` and `SpanContext`
 * - W3C traceparent header export/parse for cross-service propagation
 * - Span queries: `getActiveSpans()`, `getRecentSpans()`, `getSpansByTrace()`
 * - EventEmitter: `spanStarted`, `spanEnded`
 * - Singleton via `TelemetryService.getInstance()`
 *
 * This is an in-process tracing implementation for local operator observability.
 * It does NOT export to external collectors (OTLP/Jaeger/etc.) — that is a
 * future enhancement. It stores spans in memory with a ring buffer capped at
 * `maxCompletedSpans` to prevent unbounded growth.
 */
export class TelemetryService extends EventEmitter {
    private static instance: TelemetryService | undefined;

    private readonly activeSpans: Map<string, Span> = new Map();
    private readonly completedSpans: Span[] = [];
    private readonly maxCompletedSpans: number;
    private readonly enabled: boolean;

    constructor(options?: { maxCompletedSpans?: number; enabled?: boolean }) {
        super();
        this.maxCompletedSpans = options?.maxCompletedSpans ?? 1000;
        this.enabled = options?.enabled ?? true;
    }

    /**
     * Get the singleton TelemetryService instance.
     */
    static getInstance(): TelemetryService {
        if (!TelemetryService.instance) {
            TelemetryService.instance = new TelemetryService();
        }
        return TelemetryService.instance;
    }

    /**
     * Start a new span. If `parentContext` is provided, the span inherits
     * the parent's `traceId` and records `parentSpanId` for tree reconstruction.
     * Otherwise, a new trace is started.
     */
    startSpan(name: string, parentContext?: SpanContext): Span {
        const spanId = this.generateId();
        const traceId = parentContext?.traceId ?? this.generateId();

        const span: Span = {
            traceId,
            spanId,
            parentSpanId: parentContext?.spanId,
            name,
            status: 'unset',
            startTime: Date.now(),
            attributes: {
                'service.name': 'tormentnexus-core',
            },
            events: [],
        };

        this.activeSpans.set(spanId, span);
        this.emit('spanStarted', { ...span });
        return span;
    }

    /**
     * End a span by its ID. Sets the status (defaults to 'unset') and
     * records the end time. Moves the span from active to completed.
     */
    endSpan(spanId: string, status?: 'ok' | 'error' | 'unset'): void {
        const span = this.activeSpans.get(spanId);
        if (!span) return;

        span.status = status ?? span.status;
        span.endTime = Date.now();
        this.activeSpans.delete(spanId);

        // Ring buffer: drop oldest completed spans when over capacity
        this.completedSpans.push(span);
        while (this.completedSpans.length > this.maxCompletedSpans) {
            this.completedSpans.shift();
        }

        this.emit('spanEnded', { ...span });
    }

    /**
     * Add a key-value attribute to an active span.
     * Attributes are used for structured metadata (user IDs, request counts, etc.).
     */
    addSpanAttribute(spanId: string, key: string, value: string | number | boolean): void {
        const span = this.activeSpans.get(spanId);
        if (span) {
            span.attributes[key] = value;
        }
    }

    /**
     * Add a timestamped event (annotation) to an active span.
     * Events are used for recording discrete occurrences within a span
     * (cache misses, retries, database queries, etc.).
     */
    addSpanEvent(spanId: string, name: string, attributes?: Record<string, string | number | boolean>): void {
        const span = this.activeSpans.get(spanId);
        if (span) {
            span.events.push({
                name,
                timestamp: Date.now(),
                attributes: attributes ?? {},
            });
        }
    }

    /**
     * Get a context object for parent-child propagation.
     * Pass this to `startSpan()` to create a child span.
     */
    getContext(spanId: string): SpanContext | undefined {
        const span = this.activeSpans.get(spanId) ?? this.completedSpans.find((s) => s.spanId === spanId);
        if (!span) return undefined;
        return { traceId: span.traceId, spanId: span.spanId };
    }

    /**
     * Convenience wrapper: executes an async function within an automatically
     * managed span. Sets status to 'ok' on success, 'error' on exception
     * (with an `exception` event), and always ends the span.
     */
    async trace<R>(
        name: string,
        fn: (span: Span) => Promise<R>,
        parentContext?: SpanContext,
    ): Promise<R> {
        const span = this.startSpan(name, parentContext);
        try {
            const result = await fn(span);
            this.endSpan(span.spanId, 'ok');
            return result;
        } catch (error) {
            this.addSpanEvent(span.spanId, 'exception', {
                'exception.message': error instanceof Error ? error.message : String(error),
            });
            this.endSpan(span.spanId, 'error');
            throw error;
        }
    }

    /**
     * Export a W3C Trace Context `traceparent` header for a given span.
     * Format: `{version}-{traceId}-{spanId}-{flags}`
     *
     * @see https://www.w3.org/TR/trace-context/
     */
    exportW3CTraceContext(spanId: string): string | undefined {
        const span = this.activeSpans.get(spanId);
        if (!span) return undefined;
        // version-traceid-spanid-flags (01 = sampled)
        return `00-${span.traceId}-${span.spanId}-01`;
    }

    /**
     * Parse a W3C Trace Context `traceparent` header into a SpanContext.
     * Returns undefined if the header is malformed.
     */
    parseW3CTraceContext(header: string): SpanContext | undefined {
        const parts = header.split('-');
        if (parts.length !== 4) return undefined;
        const [version, traceId, spanId] = parts;
        if (version !== '00' || traceId.length < 8 || spanId.length < 8) return undefined;
        return { traceId, spanId };
    }

    /**
     * Get all currently active (not yet ended) spans.
     */
    getActiveSpans(): Span[] {
        return Array.from(this.activeSpans.values());
    }

    /**
     * Get recently completed spans, optionally limited.
     */
    getRecentSpans(limit?: number): Span[] {
        if (limit !== undefined) {
            return this.completedSpans.slice(-limit);
        }
        return [...this.completedSpans];
    }

    /**
     * Get all spans (active and completed) belonging to a specific trace.
     */
    getSpansByTrace(traceId: string): Span[] {
        const result: Span[] = [];
        for (const span of this.activeSpans.values()) {
            if (span.traceId === traceId) result.push(span);
        }
        for (const span of this.completedSpans) {
            if (span.traceId === traceId) result.push(span);
        }
        return result;
    }

    /**
     * Get aggregate telemetry statistics for the dashboard.
     */
    getStats(): { activeSpans: number; completedSpans: number; enabled: boolean } {
        return {
            activeSpans: this.activeSpans.size,
            completedSpans: this.completedSpans.length,
            enabled: this.enabled,
        };
    }

    /**
     * Generate a 16-character hex ID for trace and span identifiers.
     */
    private generateId(): string {
        return crypto.randomBytes(8).toString('hex');
    }
}
