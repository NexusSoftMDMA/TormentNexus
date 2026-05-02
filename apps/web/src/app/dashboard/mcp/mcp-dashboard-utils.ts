export type ManagedServerDiscoveryRecord = {
    uuid?: string;
    name: string;
    _meta?: {
        status?: string | null;
        metadataSource?: string | null;
        toolCount?: number | null;
        lastSuccessfulBinaryLoadAt?: string | null;
    } | null;
    always_on?: boolean;
};

export type BulkMetadataTargetMode = 'all' | 'unresolved';

export type ManagedServerDiscoverySummary = {
    totalCount: number;
    readyCount: number;
    unresolvedCount: number;
    staleReadyCount: number;
    repairableCount: number;
    neverLoadedCount: number;
    localCompatCount: number;
};

export type ServerToolActionLinks = {
    inspectToolsHref: string;
    editToolsHref: string;
    logsHref: string;
};

export type RuntimeServerRecord = {
    name: string;
    status: string;
    toolCount: number;
    runtimeState?: string;
    warmupState?: string;
    runtimeConnected?: boolean;
    advertisedToolCount?: number;
    advertisedSource?: string;
    lastConnectedAt?: string | null;
    lastError?: string | null;
    config?: {
        command?: string;
        args?: string[];
        env?: string[];
    };
};

export type ManagedServerRuntimeRecord = ManagedServerDiscoveryRecord & {
    description?: string | null;
    command?: string | null;
    args?: string[];
    env?: Record<string, string>;
    source_published_server_uuid?: string | null;
    _meta?: (ManagedServerDiscoveryRecord['_meta'] & {
        metadataSource?: string | null;
        toolCount?: number | null;
    }) | null;
};

export type DashboardServerRecord = {
    uuid?: string;
    name: string;
    status: string;
    toolCount: number;
    runtimeState?: string;
    warmupState?: string;
    runtimeConnected?: boolean;
    advertisedToolCount?: number;
    advertisedSource?: string;
    lastConnectedAt?: string | null;
    lastError?: string | null;
    metadataStatus?: string;
    metadataSource?: string;
    metadataToolCount?: number;
    lastSuccessfulBinaryLoadAt?: string;
    always_on?: boolean;
    source_published_server_uuid?: string | null;
    config?: {
        command?: string;
        args?: string[];
        env?: string[];
    };
};

function normalizeDiscoveryStatus(status?: string | null): string {
    return status?.trim().toLowerCase() || 'pending';
}

function normalizeServerName(name?: string | null): string {
    return name?.trim().toLowerCase() || '';
}

function normalizeToolCount(toolCount?: number | null): number {
    return typeof toolCount === 'number' && Number.isFinite(toolCount)
        ? toolCount
        : 0;
}

export function isLocalCompatMetadataSource(metadataSource?: string | null): boolean {
    const normalizedSource = metadataSource?.trim().toLowerCase() || '';
    return normalizedSource.startsWith('local-');
}

export function hasStaleReadyMetadata(server: ManagedServerDiscoveryRecord): boolean {
    return normalizeDiscoveryStatus(server._meta?.status) === 'ready'
        && normalizeToolCount(server._meta?.toolCount) === 0;
}

export function getManagedServerDiscoverySummary(
    servers: ManagedServerDiscoveryRecord[],
): ManagedServerDiscoverySummary {
    return servers.reduce<ManagedServerDiscoverySummary>((summary, server) => {
        const status = normalizeDiscoveryStatus(server._meta?.status);
        const metadataSource = server._meta?.metadataSource ?? null;
        const hasStaleReadyCache = hasStaleReadyMetadata(server);

        summary.totalCount += 1;

        if (status === 'ready') {
            summary.readyCount += 1;
            if (hasStaleReadyCache) {
                summary.staleReadyCount += 1;
            }
        } else {
            summary.unresolvedCount += 1;
        }

        if (status !== 'ready' || hasStaleReadyCache) {
            summary.repairableCount += 1;
        }

        if (!server._meta?.lastSuccessfulBinaryLoadAt) {
            summary.neverLoadedCount += 1;
        }

        if (isLocalCompatMetadataSource(metadataSource)) {
            summary.localCompatCount += 1;
        }

        return summary;
    }, {
        totalCount: 0,
        readyCount: 0,
        unresolvedCount: 0,
        staleReadyCount: 0,
        repairableCount: 0,
        neverLoadedCount: 0,
        localCompatCount: 0,
    });
}

