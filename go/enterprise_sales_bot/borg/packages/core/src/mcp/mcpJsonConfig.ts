import fs from 'node:fs/promises';
import path from 'node:path';

import type { SavedScriptConfig, SavedToolSetConfig } from '../interfaces/IConfigProvider.js';

export type TormentNexusMcpToolMetadata = {
    name: string;
    title?: string | null;
    description?: string | null;
    advertisedName?: string | null;
    serverDisplayName?: string | null;
    serverTags?: string[];
    toolTags?: string[];
    semanticGroup?: string | null;
    semanticGroupLabel?: string | null;
    keywords?: string[];
    alwaysOn?: boolean;
    inputSchema?: Record<string, unknown> | null;
    outputSchema?: Record<string, unknown> | null;
    annotations?: Record<string, unknown> | null;
    raw?: Record<string, unknown> | null;
};

export type TormentNexusMcpServerDiscoveryMetadata = {
    status: 'ready' | 'failed' | 'unsupported' | 'pending';
    metadataVersion?: number;
    metadataSource?: 'binary' | 'cache' | 'derived';
    discoveredAt?: string;
    lastAttemptedBinaryLoadAt?: string;
    lastSuccessfulBinaryLoadAt?: string;
    cacheHydratedAt?: string;
    configFingerprint?: string;
    transportType?: 'STDIO' | 'SSE' | 'STREAMABLE_HTTP';
    serverName?: string;
    displayName?: string;
    description?: string | null;
    serverTags?: string[];
    alwaysOn?: boolean;
    command?: string | null;
    args?: string[];
    envKeys?: string[];
    url?: string | null;
    headerKeys?: string[];
    reloadableFromCache?: boolean;
    toolCount: number;
    tools: TormentNexusMcpToolMetadata[];
    error?: string;
};

export type TormentNexusMcpServerEntry = {
    command?: string;
    args?: string[];
    env?: Record<string, string>;
    url?: string;
    disabled?: boolean;
    description?: string | null;
    type?: 'STDIO' | 'SSE' | 'STREAMABLE_HTTP';
    _meta?: TormentNexusMcpServerDiscoveryMetadata;
};

export type TormentNexusMcpJsonConfig = {
    mcpServers: Record<string, TormentNexusMcpServerEntry>;
    alwaysVisibleTools?: string[];
    scripts?: SavedScriptConfig[];
    toolSets?: SavedToolSetConfig[];
    settings?: Record<string, any>;
    [key: string]: unknown;
};

import os from 'node:os';

const JSONC_HEADER = `// TormentNexus MCP configuration\n// This file is TormentNexus-owned and may include cached server metadata under mcpServers.<name>._meta.\n`;

export function getTormentNexusConfigDir(): string {
    // If there is an mcp.jsonc in the current working directory, use it
    // This allows project-level config to be the source of truth if intended.
    if (process.env.TORMENTNEXUS_CONFIG_DIR) {
        return process.env.TORMENTNEXUS_CONFIG_DIR;
    }
    const cwdPath = process.cwd();
    try {
        const cwdStat = require('node:fs').statSync(path.join(cwdPath, 'mcp.jsonc'));
        if (cwdStat.isFile()) {
            return cwdPath;
        }
    } catch {
        // Fall back
    }
    return path.join(os.homedir(), '.tormentnexus');
}

export function getTormentNexusMcpJsoncPath(configDir: string = getTormentNexusConfigDir()): string {
    return path.join(configDir, 'mcp.jsonc');
}

export function getTormentNexusMcpJsonPath(configDir: string = getTormentNexusConfigDir()): string {
    return path.join(configDir, 'mcp.json');
}

export function getTormentNexusToolCachePath(configDir: string = getTormentNexusConfigDir()): string {
    return path.join(configDir, 'mcp-cache.json');
}

