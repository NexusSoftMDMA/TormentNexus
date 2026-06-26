'use client';

import { trpc } from "@/utils/trpc";
import { normalizeMetricsData } from './metrics-page-normalizers';


export default function MetricsPage() {
    const { data, error, isLoading } = trpc.metrics.getStats.useQuery(
        { windowMs: 3600000 },
        { refetchInterval: 5000 }
    );

    const formatBytes = (b: number) => {
        if (b > 1073741824) return `${(b / 1073741824).toFixed(1)} GB`;
        if (b > 1048576) return `${(b / 1048576).toFixed(1)} MB`;
        if (b > 1024) return `${(b / 1024).toFixed(1)} KB`;
        return `${b} B`;
    };

    const normalized = normalizeMetricsData(data);

    return (
        <div className="p-6 space-y-6">
            <div className="flex flex-col gap-2">
                <h1 className="text-3xl font-bold tracking-tight">📊 Metrics</h1>
                <p className="text-muted-foreground">
                    System performance telemetry and resource monitoring.
                </p>
            </div>

            {error && (
                <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 text-destructive">
                    Error: {error.message}
                </div>
            )}

            {isLoading && (
                <div className="text-muted-foreground animate-pulse">Loading metrics...</div>
            )}

            {data && (
                <>
                    {/* Summary Cards */}
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                        <div className="bg-card border rounded-lg p-4">
                            <div className="text-sm text-muted-foreground">Total Events</div>
                            <div className="text-2xl font-bold">{normalized.totalEvents.toLocaleString()}</div>
                        </div>
                        <div className="bg-card border rounded-lg p-4">
                            <div className="text-sm text-muted-foreground">Heap Usage</div>
                            <div className="text-2xl font-bold">
                                {normalized.averages.memoryHeap !== null ? formatBytes(normalized.averages.memoryHeap) : 'N/A'}
                            </div>
                        </div>
                        <div className="bg-card border rounded-lg p-4">
                            <div className="text-sm text-muted-foreground">RSS</div>
                            <div className="text-2xl font-bold">
                                {normalized.averages.memoryRss !== null ? formatBytes(normalized.averages.memoryRss) : 'N/A'}
                            </div>
                        </div>
                        <div className="bg-card border rounded-lg p-4">
                            <div className="text-sm text-muted-foreground">System Load</div>
                            <div className="text-2xl font-bold">
                                {normalized.averages.systemLoad !== null ? normalized.averages.systemLoad.toFixed(2) : 'N/A'}
                            </div>
                        </div>
                    </div>

                    {/* Event Type Breakdown */}
                    <div className="bg-card border rounded-lg p-6">
                        <h2 className="text-lg font-semibold mb-4">Event Type Breakdown</h2>
                        <div className="space-y-2">
                            {normalized.countRows.map(({ type, count }) => (
                                <div key={type} className="flex justify-between items-center">
                                    <span className="font-mono text-sm">{type}</span>
                                    <div className="flex items-center gap-4">
                                        <div className="w-32 bg-muted rounded-full h-2">
                                            <div
                                                className="bg-primary rounded-full h-2"
                                                style={{ width: `${Math.min(100, normalized.totalEvents > 0 ? (count / normalized.totalEvents) * 100 : 0)}%` }}
                                            />
                                        </div>
                                        <span className="text-sm text-muted-foreground w-16 text-right">
                                            {count}
                                        </span>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Sparkline Chart */}
                    {normalized.series.length > 0 && (
                        <div className="bg-card border rounded-lg p-6">
                            <h2 className="text-lg font-semibold mb-4">Activity Over Time</h2>
                            <div className="flex items-end h-32 gap-px">
                                {normalized.series.map((bucket, i: number) => {
                                    const height = (bucket.count / normalized.maxSeriesCount) * 100;
                                    return (
                                        <div
                                            key={i}
                                            className="flex-1 bg-primary/60 hover:bg-primary rounded-t transition-colors"
                                            style={{ height: `${Math.max(2, height)}%` }}
                                            title={`${new Date(bucket.time).toLocaleTimeString()}: ${bucket.count} events`}
                                        />
                                    );
                                })}
                            </div>
                            <div className="flex justify-between text-xs text-muted-foreground mt-1">
                                <span>{normalized.firstSeriesTime !== null ? new Date(normalized.firstSeriesTime).toLocaleTimeString() : 'N/A'}</span>
                                <span>{normalized.lastSeriesTime !== null ? new Date(normalized.lastSeriesTime).toLocaleTimeString() : 'N/A'}</span>
                            </div>
                        </div>
                    )}
                </>
            )}
        </div>
    );
}
