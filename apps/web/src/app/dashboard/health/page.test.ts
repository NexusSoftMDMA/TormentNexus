import { describe, expect, it } from 'vitest';

import { getEventBusMetric, getMcpRouterMetric } from './health-metrics';

describe('health page helpers', () => {
    it('reports healthy MCP router status once the router is initialized', () => {
        expect(getMcpRouterMetric({ ready: true } as any, true)).toEqual({
            status: 'Healthy',
            color: 'text-green-500',
            detail: 'Cached inventory advertised to clients',
        });
    });

    it('reports initializing MCP router status while startup is still warming', () => {
        expect(getMcpRouterMetric({ ready: false, summary: 'Waiting for router warmup' } as any, false)).toEqual({
            status: 'Initializing',
            color: 'text-yellow-500',
            detail: 'Waiting for router warmup',
        });
    });

    it('reports degraded MCP router status when startup is serving compat fallback', () => {
        expect(getMcpRouterMetric({
            ready: false,
            status: 'degraded',
            summary: 'Using local MCP config fallback for 64 configured server(s); live startup telemetry is unavailable.',
        } as any, false)).toEqual({
            status: 'Degraded',
            color: 'text-amber-500',
            detail: 'Using local MCP config fallback for 64 configured server(s); live startup telemetry is unavailable.',
        });
    });

    it('reports active event-bus status once startup is ready', () => {
        expect(getEventBusMetric({ ready: true } as any)).toEqual({
            status: 'Active',
            color: 'text-green-500',
            detail: 'In-process pub/sub',
        });
    });

    it('reports starting event-bus status while startup is still warming', () => {
        expect(getEventBusMetric({ ready: false } as any)).toEqual({
            status: 'Starting',
            color: 'text-yellow-500',
            detail: 'In-process pub/sub',
        });
    });

    it('reports degraded event-bus status when startup is serving compat fallback', () => {
        expect(getEventBusMetric({
            ready: false,
            status: 'degraded',
            summary: 'Using local MCP config fallback for 64 configured server(s); live startup telemetry is unavailable.',
        } as any)).toEqual({
            status: 'Degraded',
            color: 'text-amber-500',
            detail: 'Using local MCP config fallback for 64 configured server(s); live startup telemetry is unavailable.',
        });
    });

    it('handles malformed non-string summaries without crashing', () => {
        expect(getMcpRouterMetric({ ready: false, summary: 42 } as any, false)).toEqual({
            status: 'Initializing',
            color: 'text-yellow-500',
            detail: 'Waiting for cached inventory',
        });

        expect(getMcpRouterMetric({ ready: false, status: 'degraded', summary: { bad: true } } as any, false)).toEqual({
            status: 'Degraded',
            color: 'text-amber-500',
            detail: 'Live startup telemetry is unavailable while TormentNexus serves a compat-fallback router snapshot.',
        });

        expect(getEventBusMetric({ ready: false, status: 'degraded', summary: ['bad'] } as any)).toEqual({
            status: 'Degraded',
            color: 'text-amber-500',
            detail: 'Live startup telemetry is unavailable while TormentNexus serves a compat-fallback snapshot.',
        });
    });
});