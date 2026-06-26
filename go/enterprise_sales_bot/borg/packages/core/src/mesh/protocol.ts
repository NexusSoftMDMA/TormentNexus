import { z } from 'zod';

export const MeshProtocolVersionSchema = z.literal('1.0');

export const SwarmMessageTypeSchema = z.enum([
    'CAPABILITY_QUERY',
    'CAPABILITY_RESPONSE',
    'TASK_OFFER',
    'TASK_ACCEPT',
    'TASK_RESULT',
    'VERIFY_OFFER',
    'VERIFY_RESULT',
    'HEARTBEAT',
    'DIRECT_MESSAGE',
    'TASK_BID',
    'TASK_ASSIGN',
    'ARTIFACT_READ_REQUEST',
    'ARTIFACT_READ_RESPONSE',
    'A2A_BRIDGE_SIGNAL',
]);

export const SwarmMessageMetadataSchema = z.object({
    priority: z.number().int().default(0),
    ttlMs: z.number().int().positive().default(300_000),
    requiresAck: z.boolean().default(false),
}).partial();

export const SwarmMessageSchema = z.object({
    version: MeshProtocolVersionSchema.default('1.0'),
    id: z.string().uuid(),
    correlationId: z.string().uuid().optional(),
    sender: z.string().min(1),
    target: z.string().min(1).optional(),
    type: SwarmMessageTypeSchema,
    payload: z.unknown(),
    timestamp: z.number().int(),
    metadata: SwarmMessageMetadataSchema.optional(),
});

export type SwarmMessageMetadata = z.infer<typeof SwarmMessageMetadataSchema>;
