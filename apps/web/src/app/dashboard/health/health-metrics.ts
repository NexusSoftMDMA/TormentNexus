import type { DashboardStartupStatus } from '../dashboard-home-view';

function getStartupSummary(startupStatus?: DashboardStartupStatus): string {
    return typeof startupStatus?.summary === 'string'
        ? startupStatus.summary.trim()
        : '';
}

export function getMcpRouterMetric(
    startupStatus?: DashboardStartupStatus,
    mcpInitialized?: boolean,
): {
    status: string;
    color: string;
    detail: string;
} {
    const summary = getStartupSummary(startupStatus);

    if (startupStatus?.status === 'degraded') {
        return {
            status: 'Degraded',
            color: 'text-amber-500',
            detail: summary || 'Live startup telemetry is unavailable while TormentNexus serves a compat-fallback router snapshot.',
        };
    }

    if (mcpInitialized) {
        return {
            status: 'Healthy',
            color: 'text-green-500',
            detail: 'Cached inventory advertised to clients',
        };
    }

    return {
        status: 'Initializing',
        color: 'text-yellow-500',
        detail: summary || 'Waiting for cached inventory',
    };
}

export function getEventBusMetric(startupStatus?: DashboardStartupStatus): {
    status: string;
    color: string;
    detail: string;
} {
    const summary = getStartupSummary(startupStatus);

    if (startupStatus?.status === 'degraded') {
        return {
            status: 'Degraded',
            color: 'text-amber-500',
            detail: summary || 'Live startup telemetry is unavailable while TormentNexus serves a compat-fallback snapshot.',
        };
    }

    if (startupStatus?.ready) {
        return {
            status: 'Active',
            color: 'text-green-500',
            detail: 'In-process pub/sub',
        };
    }

    return {
        status: 'Starting',
        color: 'text-yellow-500',
        detail: 'In-process pub/sub',
    };
}