/**
 * @file saved-scripts.zod.ts
 * @module packages/core/src/types/tormentnexus/saved-scripts.zod
 *
 * WHAT:
 * Zod definitions for User Scripts.
 *
 * WHY:
 * Validates the storage of executable code snippets (JS/Python) that users can save and run via the dashboard.
 */

import { z } from "zod";

export const SavedScriptSchema = z.object({
    uuid: z.string(),
    name: z.string(),
    description: z.string().nullable().optional(),
    code: z.string(),
    userId: z.string().nullable().optional(),
    createdAt: z.date().optional(),
    updatedAt: z.date().optional(),
});

export const GetSavedScriptsResponseSchema = z.object({
    success: z.literal(true),
    data: z.array(SavedScriptSchema),
});

export const DeleteSavedScriptRequestSchema = z.object({
    uuid: z.string(),
});

export const DeleteSavedScriptResponseSchema = z.object({
    success: z.boolean(),
    message: z.string().optional(),
});

export type SavedScript = z.infer<typeof SavedScriptSchema>;
