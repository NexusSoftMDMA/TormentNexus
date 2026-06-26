import { z } from 'zod';
import { t, publicProcedure } from '../../lib/trpc-core.js';
import { diagramService } from '../../orchestrator/council/services/diagram-service.js';
import { cliRegistry } from '../../orchestrator/council/services/cli-registry.js';
import { sessionManager } from '../../orchestrator/council/services/session-manager.js';

export const visualRouter = t.router({
  systemDiagram: publicProcedure.query(async () => {
    const tools = cliRegistry.getAvailableTools().map(t => t.name);
    const sessions = sessionManager.getActiveSessions().map(s => ({
      id: s.id,
      cliType: sessionManager.getSessionCLIType(s.id) || 'unknown',
    }));
    
    return {
      mermaid: diagramService.generateSystemMermaid(tools, sessions),
    };
  }),

  planDiagram: publicProcedure.input(z.any()).query(async ({ input }) => {
    return {
      mermaid: diagramService.generateSwarmMermaid(input),
    };
  }),

  parsePlan: publicProcedure.input(z.object({ mermaid: z.string() })).mutation(async ({ input }) => {
    return diagramService.parseMermaidToPlan(input.mermaid);
  }),
});
