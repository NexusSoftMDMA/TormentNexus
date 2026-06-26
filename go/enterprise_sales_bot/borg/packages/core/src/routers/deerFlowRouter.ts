import { t, publicProcedure } from '../lib/trpc-core.js';
import { deerFlowBridgeService } from '../services/DeerFlowBridgeService.js';

export const deerFlowRouter = t.router({
    /**
     * Check if the Bytedance DeerFlow Python backend is active and accessible.
     */
    status: publicProcedure.query(async () => {
        const active = await deerFlowBridgeService.isAvailable();
        return { active };
    }),

    /**
     * Proxy to fetch configured models in the DeerFlow environment.
     */
    models: publicProcedure.query(async () => {
        if (!(await deerFlowBridgeService.isAvailable())) {
            return { error: 'DeerFlow is not currently running.', models: [] };
        }
        const models = await deerFlowBridgeService.getModels();
        return { models };
    }),

    /**
     * Proxy to fetch available Agent skills in DeerFlow.
     */
    skills: publicProcedure.query(async () => {
        if (!(await deerFlowBridgeService.isAvailable())) {
            return { error: 'DeerFlow is not currently running.', skills: [] };
        }
        const skills = await deerFlowBridgeService.getSkills();
        return { skills };
    }),

    /**
     * Proxy to fetch top-of-mind and long-term extracted User memory from DeerFlow.
     */
    memory: publicProcedure.query(async () => {
        if (!(await deerFlowBridgeService.isAvailable())) {
            return { error: 'DeerFlow is not currently running.', memory: null };
        }
        const status = await deerFlowBridgeService.getMemoryStatus();
        return { memory: status };
    }),
});
