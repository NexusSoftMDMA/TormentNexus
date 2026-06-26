
import os from 'os';
import { EventEmitter } from 'events';

export interface MetricEvent {
    timestamp: number;
    type: string;
    value: number;
    tags?: Record<string, string>;
}

export interface MetricsServiceOptions {
    maxEvents?: number;
    monitoringIntervalMs?: number;
}

export interface HistogramStats {
    count: number;
    sum: number;
    avg: number;
    min: number;
    max: number;
    p50: number;
    p95: number;
    p99: number;
}

function labelKey(labels: Record<string, string> | undefined): string {
    if (!labels || Object.keys(labels).length === 0) return '';
    return Object.entries(labels)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([k, v]) => `${k}=${v}`)
        .join(',');
}

function compositeKey(name: string, labels?: Record<string, string>): string {
    const lk = labelKey(labels);
    return lk ? `${name}{${lk}}` : name;
}

export class MetricsService extends EventEmitter {
    private static instance: MetricsService;

    // Legacy event-stream storage
    private events: MetricEvent[] = [];
    private readonly MAX_EVENTS = 10000;
    private monitorInterval: NodeJS.Timeout | null = null;

    // Typed metric storage
    private counters: Map<string, number> = new Map();
    private counterLabels: Map<string, Record<string, string>> = new Map();
    private gauges: Map<string, number> = new Map();
    private gaugeLabels: Map<string, Record<string, string>> = new Map();
    private histograms: Map<string, number[]> = new Map();

    constructor() {
        super();
    }

    public static getInstance(): MetricsService {
        if (!MetricsService.instance) {
            MetricsService.instance = new MetricsService();
        }
        return MetricsService.instance;
    }

    // ---------- reset / dispose ----------

    public reset(): void {
        this.counters.clear();
        this.counterLabels.clear();
        this.gauges.clear();
        this.gaugeLabels.clear();
        this.histograms.clear();
        this.events = [];
        this.removeAllListeners();
    }

    public dispose(): void {
        if (this.monitorInterval) {
            clearInterval(this.monitorInterval);
            this.monitorInterval = null;
        }
        this.reset();
    }

    // ---------- counters ----------

    public incCounter(name: string, value: number = 1, labels?: Record<string, string>): void {
        const key = compositeKey(name, labels);
        this.counters.set(key, (this.counters.get(key) ?? 0) + value);
        this.counterLabels.set(key, labels ?? {});
        this.emit('metric', { name, type: 'counter', value: this.counters.get(key)!, labels: labels ?? {} });
    }

    public getCounter(name: string, labels?: Record<string, string>): number {
        return this.counters.get(compositeKey(name, labels)) ?? 0;
    }

    // ---------- gauges ----------

    public setGauge(name: string, value: number, labels?: Record<string, string>): void {
        const key = compositeKey(name, labels);
        this.gauges.set(key, value);
        this.gaugeLabels.set(key, labels ?? {});
        this.emit('metric', { name, type: 'gauge', value, labels: labels ?? {} });
    }

    public getGauge(name: string, labels?: Record<string, string>): number {
        return this.gauges.get(compositeKey(name, labels)) ?? 0;
    }

    public incGauge(name: string, value: number = 1, labels?: Record<string, string>): void {
        const key = compositeKey(name, labels);
        const current = this.gauges.get(key) ?? 0;
        this.setGauge(name, current + value, labels);
    }

    public decGauge(name: string, value: number = 1, labels?: Record<string, string>): void {
        this.incGauge(name, -value, labels);
    }

    // ---------- histograms ----------

    public observeHistogram(name: string, value: number): void {
        if (!this.histograms.has(name)) {
            this.histograms.set(name, []);
        }
        this.histograms.get(name)!.push(value);
    }

    public getHistogramStats(name: string): HistogramStats | undefined {
        const values = this.histograms.get(name);
        if (!values || values.length === 0) return undefined;

        const sorted = [...values].sort((a, b) => a - b);
        const count = sorted.length;
        const sum = sorted.reduce((a, b) => a + b, 0);

        return {
            count,
            sum,
            avg: sum / count,
            min: sorted[0],
            max: sorted[count - 1],
            p50: sorted[Math.floor(count * 0.5)],
            p95: sorted[Math.floor(count * 0.95)],
            p99: sorted[Math.floor(count * 0.99)],
        };
    }

    // ---------- timer ----------

    public timer(name: string): () => void {
        const start = performance.now();
        return () => {
            const duration = performance.now() - start;
            this.observeHistogram(name, duration);
        };
    }

    // ---------- Prometheus export ----------