export function getBulkMetadataTargetUuids(
    servers: ManagedServerDiscoveryRecord[],
    mode: BulkMetadataTargetMode,
): string[] {
    return servers
        .filter((server) => {
            if (!server.uuid) {
                return false;
            }

            if (mode === 'all') {
                return true;
            }

            return normalizeDiscoveryStatus(server._meta?.status) !== 'ready'
                || hasStaleReadyMetadata(server);
        })
        .map((server) => server.uuid as string);
}

export function buildDashboardServerRecords(
    runtimeServers: RuntimeServerRecord[],
    managedServers: ManagedServerRuntimeRecord[],
): DashboardServerRecord[] {
    const runtimeByName = new Map(runtimeServers.map((server) => [normalizeServerName(server.name), server]));
    const matchedRuntimeNames = new Set<string>();
    const seenManagedNames = new Set<string>();

    const managedFirst = managedServers.reduce<DashboardServerRecord[]>((acc, server) => {
        const normalizedName = normalizeServerName(server.name);
        if (seenManagedNames.has(normalizedName)) {
            return acc;
        }
        seenManagedNames.add(normalizedName);

        const runtime = runtimeByName.get(normalizedName);
        if (runtime) {
            matchedRuntimeNames.add(normalizedName);
        }

        acc.push({
            uuid: server.uuid,
            name: server.name,
            status: runtime?.status ?? 'configured',
            toolCount: runtime?.toolCount ?? 0,
            runtimeState: runtime?.runtimeState ?? runtime?.status ?? 'configured',
            warmupState: runtime?.warmupState ?? (server.always_on ? 'scheduled' : 'idle'),
            runtimeConnected: runtime?.runtimeConnected ?? (runtime?.status === 'connected'),
            advertisedToolCount: runtime?.advertisedToolCount,
            advertisedSource: runtime?.advertisedSource,
            lastConnectedAt: runtime?.lastConnectedAt ?? null,
            lastError: runtime?.lastError ?? null,
            metadataStatus: server._meta?.status ?? undefined,
            metadataSource: server._meta?.metadataSource ?? undefined,
            metadataToolCount: server._meta?.toolCount ?? undefined,
            lastSuccessfulBinaryLoadAt: server._meta?.lastSuccessfulBinaryLoadAt ?? undefined,
            always_on: server.always_on,
            source_published_server_uuid: server.source_published_server_uuid ?? null,
            config: {
                command: runtime?.config?.command ?? server.command ?? undefined,
                args: runtime?.config?.args ?? server.args ?? [],
                env: runtime?.config?.env ?? Object.keys(server.env ?? {}),
            },
        });
        return acc;
    }, []);

    const runtimeOnly = runtimeServers
        .filter((server) => !matchedRuntimeNames.has(normalizeServerName(server.name)))
        .map<DashboardServerRecord>((server) => ({
            name: server.name,
            status: server.status,
            toolCount: server.toolCount,
            runtimeState: server.runtimeState ?? server.status,
            warmupState: server.warmupState ?? 'idle',
            runtimeConnected: server.runtimeConnected ?? (server.status === 'connected'),
            advertisedToolCount: server.advertisedToolCount,
            advertisedSource: server.advertisedSource,
            lastConnectedAt: server.lastConnectedAt ?? null,
            lastError: server.lastError ?? null,
            config: server.config,
        }));

    return [...managedFirst, ...runtimeOnly];
}

export function buildServerToolActionLinks(serverName: string): ServerToolActionLinks {
    const encodedServerName = encodeURIComponent(serverName);

    return {
        inspectToolsHref: `/dashboard/mcp/inspector?server=${encodedServerName}`,
        editToolsHref: `/dashboard/mcp/inspector?server=${encodedServerName}&mode=edit-tools`,
        logsHref: `/dashboard/mcp/logs?server=${encodedServerName}`,
    };
}
