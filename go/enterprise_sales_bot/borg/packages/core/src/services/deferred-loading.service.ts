import crypto from "crypto";
import { Tool } from "@modelcontextprotocol/sdk/types.js";

export interface DeferredTool {
    name: string;
    summary: string;
    serverUuid: string;
    hasSchema: boolean;
}

export interface ToolManifest {
    tools: DeferredTool[];
    totalCount: number;
    lastUpdated: number;
}

const SCHEMA_CACHE_TTL = 5 * 60 * 1000; // 5 minutes
const MANIFEST_CACHE_TTL = 60 * 1000; // 1 minute

interface CachedSchema {
    schema: Tool;
    timestamp: number;
}

interface CachedManifest {
    manifest: ToolManifest;
    timestamp: number;
}

export class DeferredLoadingService {
    private schemaCache: Map<string, CachedSchema> = new Map();
    private manifestCache: Map<string, CachedManifest> = new Map();

    createDeferredTool(
        tool: Tool,
        serverUuid: string,
        maxSummaryLength: number = 100,
    ): DeferredTool {
        const summary = tool.description
            ? tool.description.length > maxSummaryLength
                ? tool.description.substring(0, maxSummaryLength) + "..."
                : tool.description
            : "No description";

        return {
            name: tool.name,
            summary,
            serverUuid,
            hasSchema: !!tool.inputSchema,
        };
    }

    createMinimalTool(deferred: DeferredTool): Tool {
        return {
            name: deferred.name,
            description: `[Deferred] ${deferred.summary}`,
            inputSchema: {
                type: "object",
                properties: {},
            },
        };
    }

    cacheToolSchema(toolName: string, schema: Tool): void {
        this.schemaCache.set(toolName, {
            schema,
            timestamp: Date.now(),
        });
    }

    getCachedSchema(toolName: string): Tool | null {
        const cached = this.schemaCache.get(toolName);
        if (!cached) return null;

        if (Date.now() - cached.timestamp > SCHEMA_CACHE_TTL) {
            this.schemaCache.delete(toolName);
            return null;
        }

        return cached.schema;
    }

    cacheManifest(namespaceUuid: string, manifest: ToolManifest): void {
        this.manifestCache.set(namespaceUuid, {
            manifest,
            timestamp: Date.now(),
        });
    }

    getCachedManifest(namespaceUuid: string): ToolManifest | null {
        const cached = this.manifestCache.get(namespaceUuid);
        if (!cached) return null;

        if (Date.now() - cached.timestamp > MANIFEST_CACHE_TTL) {
            this.manifestCache.delete(namespaceUuid);
            return null;
        }

        return cached.manifest;
    }

    invalidateManifestCache(namespaceUuid: string): void {
        this.manifestCache.delete(namespaceUuid);
    }

    invalidateSchemaCache(toolName?: string): void {
        if (toolName) {
            this.schemaCache.delete(toolName);
        } else {
            this.schemaCache.clear();
        }
    }

    getSchemaStats(): { cached: number; oldestMs: number | null } {
        const now = Date.now();
        let oldest: number | null = null;

        for (const [, value] of this.schemaCache) {
            const age = now - value.timestamp;
            if (oldest === null || age > oldest) {
                oldest = age;
            }
        }

        return {
            cached: this.schemaCache.size,
            oldestMs: oldest,
        };
    }

    calculateTokenSavings(
        fullToolCount: number,
        avgSchemaTokens: number = 150,
    ): { before: number; after: number; savingsPercent: number } {
        const metaToolTokens = 8 * 100; // 8 meta-tools * ~100 tokens each
        const summaryTokens = fullToolCount * 20; // ~20 tokens per summary

        const before = metaToolTokens + fullToolCount * avgSchemaTokens;
        const after = metaToolTokens + summaryTokens;
        const savingsPercent = Math.round(((before - after) / before) * 100);

        return { before, after, savingsPercent };
    }
}

export const deferredLoadingService = new DeferredLoadingService();
