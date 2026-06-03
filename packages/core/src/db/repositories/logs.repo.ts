import { and, desc, eq } from "drizzle-orm";
import { randomUUID } from "node:crypto";
import { db } from "../index.js";
import { mcpServersTable, toolCallLogsTable } from "../mcp-admin-schema.js";
import { TormentNexusLogEntry } from "../../types/tormentnexus/logs.zod.js";

type ToolCallLogInsert = typeof toolCallLogsTable.$inferInsert;

export class LogsRepository {
    async create(input: {
        toolName: string;
        level: "error" | "info" | "warn";
        message: string;
        mcpServerUuid?: string;
        error?: string | null;
        arguments?: Record<string, unknown>;
        result?: Record<string, unknown>;
        durationMs?: number;
        sessionId?: string;
        parentCallUuid?: string;
    }): Promise<void> {
        const payload: ToolCallLogInsert = {
            uuid: randomUUID(),
            tool_name: input.toolName,
            args: input.arguments,
            result: input.result,
            error: input.error,
            duration_ms: input.durationMs,
            session_id: input.sessionId,
            parent_call_uuid: input.parentCallUuid,
            mcp_server_uuid: input.mcpServerUuid,
        };

        await db.insert(toolCallLogsTable).values(payload);
    }

    async findAll(input?: {
        limit?: number;
        sessionId?: string;
        serverName?: string;
    }): Promise<TormentNexusLogEntry[]> {
        const limit = input?.limit ?? 100;
        const filters = [
            input?.sessionId ? eq(toolCallLogsTable.session_id, input.sessionId) : undefined,
            input?.serverName ? eq(mcpServersTable.name, input.serverName) : undefined,
        ].filter((value): value is NonNullable<typeof value> => Boolean(value));

        const logs = await db
            .select({
                uuid: toolCallLogsTable.uuid,
                created_at: toolCallLogsTable.created_at,
                tool_name: toolCallLogsTable.tool_name,
                error: toolCallLogsTable.error,
                args: toolCallLogsTable.args,
                result: toolCallLogsTable.result,
                duration_ms: toolCallLogsTable.duration_ms,
                session_id: toolCallLogsTable.session_id,
                parent_call_uuid: toolCallLogsTable.parent_call_uuid,
                server_name: mcpServersTable.name,
            })
            .from(toolCallLogsTable)
            .leftJoin(mcpServersTable, eq(toolCallLogsTable.mcp_server_uuid, mcpServersTable.uuid))
            .where(filters.length > 0 ? and(...filters) : undefined)
            .orderBy(desc(toolCallLogsTable.created_at))
            .limit(limit);

        return logs.map((log) => ({
            id: log.uuid,
            timestamp: new Date(log.created_at),
            serverName: log.server_name ?? undefined,
            level: log.error ? "error" : "info",
            message: `Tool call: ${log.tool_name}`,
            toolName: log.tool_name,
            error: log.error,
            arguments: log.args ?? undefined,
            result: log.result ?? undefined,
            durationMs: log.duration_ms?.toString(),
            sessionId: log.session_id ?? undefined,
            parentCallUuid: log.parent_call_uuid ?? undefined,
        }));
    }

    async clear(): Promise<void> {
        await db.delete(toolCallLogsTable);
    }
}

export const logsRepository = new LogsRepository();
