import { McpServerCreateInput, McpServerTypeEnum } from "../types/mcp-admin/index.js";
import { mcpServersRepository } from "../db/repositories/mcp-servers.repo.js";

interface ClaudeMcpServerDefinition {
    command?: string;
    args?: string[];
    env?: Record<string, string>;
    url?: string;
}

interface ConfigImportResult {
    imported: number;
    skipped: string[];
}

function normalizeServerDefinition(value: unknown): ClaudeMcpServerDefinition {
    if (!value || typeof value !== "object") {
        return {};
    }

    const record = value as Record<string, unknown>;
    const args = Array.isArray(record.args)
        ? record.args.filter((arg): arg is string => typeof arg === "string")
        : [];

    const env = record.env && typeof record.env === "object"
        ? Object.fromEntries(
            Object.entries(record.env as Record<string, unknown>).filter(
                ([, val]) => typeof val === "string",
            ) as Array<[string, string]>,
        )
        : {};

    return {
        command: typeof record.command === "string" ? record.command : undefined,
        args,
        env,
        url: typeof record.url === "string" ? record.url : undefined,
    };
}

export class ConfigImportService {
    async importClaudeConfig(configJson: string, userId?: string | null): Promise<ConfigImportResult> {
        try {
            const config = JSON.parse(configJson);
            if (!config.mcpServers || typeof config.mcpServers !== "object") {
                throw new Error("Invalid configuration: 'mcpServers' object not found.");
            }

            const serversToCreate: McpServerCreateInput[] = [];
            const errors: string[] = [];

            for (const [name, definition] of Object.entries(config.mcpServers)) {
                // Validate name format (alphanumeric + underscore/hyphen)
                const safeName = name.replace(/[^a-zA-Z0-9_-]/g, "_");

                const def = normalizeServerDefinition(definition);

                if (def.command) {
                    // Stdio Server
                    serversToCreate.push({
                        name: safeName,
                        type: McpServerTypeEnum.Enum.STDIO,
                        command: def.command,
                        args: def.args || [],
                        env: def.env || {},
                        user_id: userId,
                    });
                } else if (def.url) {
                    // SSE Server (Assuming SSE for URL-based in config, usually)
                    serversToCreate.push({
                        name: safeName,
                        type: McpServerTypeEnum.Enum.SSE,
                        url: def.url,
                        user_id: userId,
                    });
                } else {
                    errors.push(`Skipped '${name}': Unknown server type (no command or url)`);
                }
            }

            if (serversToCreate.length > 0) {
                for (const server of serversToCreate) {
                    await mcpServersRepository.create(server);
                }
            }

            return {
                imported: serversToCreate.length,
                skipped: errors,
            };

        } catch (error) {
            console.error("Config import failed:", error);
            throw error;
        }
    }
}

export const configImportService = new ConfigImportService();