    public exportPrometheus(): string {
        const lines: string[] = [];

        // Counters
        const counterNames = new Set<string>();
        for (const [key] of this.counters) {
            const baseName = key.includes('{') ? key.slice(0, key.indexOf('{')) : key;
            counterNames.add(baseName);
        }
        for (const name of counterNames) {
            lines.push(`# TYPE ${name} counter`);
            for (const [key, value] of this.counters) {
                const baseName = key.includes('{') ? key.slice(0, key.indexOf('{')) : key;
                if (baseName !== name) continue;
                const labels = this.counterLabels.get(key) ?? {};
                const labelStr = Object.entries(labels)
                    .map(([k, v]) => `${k}="${v}"`)
                    .join(',');
                lines.push(labelStr ? `${name}{${labelStr}} ${value}` : `${name} ${value}`);
            }
        }

        // Gauges
        const gaugeNames = new Set<string>();
        for (const [key] of this.gauges) {
            const baseName = key.includes('{') ? key.slice(0, key.indexOf('{')) : key;
            gaugeNames.add(baseName);
        }
        for (const name of gaugeNames) {
            lines.push(`# TYPE ${name} gauge`);
            for (const [key, value] of this.gauges) {
                const baseName = key.includes('{') ? key.slice(0, key.indexOf('{')) : key;
                if (baseName !== name) continue;
                const labels = this.gaugeLabels.get(key) ?? {};
                const labelStr = Object.entries(labels)
                    .map(([k, v]) => `${k}="${v}"`)
                    .join(',');
                lines.push(labelStr ? `${name}{${labelStr}} ${value}` : `${name} ${value}`);
            }
        }

        return lines.join('\n') + '\n';
    }

    // ---------- getAll ----------

    public getAll(): { counters: Record<string, number>; gauges: Record<string, number>; histograms: Record<string, HistogramStats> } {
        const counters: Record<string, number> = {};
        for (const [key, value] of this.counters) {
            const baseName = key.includes('{') ? key.slice(0, key.indexOf('{')) : key;
            counters[baseName] = (counters[baseName] ?? 0) + value;
        }
        const gauges: Record<string, number> = {};
        for (const [key, value] of this.gauges) {
            const baseName = key.includes('{') ? key.slice(0, key.indexOf('{')) : key;
            gauges[baseName] = value;
        }
        const histograms: Record<string, HistogramStats> = {};
        for (const name of this.histograms.keys()) {
            const stats = this.getHistogramStats(name);
            if (stats) histograms[name] = stats;
        }
        return { counters, gauges, histograms };
    }

    // ---------- legacy API (backward compat for metricsRouter / MCPServer) ----------

    public startMonitoring(intervalMs: number = 5000) {
        if (this.monitorInterval) clearInterval(this.monitorInterval);

        this.monitorInterval = setInterval(() => {
            const mem = process.memoryUsage();
            const totalMem = os.totalmem();
            const freeMem = os.freemem();

            this.track('memory_heap', mem.heapUsed);
            this.track('memory_rss', mem.rss);
            this.track('system_load', os.loadavg()[0]);
            this.track('system_free_mem', freeMem);
        }, intervalMs);
    }

    public stopMonitoring() {
        if (this.monitorInterval) clearInterval(this.monitorInterval);
        this.monitorInterval = null;
    }

    track(type: string, value: number, tags?: Record<string, string>) {
        this.events.push({
            timestamp: Date.now(),
            type,
            value,
            tags
        });

        if (this.events.length > this.MAX_EVENTS) {
            this.events = this.events.slice(-this.MAX_EVENTS / 2);
        }

        // Bridge to typed metrics
        if (type === 'duration' || type === 'tool_execution') {
            const name = tags?.tool || tags?.name || type;
            this.observeHistogram(`duration_${name}`, value);
        } else if (type === 'memory_heap' || type === 'memory_rss' || type === 'system_load' || type === 'system_free_mem') {
            this.setGauge(type, value, tags);
        } else {
            const name = tags?.tool || tags?.name || type;
            this.incCounter(name, value, tags);
        }
    }

    trackDuration(name: string, ms: number, tags?: Record<string, string>) {
        this.track('duration', ms, { name, ...tags });
    }

    getStats(windowMs: number = 3600000) {
        const now = Date.now();
        const relevant = this.events.filter(e => e.timestamp > now - windowMs);

        const counts: Record<string, number> = {};
        const sums: Record<string, number> = {};

        relevant.forEach(e => {
            const key = e.type;
            counts[key] = (counts[key] || 0) + e.value; // For tool_call, value is usually 1, but we should sum it
            sums[key] = (sums[key] || 0) + e.value;
        });

        const averages: Record<string, number> = {};
        Object.keys(sums).forEach(k => {
            averages[k] = sums[k] / (relevant.filter(e => e.type === k).length || 1);
        });

        // Enrich with detailed histograms and typed stats
        const typedStats = this.getAll();

        return {
            windowMs,
            totalEvents: relevant.length,
            counts,
            averages,
            counters: typedStats.counters,
            gauges: typedStats.gauges,
            histograms: typedStats.histograms,
            series: this.downsample(relevant, 60)
        };
    }

    private downsample(events: MetricEvent[], buckets: number) {
        if (events.length === 0) return [];
        const start = events[0].timestamp;
        const end = Date.now();
        const interval = (end - start) / buckets;

        const res = [];
        for (let i = 0; i < buckets; i++) {
            const bucketStart = start + (i * interval);
            const bucketEnd = bucketStart + interval;
            const inBucket = events.filter(e => e.timestamp >= bucketStart && e.timestamp < bucketEnd);
            res.push({
                time: bucketStart,
                count: inBucket.length,
                value_avg: inBucket.reduce((acc, curr) => acc + curr.value, 0) / (inBucket.length || 1)
            });
        }
        return res;
    }
}
