import { describe, expect, it } from 'vitest';

import { normalizeMetricsData } from './metrics-page-normalizers';

describe('metrics page normalizers', () => {
  it('normalizes malformed metrics payload into safe defaults', () => {
    const normalized = normalizeMetricsData({
      totalEvents: 'bad',
      averages: {
        memory_heap: 1024,
        memory_rss: null,
        system_load: 'n/a',
      },
      counts: {
        requests: 8,
        errors: 'oops',
        '': 99,
      },
      series: [
        { time: 1710000000000, count: 5, value_avg: 1.2 },
        { time: 'bad', count: -4, value_avg: undefined },
        null,
      ],
    } as any);

    expect(normalized.totalEvents).toBe(0);
    expect(normalized.averages).toEqual({
      memoryHeap: 1024,
      memoryRss: null,
      systemLoad: null,
    });
    expect(normalized.countRows).toEqual([
      { type: 'requests', count: 8 },
      { type: 'errors', count: 0 },
    ]);
    expect(normalized.series).toEqual([
      { time: 1710000000000, count: 5, valueAvg: 1.2 },
      { time: 0, count: 0, valueAvg: 0 },
      { time: 0, count: 0, valueAvg: 0 },
    ]);
    expect(normalized.maxSeriesCount).toBe(5);
    expect(normalized.firstSeriesTime).toBe(1710000000000);
    expect(normalized.lastSeriesTime).toBe(0);
  });

  it('returns safe empty structures for non-object payload', () => {
    const normalized = normalizeMetricsData(undefined);

    expect(normalized).toEqual({
      totalEvents: 0,
      averages: {
        memoryHeap: null,
        memoryRss: null,
        systemLoad: null,
      },
      countRows: [],
      series: [],
      maxSeriesCount: 1,
      firstSeriesTime: null,
      lastSeriesTime: null,
    });
  });
});
