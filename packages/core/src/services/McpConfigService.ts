
import * as fs from 'fs/promises';
import { mcpServersRepository } from '../db/repositories/mcp-servers.repo.js';
import { toolsRepository } from '../db/repositories/tools.repo.js';
import { formatOptionalSqliteFailure, isSqliteUnavailableError, sqliteErrorMessage } from '../db/sqliteAvailability.js';
import { loadTormentNexusMcpConfig } from '../mcp/mcpJsonConfig.js';

export class McpConfigService {
    private syncState: {
        inProgress: boolean;
        lastStartedAt?: number;
        lastCompletedAt?: number;
        lastSuccessAt?: number;
        lastError?: string;
        lastServerCount: number;
        lastToolCount: number;
    } = {
        inProgress: false,
        lastServerCount: 0,
        lastToolCount: 0,
    };

    /**
     * Reads TormentNexus's mcp.jsonc (falling back to mcp.json) and updates the database to match.
     * This makes TormentNexus's config file the authoritative source for config entry existence/content.
     */
    async syncWithDatabase() {
        console.log('[McpConfigService] Syncing Database with mcp.jsonc...');
        this.syncState.inProgress = true;
        this.syncState.lastStartedAt = Date.now();
        this.syncState.lastError = undefined;

        try {
            try {
                await fs.access(process.cwd());
            } catch (e: any) {
                if (e.code === 'ENOENT') {
                    console.log('[McpConfigService] Workspace not available. Skipping sync.');
                    this.syncState.lastCompletedAt = Date.now();
                    this.syncState.inProgress = false;
                    return;
                }
                throw e;
            }

            const config = await loadTormentNexusMcpConfig();
            const servers = config.mcpServers || {};
            this.syncState.lastServerCount = Object.keys(servers).length;

            for (const [name, serverConfig] of Object.entries(servers)) {
                // Determine type
                let type: any = 'STDIO';
                if ((serverConfig as any).url) {
                    type = 'SSE'; // Simplified assumption, could be STREAMABLE_HTTP
                }

                // Check if exists
                const existing = await mcpServersRepository.findByName(name);

                if (existing) {
                    // Update
                    const updatedServer = await mcpServersRepository.update({
                        uuid: existing.uuid,
                        name: name, // shouldn't change
                        command: (serverConfig as any).command,
                        args: (serverConfig as any).args,
                        env: (serverConfig as any).env,
                        url: (serverConfig as any).url,
                        description: (serverConfig as any).description,
                        // Preserve other fields
                    }, { skipSync: true, skipDiscovery: true });
                    await this.syncStoredMetadataTools(updatedServer.uuid, serverConfig);
                    console.log(`[McpConfigService] Updated server: ${name}`);
                } else {
                    // Create
                    const createdServer = await mcpServersRepository.create({
                        name: name,
                        type: type,
                        command: (serverConfig as any).command,
                        args: (serverConfig as any).args,
                        env: (serverConfig as any).env,
                        url: (serverConfig as any).url,
                        description: (serverConfig as any).description,
                    }, { skipSync: true, skipDiscovery: true });
                    await this.syncStoredMetadataTools(createdServer.uuid, serverConfig);
                    console.log(`[McpConfigService] Created server: ${name}`);
                }
            }

            this.syncState.lastToolCount = (await toolsRepository.findAll()).length;
            this.syncState.lastCompletedAt = Date.now();
            this.syncState.lastSuccessAt = this.syncState.lastCompletedAt;
            console.log('[McpConfigService] Sync complete.');

        } catch (error) {
            this.syncState.lastCompletedAt = Date.now();
            this.syncState.lastError = isSqliteUnavailableError(error)
                ? 'SQLite runtime is unavailable for this run.'
                : sqliteErrorMessage(error);
            const message = formatOptionalSqliteFailure('[McpConfigService] Sync failed', error);
            if (isSqliteUnavailableError(error)) {
                console.warn(message);
            } else {
                console.error('[McpConfigService] Sync failed:', error);
            }
        } finally {
            this.syncState.inProgress = false;
        }
    }

    getStatus() {
        return { ...this.syncState };
    }

    async addServerConfig(name: string, config: any) {
        console.log(`[McpConfigService] Adding discovered server: ${name}`);
        // Create in DB
        let type: any = 'STDIO';
        if (config.url) {
            type = 'SSE';
        }

        const createdServer = await mcpServersRepository.create({
            name: name,
            type: type,
            command: config.command,
            args: config.args,
            env: config.env,
            url: config.url,
            description: config.description,
        }, { skipSync: true, skipDiscovery: true });

        // Trigger a sync back to mcp.jsonc if desired, 
        // or just let it exist in DB until the next full sync.
        // For now, we'll keep it in DB.
        return createdServer;
    }

    private async syncStoredMetadataTools(serverUuid: string, serverConfig: unknown): Promise<void> {
        // Do not overwrite database tools if the config doesn't have tools populated.
        // This prevents wiping out tools (and their always_on status) that were discovered by the daemon.
        const metadataTools = Array.isArray((serverConfig as { _meta?: { tools?: unknown[] } })?._meta?.tools)
            ? (serverConfig as { _meta?: { tools?: Array<Record<string, unknown>> } })._meta?.tools ?? []
            : [];

        if (metadataTools.length === 0) {
            return; // Skip syncing tools if mcp.jsonc doesn't have any cached tools for this server
        }

        await toolsRepository.syncTools({
            mcpServerUuid: serverUuid,
            tools: metadataTools
                .filter((tool): tool is Record<string, unknown> => Boolean(tool && typeof tool === 'object' && typeof tool.name === 'string'))
                .map((tool) => ({
                    name: String(tool.name),
                    description: typeof tool.description === 'string' ? tool.description : undefined,
                    inputSchema: tool.inputSchema && typeof tool.inputSchema === 'object'
                        ? {
                            properties: (tool.inputSchema as { properties?: Record<string, unknown> }).properties,
                            required: Array.isArray((tool.inputSchema as { required?: unknown[] }).required)
                                ? (tool.inputSchema as { required: unknown[] }).required.filter((value): value is string => typeof value === 'string')
                                : undefined,
                        }
                        : undefined,
                })),
        });
    }
}
