import { mcpServersRepository, toolsRepository } from '../db/repositories/index.js';

import { deriveSemanticCatalogForServer } from './catalogMetadata.js';
import { loadTormentNexusMcpConfig, type TormentNexusMcpServerEntry, type TormentNexusMcpToolMetadata } from './mcpJsonConfig.js';
import { namespaceToolName } from './namespaces.js';
import { formatOptionalSqliteFailure } from '../db/sqliteAvailability.js';

export type CachedMcpServerInventory = Awaited<ReturnType<typeof mcpServersRepository.findAll>>[number] & {
    displayName: string;
    tags: string[];
    alwaysOnAdvertised: boolean;
};

export type CachedMcpToolInventory = {
    name: string;
    description: string;
    server: string;
    serverDisplayName: string;
    serverTags: string[];
    toolTags: string[];
    semanticGroup: string;
    semanticGroupLabel: string;
    advertisedName: string;
    keywords: string[];
    alwaysOn: boolean;
    originalName: string;
    inputSchema: unknown;
};

export type CachedMcpInventorySource = 'database' | 'config' | 'empty';

type CachedMcpInventorySnapshot = {
    servers: CachedMcpServerInventory[];
    toolCounts: Map<string, number>;
    tools: CachedMcpToolInventory[];
    source: CachedMcpInventorySource;
    snapshotUpdatedAt: string | null;
    databaseAvailable: boolean;
    databaseError: string | null;
    fallbackUsed: boolean;
};

function toSnapshotUpdatedAt(values: Array<string | undefined | null>): string | null {
    const normalized = values
        .filter((value): value is string => typeof value === 'string' && value.trim().length > 0)
        .sort((left, right) => right.localeCompare(left));

    return normalized[0] ?? null;
}

function toDateOrNull(value: string | undefined | null): Date | null {
    if (!value) {
        return null;
    }

    const parsed = new Date(value);
    return Number.isNaN(parsed.getTime()) ? null : parsed;
}

function buildConfigSnapshot(configServers: Record<string, TormentNexusMcpServerEntry>): CachedMcpInventorySnapshot {
    const servers = Object.entries(configServers).map(([name, server], index) => {
        const metadata = server._meta;
        return {
            uuid: `config:${name}:${index}`,
            name,
            type: server.type ?? (server.url ? 'SSE' : 'STDIO'),
            command: server.command ?? '',
            args: server.args ?? [],
            env: server.env ?? {},
            bearerToken: null,
            headers: {},
            url: server.url ?? null,
            description: metadata?.description ?? server.description ?? null,
            created_at: toDateOrNull(metadata?.discoveredAt),
            updated_at: toDateOrNull(metadata?.cacheHydratedAt ?? metadata?.discoveredAt),
            error_status: metadata?.status === 'failed' ? 'ERROR' : 'NONE',
            error_message: metadata?.error ?? null,
            enabled: !server.disabled,
            always_on: Boolean(metadata?.alwaysOn),
            user_id: null,
            source_published_server_uuid: null,
            displayName: metadata?.displayName ?? metadata?.serverName ?? name,
            tags: metadata?.serverTags ?? [],
            alwaysOnAdvertised: Boolean(metadata?.alwaysOn),
        } as CachedMcpServerInventory;
    });

    const tools: CachedMcpToolInventory[] = [];
    const toolCounts = new Map<string, number>();
    const updatedAtCandidates: string[] = [];

    servers.forEach((server) => {
        const configEntry = configServers[server.name];
        const metadataTools = Array.isArray(configEntry?._meta?.tools) ? configEntry._meta?.tools ?? [] : [];
        toolCounts.set(server.uuid, metadataTools.length);

        if (configEntry?._meta?.cacheHydratedAt) {
            updatedAtCandidates.push(configEntry._meta.cacheHydratedAt);
        }
        if (configEntry?._meta?.discoveredAt) {
            updatedAtCandidates.push(configEntry._meta.discoveredAt);
        }

        metadataTools.forEach((tool) => {
            const typedTool = tool as TormentNexusMcpToolMetadata;
            tools.push({
                name: namespaceToolName(server.name, typedTool.name),
                description: typedTool.description ?? '',
                server: server.name,
                serverDisplayName: typedTool.serverDisplayName ?? server.displayName ?? server.name,
                serverTags: typedTool.serverTags ?? server.tags ?? [],
                toolTags: typedTool.toolTags ?? [],
                semanticGroup: typedTool.semanticGroup ?? 'general-utility',
                semanticGroupLabel: typedTool.semanticGroupLabel ?? 'general utility',
                advertisedName: typedTool.advertisedName ?? namespaceToolName(server.name, typedTool.name),
                keywords: typedTool.keywords ?? [],
                alwaysOn: Boolean(typedTool.alwaysOn ?? server.alwaysOnAdvertised),
                originalName: typedTool.name,
                inputSchema: typedTool.inputSchema ?? null,
            });
        });
    });

    return {
        servers,
        toolCounts,
        tools,
        source: tools.length > 0 || servers.length > 0 ? 'config' : 'empty',
        snapshotUpdatedAt: toSnapshotUpdatedAt(updatedAtCandidates),
        databaseAvailable: true,
        databaseError: null,
        fallbackUsed: false,
    };
}

