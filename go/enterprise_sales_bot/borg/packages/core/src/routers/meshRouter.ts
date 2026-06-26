import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { z } from 'zod';
import { MeshService, SwarmMessageType } from '../mesh/MeshService.js';
import { AgentCapabilityRegistry } from '../services/AgentCapabilityRegistry.js';

// We create a singleton or a way to access the global mesh.
// In actual implementation, this might be attached to the global orchestrator
let meshServiceInstance: MeshService | null = null;
let capabilityRegistryInstance: AgentCapabilityRegistry | null = null;

function getMeshService(): MeshService {
    if (!meshServiceInstance) {
        meshServiceInstance = new MeshService();
    }
    return meshServiceInstance;
}

function getCapabilityRegistry(): AgentCapabilityRegistry {
    if (!capabilityRegistryInstance) {
        capabilityRegistryInstance = new AgentCapabilityRegistry(getMeshService());
    }
    return capabilityRegistryInstance;
}

export const meshRouter = t.router({
    getStatus: adminProcedure.query(async () => {
        const mesh = getMeshService();
        return {
            nodeId: mesh.nodeId,
            peersCount: mesh.getPeers().length
        };
    }),
    
    getPeers: adminProcedure.query(async () => {
        const mesh = getMeshService();
        return mesh.getPeers();
    }),

    getCapabilities: adminProcedure.query(async () => {
        const mesh = getMeshService();
        return mesh.getMeshCapabilities();
    }),

    queryCapabilities: adminProcedure.input(z.object({
        nodeId: z.string().min(1),
        timeoutMs: z.number().int().min(100).max(30_000).optional(),
    })).query(async ({ input }) => {
        return await getCapabilityRegistry().discoverCapabilities(input.nodeId, input.timeoutMs);
    }),

    findPeerForCapabilities: adminProcedure.input(z.object({
        requiredCapabilities: z.array(z.string().min(1)).min(1),
        timeoutMs: z.number().int().min(100).max(30_000).optional(),
    })).query(async ({ input }) => {
        return await getCapabilityRegistry().findPeerForCapabilities(
            input.requiredCapabilities,
            input.timeoutMs,
        );
    }),

    broadcast: adminProcedure
        .input(z.object({
            type: z.nativeEnum(SwarmMessageType),
            payload: z.any()
        }))
        .mutation(async ({ input }) => {
            const mesh = getMeshService();
            mesh.broadcast(input.type, input.payload);
            return { success: true };
        })
});
