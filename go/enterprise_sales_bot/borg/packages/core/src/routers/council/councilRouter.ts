import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../../lib/trpc-core.js';
import { council } from '../../orchestrator/council/services/council.js';
import { wsManager } from '../../orchestrator/council/services/ws-manager.js';
import { councilHierarchy } from '../../orchestrator/council/services/council-hierarchy.js';
import { createSupervisor, createSupervisors, createMockSupervisor } from '../../orchestrator/council/supervisors/index.js';

const supervisorConfigSchema = z.object({
  name: z.string(),
  provider: z.enum(['openai', 'anthropic', 'google', 'xai', 'moonshot', 'deepseek', 'qwen', 'custom', 'gemini', 'grok', 'kimi']),
  apiKey: z.string().optional(),
  model: z.string().optional(),
  temperature: z.number().optional(),
  baseURL: z.string().optional(),
  systemPrompt: z.string().optional(),
  weight: z.number().optional(),
});

const configSchema = z.object({
  supervisors: z.array(supervisorConfigSchema).optional(),
  debateRounds: z.number().int().min(1).max(10).optional(),
  consensusThreshold: z.number().min(0).max(1).optional(),
  enabled: z.boolean().optional(),
  smartPilot: z.boolean().optional(),
  weightedVoting: z.boolean().optional(),
  consensusMode: z.enum(['simple-majority', 'supermajority', 'unanimous', 'weighted', 'ceo-override', 'ceo-veto', 'hybrid-ceo-majority', 'ranked-choice']).optional(),
  leadSupervisor: z.string().optional(),
  fallbackSupervisors: z.array(z.string()).optional(),
});

const developmentTaskSchema = z.object({
  id: z.string(),
  description: z.string(),
  context: z.string(),
  files: z.array(z.string()),
});

export const councilRouter = t.router({
  status: publicProcedure.query(async () => {
    const available = await council.getAvailableSupervisors();
    const hierarchy = councilHierarchy.getAllSpecializedCouncils().map((s) => ({
      id: s.id,
      name: s.name,
      specialties: s.specialties,
      supervisorCount: s.council.getSupervisors().length
    }));
    
    const config = council.getConfig();
    
    return {
      enabled: config.enabled ?? true,
      supervisorCount: council.getSupervisors().length,
      availableCount: available.length,
      config,
      hierarchy,
    };
  }),

  updateConfig: adminProcedure.input(configSchema).mutation(async ({ input }) => {
    if (input.consensusMode) council.setConsensusMode(input.consensusMode);
    if (input.leadSupervisor) council.setLeadSupervisor(input.leadSupervisor);
    if (input.fallbackSupervisors) council.setFallbackChain(input.fallbackSupervisors);
    if (input.debateRounds) council.setDebateRounds(input.debateRounds);
    if (input.consensusThreshold) council.setConsensusThreshold(input.consensusThreshold);
    
    wsManager.broadcast({
      type: 'log',
      payload: { level: 'info', message: `Council config updated`, timestamp: Date.now() },
      timestamp: Date.now(),
    });
    
    return council.getConfig();
  }),

  addSupervisors: adminProcedure.input(z.object({
    supervisors: z.array(supervisorConfigSchema)
  })).mutation(async ({ input }) => {
    const added: string[] = [];
    const failed: string[] = [];

    for (const supervisorConfig of input.supervisors) {
      try {
        const supervisor = createSupervisor(supervisorConfig as any);
        council.addSupervisor(supervisor);
        added.push(supervisorConfig.name);
      } catch (error) {
        const msg = error instanceof Error ? error.message : 'Unknown error';
        failed.push(`${supervisorConfig.name}: ${msg}`);
      }
    }

    wsManager.broadcast({
      type: 'log',
      payload: { level: 'info', message: `Added ${added.length} supervisors: ${added.join(', ')}`, timestamp: Date.now() },
      timestamp: Date.now(),
    });

    return { added, failed };
  }),

  clearSupervisors: adminProcedure.mutation(async () => {
    council.clearSupervisors();
    
    wsManager.broadcast({
      type: 'log',
      payload: { level: 'info', message: 'All supervisors removed', timestamp: Date.now() },
      timestamp: Date.now(),
    });
    
    return { success: true };
  }),

  debate: adminProcedure.input(developmentTaskSchema).mutation(async ({ input }) => {
    wsManager.broadcast({
      type: 'log',
      payload: { level: 'info', message: `Debate started: ${input.description}`, timestamp: Date.now(), source: 'council' },
      timestamp: Date.now(),
    });
    
    try {
      // @ts-expect-error -- type mismatch from inferred schema types
      const decision = await council.debate(input);
      
      wsManager.broadcast({
        type: 'council_decision',
        payload: decision,
        timestamp: Date.now(),
      });
      
      return decision;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      wsManager.notifyError(`Debate failed: ${message}`);
      throw new Error(`Debate failed: ${message}`);
    }
  }),

  toggle: adminProcedure.mutation(async () => {
    const config = council.getConfig();
    const newEnabled = !config.enabled;
    config.enabled = newEnabled;
    
    wsManager.broadcast({
      type: 'log',
      payload: { level: 'info', message: `Council ${newEnabled ? 'enabled' : 'disabled'}`, timestamp: Date.now() },
      timestamp: Date.now(),
    });
    
    return { enabled: newEnabled };
  }),

  addMock: adminProcedure.mutation(async () => {
    const mockName = `MockSupervisor-${Date.now()}`;
    council.addSupervisor(createMockSupervisor(mockName));
    
    wsManager.broadcast({
      type: 'log',
      payload: { level: 'info', message: `Mock supervisor added: ${mockName}`, timestamp: Date.now() },
      timestamp: Date.now(),
    });
    
    return { added: mockName };
  }),
});
