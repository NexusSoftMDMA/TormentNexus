import { createHash } from 'node:crypto';

import type { DatabaseMcpServer } from '../types/mcp-admin/index.js';
import { deriveSemanticCatalogForServer } from './catalogMetadata.js';
import type { TormentNexusMcpServerDiscoveryMetadata, TormentNexusMcpToolMetadata } from './mcpJsonConfig.js';

export type MetadataReloadStrategy = 'auto' | 'binary' | 'cache' | 'skip';

type JsonLike = null | boolean | number | string | JsonLike[] | { [key: string]: JsonLike };

function normalizeJsonLike(value: unknown, depth: number = 0): JsonLike | undefined {
    if (depth > 8) {
        return '[MaxDepthExceeded]';
    }

    if (
        value === null ||
        typeof value === 'string' ||
        typeof value === 'number' ||
        typeof value === 'boolean'
    ) {
        // @ts-expect-error -- type mismatch from inferred schema types
        return value;
    }

    if (typeof value === 'bigint') {
        return value.toString();
    }

    if (value instanceof Date) {
        return value.toISOString();
    }

    if (Array.isArray(value)) {
        return value
            .map((entry) => normalizeJsonLike(entry, depth + 1))
            .filter((entry): entry is JsonLike => entry !== undefined);
    }

    if (value && typeof value === 'object') {
        const normalizedEntries = Object.entries(value)
            .map(([key, entry]) => [key, normalizeJsonLike(entry, depth + 1)] as const)
            .filter(([, entry]) => entry !== undefined)
            .sort(([left], [right]) => left.localeCompare(right));

        return Object.fromEntries(normalizedEntries) as JsonLike;
    }

    return undefined;
}

function toSortedJson(value: unknown): string {
    return JSON.stringify(normalizeJsonLike(value) ?? null);
}

function toRecord(value: unknown): Record<string, unknown> | null {
    const normalized = normalizeJsonLike(value);
    if (!normalized || Array.isArray(normalized) || typeof normalized !== 'object') {
        return null;
    }

    return normalized as Record<string, unknown>;
}

export function buildServerConfigFingerprint(server: Pick<DatabaseMcpServer, 'name' | 'type' | 'command' | 'args' | 'env' | 'url' | 'headers' | 'bearerToken'>): string {
    return createHash('sha256')
        .update(toSortedJson({
            name: server.name,
            type: server.type,
            command: server.command ?? null,
            args: server.args ?? [],
            env: server.env ?? {},
            url: server.url ?? null,
            headers: server.headers ?? {},
            bearerToken: server.bearerToken ?? null,
        }))
        .digest('hex');
}

export function normalizeDiscoveredToolMetadata(rawTool: unknown): TormentNexusMcpToolMetadata | null {
    const rawRecord = toRecord(rawTool);
    const name = typeof rawRecord?.name === 'string' ? rawRecord.name.trim() : '';
    if (!name) {
        return null;
    }

    return {
        name,
        title: typeof rawRecord?.title === 'string' ? rawRecord.title : null,
        description: typeof rawRecord?.description === 'string' ? rawRecord.description : null,
        inputSchema: toRecord(rawRecord?.inputSchema),
        outputSchema: toRecord(rawRecord?.outputSchema),
        annotations: toRecord(rawRecord?.annotations),
        raw: rawRecord,
    };
}

export function buildBaseServerMetadata(server: Pick<DatabaseMcpServer, 'name' | 'type' | 'command' | 'args' | 'env' | 'url' | 'headers' | 'bearerToken'>): Omit<TormentNexusMcpServerDiscoveryMetadata, 'status' | 'toolCount' | 'tools'> {
    return {
        metadataVersion: 2,
        configFingerprint: buildServerConfigFingerprint(server),
        transportType: server.type,
        serverName: server.name,
        displayName: server.name,
        command: server.command ?? null,
        args: server.args ?? [],
        envKeys: Object.keys(server.env ?? {}).sort((left, right) => left.localeCompare(right)),
        url: server.url ?? null,
        headerKeys: Object.keys(server.headers ?? {}).sort((left, right) => left.localeCompare(right)),
    };
}

