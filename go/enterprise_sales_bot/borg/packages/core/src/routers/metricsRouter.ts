import { TRPCError } from '@trpc/server';
import { z } from 'zod';
import { t, getMcpServer } from '../lib/trpc-core.js';
import os from 'os';

type ProviderBreakdownEntry = {
    provider: string;
    cost: number;
    requests: number;
};

type ProviderBreakdownResult = {
    totalCost: number | null;
    totalRequests: number | null;
    averageLatency: number | null;
    providers: ProviderBreakdownEntry[];
    error?: string;
};

export const metricsRouter = t.router({
    /** Get aggregated metric stats for a time window */
    getStats: t.procedure.input(z.object({
        windowMs: z.number().optional()
    }).optional()).query(async ({ input }) => {
        const server = getMcpServer();
        return server.metricsService.getStats(input?.windowMs);
    }),

    /** Track a custom metric event */
    track: t.procedure.input(z.object({
        type: z.string(),
        value: z.number(),
        tags: z.record(z.string()).optional()
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        server.metricsService.track(input.type, input.value, input.tags);
        return { success: true };
    }),

    /** Get real-time system resource snapshot */
    systemSnapshot: t.procedure.query(async () => {
        const mem = process.memoryUsage();
        const cpus = os.cpus();
        const totalMem = os.totalmem();
        const freeMem = os.freemem();
        const uptime = process.uptime();

        return {
            timestamp: Date.now(),
            process: {
                heapUsed: mem.heapUsed,
                heapTotal: mem.heapTotal,
                rss: mem.rss,
                external: mem.external,
                arrayBuffers: mem.arrayBuffers,
                uptimeSeconds: Math.round(uptime),
                pid: process.pid,
            },
            system: {
                totalMemory: totalMem,
                freeMemory: freeMem,
                usedMemory: totalMem - freeMem,
                memoryUsagePercent: Math.round(((totalMem - freeMem) / totalMem) * 100),
                cpuCount: cpus.length,
                cpuModel: cpus[0]?.model ?? 'unknown',
                loadAvg: os.loadavg(),
                platform: os.platform(),
                arch: os.arch(),
                hostname: os.hostname(),
            },
        };
    }),

    /** Get timeline data for dashboard charts — downsampled time series */
    getTimeline: t.procedure.input(z.object({
        windowMs: z.number().default(3600000), // 1 hour default
        buckets: z.number().min(10).max(200).default(60),
        metricType: z.string().optional(), // filter by type
    })).query(async ({ input }) => {
        const server = getMcpServer();
        const stats = server.metricsService.getStats(input.windowMs);
        // If a specific metric type is requested, filter the series
        const series = stats.series;
        return {
            windowMs: input.windowMs,
            buckets: input.buckets,
            metricType: input.metricType ?? 'all',
            series,
            counts: stats.counts,
            averages: stats.averages,
        };
    }),

    /** Get provider-level breakdown — requests, latency, cost per provider */
    getProviderBreakdown: t.procedure.query(async () => {
        const server = getMcpServer();
        try {
            const costStats = server.llmService.getCostStats();
            const quota = server.llmService.modelSelector.getQuotaService();
            const breakdown = quota.getUsageByModel();

            const result: ProviderBreakdownResult = {
                totalCost: costStats.estimatedCostUSD,
                totalRequests: costStats.totalRequests ?? 0,
                averageLatency: costStats.averageLatencyMs ?? 0,
                providers: breakdown.length > 0 ? breakdown : [
                    { provider: 'No Usage Yet', cost: 0, requests: 0 }
                ],
            };
            return result;
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            return {
                totalCost: null,
                totalRequests: null,
                averageLatency: null,
                providers: [],
                error: `Provider metrics unavailable: ${message}`,
            } satisfies ProviderBreakdownResult;
        }
    }),

    /** Start/stop system monitoring */
    toggleMonitoring: t.procedure.input(z.object({
        enabled: z.boolean(),
        intervalMs: z.number().min(1000).max(60000).default(5000),
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        if (input.enabled) {
            server.metricsService.startMonitoring(input.intervalMs);
        } else {
            server.metricsService.stopMonitoring();
        }
        return { success: true, monitoring: input.enabled };
    }),

    /**
     * Get recent LLM routing / failover decisions.
     * Returns events newest-first, up to the ring-buffer limit (50).
     */
    getRoutingHistory: t.procedure.input(z.object({
        limit: z.number().min(1).max(50).default(20),
    }).optional()).query(async ({ input }) => {
        const server = getMcpServer();
        try {
            const history = server.llmService.getRoutingHistory();
            const limit = input?.limit ?? 20;
            return history.slice(0, limit);
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            throw new TRPCError({
                code: 'INTERNAL_SERVER_ERROR',
                message: `Routing history is unavailable: ${message}`,
            });
        }
    }),
});
