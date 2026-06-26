import { CallToolResult } from "@modelcontextprotocol/sdk/types.js";
import { eq } from "drizzle-orm";
import { db } from "../../db/index.js";
import { mcpServersTable, toolCallLogsTable } from "../../db/mcp-admin-schema.js";
import { CallToolMiddleware } from "./functional-middleware.js";
import { randomUUID } from "node:crypto";
import { parseToolName } from "../tool-name-parser.service.js";

const serverUuidCache = new Map<string, string | null>();

async function getServerUuidByName(serverName: string): Promise<string | null> {
    if (serverUuidCache.has(serverName)) {
        return serverUuidCache.get(serverName) ?? null;
    }

    try {
        const [server] = await db
            .select({ uuid: mcpServersTable.uuid })
            .from(mcpServersTable)
            .where(eq(mcpServersTable.name, serverName))
            .limit(1);

        const serverUuid = server?.uuid ?? null;
        serverUuidCache.set(serverName, serverUuid);
        return serverUuid;
    } catch {
        return null;
    }
}

export function createLoggingMiddleware(options?: {
    enabled?: boolean;
}): CallToolMiddleware {
    const enabled = options?.enabled ?? true;

    return (next) => async (request, context) => {
        if (!enabled) {
            return next(request, context);
        }

        const startTime = Date.now();
        let result: CallToolResult | null = null;
        let error: unknown = null;

        // Check for parent call ID in _meta (passed from run_code recursion)
        const paramsWithMeta = request.params as typeof request.params & {
            _meta?: { parentCallUuid?: string };
        };
        const parentCallUuid = paramsWithMeta._meta?.parentCallUuid;

        try {
            result = await next(request, context);
            return result;
        } catch (e) {
            error = e;
            throw e;
        } finally {
            const duration = Date.now() - startTime;
            const parsedToolName = parseToolName(request.params.name);
            const mcpServerUuid = parsedToolName
                ? await getServerUuidByName(parsedToolName.serverName)
                : null;

            // Log to DB asynchronously
            db.insert(toolCallLogsTable).values({
                uuid: randomUUID(),
                session_id: context.sessionId,
                tool_name: request.params.name,
                mcp_server_uuid: mcpServerUuid,
                args: (request.params.arguments as Record<string, unknown> | undefined) ?? null,
                result: (result as Record<string, unknown> | null) ?? null,
                error: error ? String(error) : null,
                duration_ms: duration,
                parent_call_uuid: parentCallUuid,
            }).catch((err) => {
                console.error("Failed to persist tool call log:", err);
            });
        }
    };
}