function enrichServerMetadata(
    server: Pick<DatabaseMcpServer, 'name' | 'description' | 'always_on'>,
    metadata: TormentNexusMcpServerDiscoveryMetadata,
): TormentNexusMcpServerDiscoveryMetadata {
    const derived = deriveSemanticCatalogForServer({
        serverName: server.name,
        description: server.description ?? null,
        alwaysOn: server.always_on ?? false,
        tools: metadata.tools.map((tool) => ({
            name: tool.name,
            title: tool.title ?? null,
            description: tool.description ?? null,
            inputSchema: tool.inputSchema ?? null,
            alwaysOn: tool.alwaysOn ?? false,
        })),
    });

    const derivedTools = new Map(derived.tools.map((tool) => [tool.name, tool]));

    return {
        ...metadata,
        displayName: derived.serverDisplayName,
        description: server.description ?? null,
        serverTags: derived.serverTags,
        alwaysOn: derived.alwaysOn,
        tools: metadata.tools.map((tool) => {
            const derivedTool = derivedTools.get(tool.name);
            if (!derivedTool) {
                return tool;
            }

            return {
                ...tool,
                advertisedName: derivedTool.advertisedName,
                serverDisplayName: derivedTool.serverDisplayName,
                serverTags: derivedTool.serverTags,
                toolTags: derivedTool.toolTags,
                semanticGroup: derivedTool.semanticGroup,
                semanticGroupLabel: derivedTool.semanticGroupLabel,
                keywords: derivedTool.keywords,
                alwaysOn: derivedTool.alwaysOn,
            };
        }),
    };
}

export function buildBinaryDiscoveryMetadata(
    server: Pick<DatabaseMcpServer, 'name' | 'description' | 'always_on' | 'type' | 'command' | 'args' | 'env' | 'url' | 'headers' | 'bearerToken'>,
    rawTools: unknown[],
    discoveredAt: string,
): TormentNexusMcpServerDiscoveryMetadata {
    const tools = rawTools
        .map((rawTool) => normalizeDiscoveredToolMetadata(rawTool))
        .filter((tool): tool is TormentNexusMcpToolMetadata => Boolean(tool));

    return enrichServerMetadata(server, {
        ...buildBaseServerMetadata(server),
        status: 'ready',
        metadataSource: 'binary',
        discoveredAt,
        lastAttemptedBinaryLoadAt: discoveredAt,
        lastSuccessfulBinaryLoadAt: discoveredAt,
        reloadableFromCache: true,
        toolCount: tools.length,
        tools,
    });
}

export function buildFailureDiscoveryMetadata(
    server: Pick<DatabaseMcpServer, 'name' | 'description' | 'always_on' | 'type' | 'command' | 'args' | 'env' | 'url' | 'headers' | 'bearerToken'>,
    status: TormentNexusMcpServerDiscoveryMetadata['status'],
    discoveredAt: string,
    error: string,
): TormentNexusMcpServerDiscoveryMetadata {
    return enrichServerMetadata(server, {
        ...buildBaseServerMetadata(server),
        status,
        metadataSource: status === 'unsupported' ? 'derived' : 'binary',
        discoveredAt,
        lastAttemptedBinaryLoadAt: discoveredAt,
        reloadableFromCache: false,
        toolCount: 0,
        tools: [],
        error,
    });
}

export function hasReusableMetadataCache(
    metadata: TormentNexusMcpServerDiscoveryMetadata | null | undefined,
    server: Pick<DatabaseMcpServer, 'name' | 'type' | 'command' | 'args' | 'env' | 'url' | 'headers' | 'bearerToken'>,
): boolean {
    if (!metadata || metadata.status !== 'ready') {
        return false;
    }

    if (!Array.isArray(metadata.tools)) {
        return false;
    }

    if (metadata.tools.length === 0) {
        return false;
    }

    const fingerprint = buildServerConfigFingerprint(server);
    if (metadata.configFingerprint && metadata.configFingerprint !== fingerprint) {
        return false;
    }

    return Boolean(metadata.lastSuccessfulBinaryLoadAt || metadata.discoveredAt);
}

export function hydrateMetadataFromCache(
    metadata: TormentNexusMcpServerDiscoveryMetadata,
    server: Pick<DatabaseMcpServer, 'name' | 'description' | 'always_on' | 'type' | 'command' | 'args' | 'env' | 'url' | 'headers' | 'bearerToken'>,
    hydratedAt: string,
): TormentNexusMcpServerDiscoveryMetadata {
    return enrichServerMetadata(server, {
        ...metadata,
        ...buildBaseServerMetadata(server),
        metadataSource: 'cache',
        cacheHydratedAt: hydratedAt,
        reloadableFromCache: Array.isArray(metadata.tools),
        toolCount: Array.isArray(metadata.tools) ? metadata.tools.length : 0,
        tools: Array.isArray(metadata.tools) ? metadata.tools : [],
    });
}