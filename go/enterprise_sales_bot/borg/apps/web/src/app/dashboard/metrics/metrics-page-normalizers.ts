export interface NormalizedMetricBucket {
    time: number;
    count: number;
    valueAvg: number;
}

export interface NormalizedMetricsData {
    totalEvents: number;
    averages: {
        memoryHeap: number | null;
        memoryRss: number | null;
        systemLoad: number | null;
    };
    countRows: Array<{ type: string; count: number }>;
    series: NormalizedMetricBucket[];
    maxSeriesCount: number;
    firstSeriesTime: number | null;
    lastSeriesTime: number | null;
}

const asFiniteNumber = (value: unknown): number | null => {
    return typeof value === 'number' && Number.isFinite(value) ? value : null;
};

const asNonNegativeNumber = (value: unknown, fallback = 0): number => {
    const parsed = asFiniteNumber(value);
    if (parsed === null) return fallback;
    return parsed >= 0 ? parsed : fallback;
};

const asRecord = (value: unknown): Record<string, unknown> => {
    return value && typeof value === 'object' ? (value as Record<string, unknown>) : {};
};

export const normalizeMetricsData = (payload: unknown): NormalizedMetricsData => {
    const source = asRecord(payload);
    const averages = asRecord(source.averages);
    const counts = asRecord(source.counts);

    const series = Array.isArray(source.series)
        ? source.series.map((rawBucket) => {
            const bucket = asRecord(rawBucket);
            return {
                time: asNonNegativeNumber(bucket.time, 0),
                count: asNonNegativeNumber(bucket.count, 0),
                valueAvg: asNonNegativeNumber(bucket.value_avg, 0),
            };
        })
        : [];

    const countRows = Object.entries(counts)
        .map(([type, rawCount]) => ({
            type,
            count: asNonNegativeNumber(rawCount, 0),
        }))
        .filter((row) => row.type.trim().length > 0);

    const maxSeriesCount = Math.max(1, ...series.map((bucket) => bucket.count));

    return {
        totalEvents: asNonNegativeNumber(source.totalEvents, 0),
        averages: {
            memoryHeap: asFiniteNumber(averages.memory_heap),
            memoryRss: asFiniteNumber(averages.memory_rss),
            systemLoad: asFiniteNumber(averages.system_load),
        },
        countRows,
        series,
        maxSeriesCount,
        firstSeriesTime: series.length > 0 ? series[0].time : null,
        lastSeriesTime: series.length > 0 ? series[series.length - 1].time : null,
    };
};