async function buildDatabaseSnapshot(): Promise<CachedMcpInventorySnapshot> {
    const [servers, tools] = await Promise.all([
        mcpServersRepository.findAll(),
        toolsRepository.findAll(),
    ]);

    const serverNames = new Map(servers.map((server) => [server.uuid, server.name]));
    const toolsByServerUuid = new Map<string, typeof tools>();
    const toolCounts = new Map<string, number>();

    for (const tool of tools) {
        toolCounts.set(tool.mcp_server_uuid, (toolCounts.get(tool.mcp_server_uuid) ?? 0) + 1);
        const bucket = toolsByServerUuid.get(tool.mcp_server_uuid) ?? [];
        bucket.push(tool);
        toolsByServerUuid.set(tool.mcp_server_uuid, bucket);
    }

    const derivedByServerUuid = new Map(servers.map((server) => [
        server.uuid,
        deriveSemanticCatalogForServer({
            serverName: server.name,
            description: server.description ?? null,
            alwaysOn: server.always_on ?? false,
            tools: (toolsByServerUuid.get(server.uuid) ?? []).map((tool) => ({
                name: tool.name,
                title: tool.title ?? null,
                description: tool.description ?? null,
                inputSchema: tool.toolSchema ?? null,
                alwaysOn: tool.always_on ?? false,
            })),
        }),
    ]));

    return {
        servers: servers.map((server) => {
            const derived = derivedByServerUuid.get(server.uuid);
            return {
                ...server,
                displayName: derived?.serverDisplayName ?? server.name,
                tags: derived?.serverTags ?? [],
                alwaysOnAdvertised: Boolean(server.always_on),
            };
        }),
        toolCounts,
        tools: tools.map((tool) => {
            const serverName = serverNames.get(tool.mcp_server_uuid) ?? 'unknown';
            const derivedServer = derivedByServerUuid.get(tool.mcp_server_uuid);
            const derivedTool = derivedServer?.tools.find((candidate) => candidate.name === tool.name);

            return {
                name: namespaceToolName(serverName, tool.name),
                description: tool.description ?? '',
                server: serverName,
                serverDisplayName: derivedTool?.serverDisplayName ?? serverName,
                serverTags: derivedTool?.serverTags ?? [],
                toolTags: derivedTool?.toolTags ?? [],
                semanticGroup: derivedTool?.semanticGroup ?? 'general-utility',
                semanticGroupLabel: derivedTool?.semanticGroupLabel ?? 'general utility',
                advertisedName: derivedTool?.advertisedName ?? namespaceToolName(serverName, tool.name),
                keywords: derivedTool?.keywords ?? [],
                alwaysOn: Boolean(tool.always_on ?? false) || Boolean(derivedServer?.alwaysOn),
                originalName: tool.name,
                inputSchema: tool.toolSchema ?? null,
            };
        }),
        source: tools.length > 0 || servers.length > 0 ? 'database' : 'empty',
        snapshotUpdatedAt: null,
        databaseAvailable: true,
        databaseError: null,
        fallbackUsed: false,
    };
}

export async function getCachedToolInventory() {
    const config = await loadTormentNexusMcpConfig().catch(() => ({ mcpServers: {} }));
    const configSnapshot = buildConfigSnapshot(config.mcpServers ?? {});

    try {
        const databaseSnapshot = await buildDatabaseSnapshot();
        
        // MERGE both sources so they don't cross-contaminate but both are available
        const mergedServers = [...configSnapshot.servers, ...databaseSnapshot.servers];
        const mergedTools = [...configSnapshot.tools, ...databaseSnapshot.tools];
        
        // Build toolCounts keyed by SERVER NAME (not UUID) for efficient lookup in MCPServer.ts
        const mergedToolCounts = new Map<string, number>();
        
        // Add config snapshot counts (keyed by server name)
        for (const [key, val] of configSnapshot.toolCounts.entries()) {
            // Config uses "config:name" format, extract actual name
            const nameKey = key.startsWith('config:') ? key.replace('config:', '') : key;
            mergedToolCounts.set(nameKey, val);
        }
        // Add database snapshot counts (keyed by server name)
        for (const server of databaseSnapshot.servers) {
            const count = databaseSnapshot.toolCounts.get(server.uuid) ?? 0;
            mergedToolCounts.set(server.name, count);
        }

        const source: CachedMcpInventorySource = mergedTools.length > 0 ? 'database' : 'empty';

        return {
            servers: mergedServers,
            toolCounts: mergedToolCounts,
            tools: mergedTools,
            source,
            snapshotUpdatedAt: databaseSnapshot.snapshotUpdatedAt ?? configSnapshot.snapshotUpdatedAt,
            databaseAvailable: true,
            databaseError: null,
            fallbackUsed: false,
        };
    } catch (error) {
        return {
            ...configSnapshot,
            databaseAvailable: false,
            databaseError: formatOptionalSqliteFailure('Persisted MCP inventory is unavailable', error),
            fallbackUsed: true,
        };
    }
}