export function stripJsonComments(content: string): string {
    let result = '';
    let inString = false;
    let stringDelimiter = '"';
    let isEscaped = false;
    let inLineComment = false;
    let inBlockComment = false;

    for (let index = 0; index < content.length; index += 1) {
        const current = content[index];
        const next = content[index + 1];

        if (inLineComment) {
            if (current === '\n' || current === '\r') {
                inLineComment = false;
                result += current;
            }
            continue;
        }

        if (inBlockComment) {
            if (current === '*' && next === '/') {
                inBlockComment = false;
                index += 1;
                continue;
            }

            if (current === '\n' || current === '\r') {
                result += current;
            }
            continue;
        }

        if (inString) {
            result += current;

            if (isEscaped) {
                isEscaped = false;
                continue;
            }

            if (current === '\\') {
                isEscaped = true;
                continue;
            }

            if (current === stringDelimiter) {
                inString = false;
            }

            continue;
        }

        if (current === '"' || current === "'") {
            inString = true;
            stringDelimiter = current;
            result += current;
            continue;
        }

        if (current === '/' && next === '/') {
            inLineComment = true;
            index += 1;
            continue;
        }

        if (current === '/' && next === '*') {
            inBlockComment = true;
            index += 1;
            continue;
        }

        result += current;
    }

    return result;
}

function normalizeConfigShape(config: unknown): TormentNexusMcpJsonConfig {
    if (!config || typeof config !== 'object') {
        return { mcpServers: {} };
    }

    const candidate = config as TormentNexusMcpJsonConfig;
    return {
        ...candidate,
        mcpServers: candidate.mcpServers && typeof candidate.mcpServers === 'object'
            ? candidate.mcpServers
            : {},
        alwaysVisibleTools: Array.isArray(candidate.alwaysVisibleTools)
            ? candidate.alwaysVisibleTools.filter((toolName): toolName is string => typeof toolName === 'string')
            : undefined,
        scripts: Array.isArray(candidate.scripts)
            ? candidate.scripts as SavedScriptConfig[]
            : undefined,
        toolSets: Array.isArray(candidate.toolSets)
            ? candidate.toolSets as SavedToolSetConfig[]
            : undefined,
        settings: candidate.settings && typeof candidate.settings === 'object'
            ? candidate.settings
            : undefined,
    };
}

function toCompatibilityConfig(config: TormentNexusMcpJsonConfig): Record<string, unknown> {
    const compatibilityServers = Object.fromEntries(
        Object.entries(config.mcpServers || {}).map(([name, server]) => {
            const { _meta: _ignoredMeta, ...serverWithoutMeta } = server;
            return [name, serverWithoutMeta];
        }),
    );

    const compatibilityConfig: Record<string, unknown> = {
        ...config,
        mcpServers: compatibilityServers,
    };

    delete compatibilityConfig.settings;

    return compatibilityConfig;
}

export async function loadTormentNexusMcpConfig(configDir?: string): Promise<TormentNexusMcpJsonConfig> {
    const jsoncPath = getTormentNexusMcpJsoncPath(configDir);

    try {
        const raw = await fs.readFile(jsoncPath, 'utf-8');
        return normalizeConfigShape(JSON.parse(stripJsonComments(raw)));
    } catch (error) {
        const errorCode = (error as NodeJS.ErrnoException).code;
        if (errorCode !== 'ENOENT') {
            throw error;
        }
    }

    return { mcpServers: {} };
}

export async function writeTormentNexusMcpConfig(config: TormentNexusMcpJsonConfig, configDir?: string): Promise<void> {
    const normalized = normalizeConfigShape(config);
    const jsoncPath = getTormentNexusMcpJsoncPath(configDir);
    const jsonPath = getTormentNexusMcpJsonPath(configDir);

    await fs.mkdir(path.dirname(jsoncPath), { recursive: true });

    const jsoncBody = `${JSONC_HEADER}${JSON.stringify(normalized, null, 2)}\n`;
    const jsonBody = `${JSON.stringify(toCompatibilityConfig(normalized), null, 2)}\n`;
    await fs.writeFile(jsoncPath, jsoncBody, 'utf-8');
    await fs.writeFile(jsonPath, jsonBody, 'utf-8');
}

export async function writeToolCache(config: TormentNexusMcpJsonConfig, configDir?: string): Promise<void> {
    const cachePath = getTormentNexusToolCachePath(configDir);
    await fs.mkdir(path.dirname(cachePath), { recursive: true });
    await fs.writeFile(cachePath, JSON.stringify(config, null, 2), 'utf-8');
}

export async function loadToolCache(configDir?: string): Promise<TormentNexusMcpJsonConfig | null> {
    try {
        const raw = await fs.readFile(getTormentNexusToolCachePath(configDir), 'utf-8');
        return normalizeConfigShape(JSON.parse(raw));
    } catch {
        return null;
    }
}


// Aliases removed — original functions already use TormentNexus naming
