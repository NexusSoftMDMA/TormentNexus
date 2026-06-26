import {
    McpServerErrorStatusEnum,
    McpServerStatusEnum,
    ServerParameters,
} from "../types/mcp-admin/index.js";
import { jsonConfigProvider } from "./config/JsonConfigProvider.js";
import { getDefaultEnvironment, IOType } from "./common-utils.js";
import { v5 as uuidv5 } from 'uuid';

const NAMESPACE = '1b671a64-40d5-491e-99b0-da01ff1f3341'; // Random namespace for UUID generation

export async function getMcpServers(
    namespaceUuid: string,
    includeInactiveServers: boolean = false,
): Promise<Record<string, ServerParameters>> {
    try {
        const servers = await jsonConfigProvider.loadMcpServers();
        const serverDict: Record<string, ServerParameters> = {};

        for (const server of servers) {
            if (server.disabled && !includeInactiveServers) {
                continue;
            }

            // Generate a stable UUID based on the name
            const uuid = uuidv5(server.name, NAMESPACE);

            const params: ServerParameters = {
                uuid: uuid,
                name: server.name,
                description: "", // Config doesn't have description yet
                type: server.type === 'stdio' ? "STDIO" : "SSE",
                command: server.command,
                args: server.args || [],
                env: server.env || {},
                url: server.url,
                headers: {},
                created_at: new Date().toISOString(),
                status: server.disabled ? "inactive" : "active",
                error_status: "none",
                stderr: "inherit" as IOType,
                oauth_tokens: null, // OAuth not supported in JSON config yet
                bearerToken: undefined,
            };

            // Process based on server type
            if (params.type === "STDIO") {
                if ("args" in params && !params.args) {
                    params.args = undefined; // SDK expects undefined for empty args sometimes?
                }

                params.env = {
                    ...getDefaultEnvironment(),
                    ...(params.env || {}),
                };
            } else if (params.type === "SSE" || params.type === "STREAMABLE_HTTP") {
                // For SSE or STREAMABLE_HTTP servers, ensure url is present
                if (!params.url) {
                    console.warn(
                        `${params.type} server ${params.uuid} is missing url field, skipping`,
                    );
                    continue;
                }
            }

            serverDict[uuid] = params;
        }

        return serverDict;

    } catch (error) {
        console.error("Error fetching MCP servers from config:", error);
        return {};
    }
}
