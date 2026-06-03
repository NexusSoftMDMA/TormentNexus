/**
 * @file logs.zod.ts
 * @module packages/core/src/types/tormentnexus/logs.zod
 *
 * WHAT:
 * Zod definitions for Observability Logs (Tool Calls & Docker events).
 *
 * WHY:
 * Validates the structure of logs stored in `tool_call_logs`, including tool args, results, duration, and errors.
 * Also handles Docker logs request objects.
 */

import { z } from "zod";

export const TormentNexusLogEntrySchema = z.object({
    id: z.string(),
    timestamp: z.date(),
    serverName: z.string().optional(), // Now derived from tool_name prefix or similar
    level: z.enum(["error", "info", "warn"]),
    message: z.string(), // Summary: "Called tool X"
    error: z.string().nullable().optional(),

    // New structured fields
    toolName: z.string().optional(),
    arguments: z.record(z.unknown()).optional(),
    result: z.record(z.unknown()).optional(),
    durationMs: z.string().optional(),
    sessionId: z.string().optional(),
    parentCallUuid: z.string().nullable().optional(),
});

export const GetLogsRequestSchema = z.object({
    limit: z.number().int().positive().max(1000).optional(),
    sessionId: z.string().optional(),
    serverName: z.string().min(1).optional(),
});

export const GetLogsResponseSchema = z.object({
    success: z.literal(true),
    data: z.array(TormentNexusLogEntrySchema),
    totalCount: z.number(),
});

export const ClearLogsResponseSchema = z.object({
    success: z.literal(true),
    message: z.string(),
});

// Docker logs (per MCP server)
export const ListDockerServersResponseSchema = z.object({
    success: z.literal(true),
    servers: z.array(
        z.object({
            serverUuid: z.string(),
            containerId: z.string(),
            containerName: z.string(),
            serverName: z.string(),
        }),
    ),
});

export const GetDockerLogsRequestSchema = z.object({
    serverUuid: z.string(),
    tail: z.number().int().positive().max(5000).optional(),
});

export const GetDockerLogsResponseSchema = z.object({
    success: z.literal(true),
    serverUuid: z.string(),
    lines: z.array(z.string()),
});

export type TormentNexusLogEntry = z.infer<typeof TormentNexusLogEntrySchema>;
export type GetLogsRequest = z.infer<typeof GetLogsRequestSchema>;
export type GetLogsResponse = z.infer<typeof GetLogsResponseSchema>;
export type ClearLogsResponse = z.infer<typeof ClearLogsResponseSchema>;
export type ListDockerServersResponse = z.infer<
    typeof ListDockerServersResponseSchema
>;
export type GetDockerLogsRequest = z.infer<typeof GetDockerLogsRequestSchema>;
export type GetDockerLogsResponse = z.infer<typeof GetDockerLogsResponseSchema>;
