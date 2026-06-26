import fs from 'node:fs/promises';
import path from 'node:path';
import crypto from 'node:crypto';
import { IConfigProvider, McpServerConfig, SavedScriptConfig, SavedToolSetConfig } from '../../interfaces/IConfigProvider.js';
import { TormentNexusMcpJsonConfig, getTormentNexusConfigDir, loadTormentNexusMcpConfig, writeTormentNexusMcpConfig } from '../../mcp/mcpJsonConfig.js';
import { buildToolPreferenceSettings, readToolPreferencesFromSettings } from '../../routers/mcp-tool-preferences.js';

export class JsonConfigProvider implements IConfigProvider {
    private configDir: string;
    private config: TormentNexusMcpJsonConfig = { mcpServers: {} };

    constructor(configDir: string = getTormentNexusConfigDir()) {
        this.configDir = configDir;
    }

    async init(): Promise<void> {
        try {
            this.config = await loadTormentNexusMcpConfig(this.configDir);
        } catch (error: any) {
            if (error.code === 'ENOENT') {
                // File doesn't exist, create default
                await this.saveConfig();
            } else {
                console.error(`Failed to load mcp.jsonc from ${this.configDir}:`, error);
                throw error;
            }
        }
    }

    async loadMcpServers(): Promise<McpServerConfig[]> {
        await this.init(); // Reload on every access to catch manual edits
        const servers: McpServerConfig[] = [];

        for (const [name, config] of Object.entries(this.config.mcpServers || {})) {
            // Basic validation/transformation
            if (config.command) {
                servers.push({
                    name,
                    type: 'stdio',
                    command: config.command,
                    args: config.args,
                    env: config.env,
                    disabled: config.disabled
                });
            } else if (config.url) {
                servers.push({
                    name,
                    type: 'sse',
                    url: config.url,
                    disabled: config.disabled
                });
            }
        }

        return servers;
    }

    async saveMcpServers(servers: McpServerConfig[]): Promise<void> {
        await this.init();
        const newMcpServers: Record<string, any> = {};
        const existingServers = this.config.mcpServers || {};

        for (const server of servers) {
            if (server.type === 'stdio') {
                newMcpServers[server.name] = {
                    ...(existingServers[server.name] || {}),
                    command: server.command,
                    args: server.args,
                    env: server.env,
                    disabled: server.disabled
                };
            } else if (server.type === 'sse') {
                newMcpServers[server.name] = {
                    ...(existingServers[server.name] || {}),
                    url: server.url,
                    disabled: server.disabled
                };
            }
        }

        this.config.mcpServers = newMcpServers;
        await this.saveConfig();
    }

    async getSettings(): Promise<Record<string, any>> {
        await this.init();
        return this.config;
    }

    async loadAlwaysVisibleTools(): Promise<string[]> {
        await this.init();
        const fromSettings = readToolPreferencesFromSettings(
            this.config.settings?.toolSelection as { importantTools?: unknown; alwaysLoadedTools?: unknown } | undefined,
        ).alwaysLoadedTools;

        if (fromSettings.length > 0) {
            return fromSettings;
        }

        return Array.isArray(this.config.alwaysVisibleTools)
            ? [...this.config.alwaysVisibleTools]
            : [];
    }

    async saveAlwaysVisibleTools(toolNames: string[]): Promise<string[]> {
        await this.init();
        const normalized = Array.from(
            new Set(
                toolNames
                    .filter((toolName): toolName is string => typeof toolName === 'string')
                    .map((toolName) => toolName.trim())
                    .filter((toolName) => toolName.length > 0),
            ),
        );

        this.config.settings = buildToolPreferenceSettings(
            this.config.settings && typeof this.config.settings === 'object'
                ? this.config.settings as Record<string, unknown>
                : {},
            {
                ...readToolPreferencesFromSettings(
                    this.config.settings?.toolSelection as { importantTools?: unknown; alwaysLoadedTools?: unknown } | undefined,
                ),
                alwaysLoadedTools: normalized,
            },
        );
        this.config.alwaysVisibleTools = normalized;
        await this.saveConfig();
        return normalized;
    }

    async loadScripts(): Promise<SavedScriptConfig[]> {
        await this.init();
        return this.config.scripts || [];
    }

    async saveScript(script: SavedScriptConfig): Promise<SavedScriptConfig> {
        await this.init();
        if (!this.config.scripts) {
            this.config.scripts = [];
        }

        // Generate UUID if missing
        if (!script.uuid) {
            script.uuid = crypto.randomUUID();
        }

        const existingIndex = this.config.scripts.findIndex((s: SavedScriptConfig) =>
            s.name === script.name || (s.uuid && s.uuid === script.uuid)
        );

        if (existingIndex >= 0) {
            this.config.scripts[existingIndex] = { ...this.config.scripts[existingIndex], ...script };
        } else {
            this.config.scripts.push(script);
        }

        await this.saveConfig();
        return existingIndex >= 0
            ? this.config.scripts[existingIndex]
            : this.config.scripts[this.config.scripts.length - 1];
    }

    async deleteScript(nameOrUuid: string): Promise<void> {
        await this.init();
        if (!this.config.scripts) return;

        this.config.scripts = this.config.scripts.filter(s => s.name !== nameOrUuid && s.uuid !== nameOrUuid);
        await this.saveConfig();
    }

    async loadToolSets(): Promise<SavedToolSetConfig[]> {
        await this.init();
        return this.config.toolSets || [];
    }

    async saveToolSet(toolSet: SavedToolSetConfig): Promise<SavedToolSetConfig> {
        await this.init();
        if (!this.config.toolSets) {
            this.config.toolSets = [];
        }

        if (!toolSet.uuid) {
            toolSet.uuid = crypto.randomUUID();
        }

        const normalizedToolSet: SavedToolSetConfig = {
            ...toolSet,
            tools: Array.from(new Set(toolSet.tools)),
        };

        const existingIndex = this.config.toolSets.findIndex((set: SavedToolSetConfig) =>
            set.name === normalizedToolSet.name || (set.uuid && set.uuid === normalizedToolSet.uuid)
        );

        if (existingIndex >= 0) {
            this.config.toolSets[existingIndex] = { ...this.config.toolSets[existingIndex], ...normalizedToolSet };
        } else {
            this.config.toolSets.push(normalizedToolSet);
        }

        await this.saveConfig();
        return existingIndex >= 0
            ? this.config.toolSets[existingIndex]
            : this.config.toolSets[this.config.toolSets.length - 1];
    }

    async deleteToolSet(nameOrUuid: string): Promise<void> {
        await this.init();
        if (!this.config.toolSets) return;

        this.config.toolSets = this.config.toolSets.filter(set => set.name !== nameOrUuid && set.uuid !== nameOrUuid);
        await this.saveConfig();
    }

    private async saveConfig(): Promise<void> {
        await writeTormentNexusMcpConfig(this.config, this.configDir);
    }
}

export const jsonConfigProvider = new JsonConfigProvider();
