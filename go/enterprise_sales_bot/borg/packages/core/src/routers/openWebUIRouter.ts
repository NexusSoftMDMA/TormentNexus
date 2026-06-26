import { z } from 'zod';
import { t, publicProcedure } from '../lib/trpc-core.js';

export const openWebUIRouter = t.router({
    getStatus: publicProcedure.query(async () => {
        // Return integration status for the dashboard
        return {
            status: 'active',
            version: '0.99.1',
            connected_tools: 0,
            message: 'Open-WebUI integration is initialized and ready.',
            timestamp: new Date().toISOString()
        };
    }),

    getEmbedUrl: publicProcedure.query(async () => {
        return {
            url: process.env.OPEN_WEBUI_URL || 'http://localhost:8080'
        };
    }),
});
