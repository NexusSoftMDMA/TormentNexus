/**
 * Get class names for status badges based on tone.
 */
export const getStatusBadgeClasses = (status: string): string => {
    switch (status) {
        case 'success':
        case 'verified':
        case 'connected':
            return 'border-emerald-500/20 bg-emerald-500/10 text-emerald-300';
        case 'warning':
        case 'detected':
        case 'pending':
            return 'border-amber-500/20 bg-amber-500/10 text-amber-300';
        case 'error':
        case 'missing':
        case 'failed':
            return 'border-red-500/20 bg-red-500/10 text-red-300';
        default:
            return 'border-zinc-700 bg-zinc-800 text-zinc-400';
    }
};

/**
 * Get cards for CLI harnesses.
 */
export const getCliHarnessCards = (detections: any[], sessions: any[]): any[] => {
    if (!Array.isArray(detections)) return [];
    return detections.map((d: any) => {
        const running = (sessions || []).filter((s: any) => s.cliType === d.id && s.status === 'running').length;
        const total = (sessions || []).filter((s: any) => s.cliType === d.id).length;
        
        return {
            ...d,
            statusTone: d.installed ? 'success' : 'muted',
            statusLabel: d.installed ? 'Installed' : 'Missing',
            runningSessions: running,
            activeSessions: total
        };
    });
};

/**
 * Get cards for Provider directory.
 */
export const getProviderDirectoryCards = (quotas: any[]): any[] => {
    if (!Array.isArray(quotas)) return [];
    return quotas.map((q: any) => ({
        provider: q.provider,
        label: q.name,
        statusTone: q.authenticated ? 'success' : 'error',
        statusLabel: q.authenticated ? 'Authenticated' : 'Unauthenticated',
        authLabel: q.authMethod,
        availabilityLabel: q.availability,
        href: `/dashboard/providers?search=${q.provider}`,
        usageLabel: q.limit ? `${q.used} / ${q.limit}` : `${q.used} used`,
        resetLabel: q.resetDate || 'None'
    }));
};
