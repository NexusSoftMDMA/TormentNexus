import { Hono } from 'hono';
import { z } from 'zod';
import type { CouncilConfig, CouncilDecision, ApiResponse, SupervisorConfig, ConsensusMode, TaskType, SpecializedCouncilConfig } from './types.js';
import { SupervisorCouncil } from '../services/council.js';
import { councilHierarchy } from '../services/council-hierarchy.js';
import { createSupervisor, createSupervisors, createMockSupervisor } from '../supervisors/index.js';
import { wsManager } from '../services/ws-manager.js';
import { debateRateLimit, apiRateLimit } from '../middleware/rate-limit.js';
import { apiKeyAuth } from '../middleware/auth.js';
import { debateRequestSchema, supervisorConfigSchema } from '../schemas.js';

const council = new Hono();

const consensusModes: ConsensusMode[] = [
  'simple-majority', 'supermajority', 'unanimous', 'weighted',
  'ceo-override', 'ceo-veto', 'hybrid-ceo-majority', 'ranked-choice'
];

let config: CouncilConfig = {
  supervisors: [],
  debateRounds: 2,
  consensusThreshold: 0.7,
  enabled: true,
  smartPilot: false,
  consensusMode: 'weighted',
};

let councilInstance: SupervisorCouncil | null = null;

function getOrCreateCouncil(): SupervisorCouncil {
  if (!councilInstance) {
    councilInstance = new SupervisorCouncil(config);
    
    if (config.supervisors.length > 0) {
      const supervisors = createSupervisors(config.supervisors);
      supervisors.forEach((s) => councilInstance!.addSupervisor(s));
    }
  }
  return councilInstance;
}

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

const specializedCouncilSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string(),
  specialties: z.array(z.string()),
  config: configSchema,
});

const supervisorsBodySchema = z.object({
  supervisors: z.array(supervisorConfigSchema),
});

council.get('/status', apiRateLimit(), async (c) => {
  const instance = getOrCreateCouncil();
  const available = await instance.getAvailableSupervisors();
  const hierarchy = councilHierarchy.getAllSpecializedCouncils().map(s => ({
    id: s.id,
    name: s.name,
    specialties: s.specialties,
    supervisorCount: s.council.getSupervisors().length
  }));
  
  return c.json<ApiResponse<{ enabled: boolean; supervisorCount: number; availableCount: number; config: CouncilConfig, hierarchy: any[] }>>({
    success: true,
    data: {
      enabled: config.enabled ?? true,
      supervisorCount: config.supervisors.length,
      availableCount: available.length,
      config,
      hierarchy,
    },
  });
});

council.post('/specialized', apiRateLimit(), apiKeyAuth, async (c) => {
  const body = specializedCouncilSchema.parse(await c.req.json());
  
  councilHierarchy.registerSpecializedCouncil({
    ...body.config,
    id: body.id,
    name: body.name,
    description: body.description,
    specialties: body.specialties as TaskType[],
    // @ts-expect-error -- type mismatch from inferred schema types
    supervisors: body.config.supervisors || [],
  });

  return c.json({ success: true, message: `Specialized council '${body.name}' registered.` });
});

council.post('/config', apiRateLimit(), apiKeyAuth, async (c) => {
  const body = configSchema.parse(await c.req.json());
  // @ts-expect-error -- type mismatch from inferred schema types
  config = { ...config, ...body };
  
  councilInstance = new SupervisorCouncil(config);
  if (config.supervisors && config.supervisors.length > 0) {
    const supervisors = createSupervisors(config.supervisors);
    supervisors.forEach((s) => councilInstance!.addSupervisor(s));
  }
  
  if (body.consensusMode) {
    councilInstance.setConsensusMode(body.consensusMode);
  }
  if (body.leadSupervisor) {
    councilInstance.setLeadSupervisor(body.leadSupervisor);
  }
  if (body.fallbackSupervisors) {
    councilInstance.setFallbackChain(body.fallbackSupervisors);
  }
  
  wsManager.broadcast({
    type: 'log',
    payload: { level: 'info', message: `Council config updated: ${config.supervisors.length} supervisors, mode: ${config.consensusMode}`, timestamp: Date.now() },
    timestamp: Date.now(),
  });
  
  return c.json<ApiResponse<CouncilConfig>>({ success: true, data: config });
});

council.post('/supervisors', apiRateLimit(), apiKeyAuth, async (c) => {
  const body = supervisorsBodySchema.parse(await c.req.json());

  const instance = getOrCreateCouncil();
  const added: string[] = [];
  const failed: string[] = [];

  for (const supervisorConfig of body.supervisors) {
    try {
      const supervisor = createSupervisor(supervisorConfig as SupervisorConfig);
      instance.addSupervisor(supervisor);
      added.push(supervisorConfig.name);
      
      config.supervisors.push(supervisorConfig as SupervisorConfig);
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

  return c.json<ApiResponse<{ added: string[]; failed: string[] }>>({
    success: true,
    data: { added, failed },
  });
});

council.delete('/supervisors', apiRateLimit(), apiKeyAuth, (c) => {
  config.supervisors = [];
  councilInstance = new SupervisorCouncil(config);
  
  wsManager.broadcast({
    type: 'log',
    payload: { level: 'info', message: 'All supervisors removed', timestamp: Date.now() },
    timestamp: Date.now(),
  });
  
  return c.json<ApiResponse<{ message: string }>>({ success: true, data: { message: 'All supervisors removed' } });
});

council.post('/debate', debateRateLimit(), apiKeyAuth, async (c) => {
  if (!config.enabled) {
    const decision: CouncilDecision = {
      approved: true,
      consensus: 1.0,
      votes: [],
      reasoning: 'Council is disabled - auto-approving',
    };
    
    wsManager.broadcast({
      type: 'council_decision',
      payload: decision,
      timestamp: Date.now(),
    });
    
    return c.json<ApiResponse<CouncilDecision>>({ success: true, data: decision });
  }

  const task = debateRequestSchema.shape.task.parse(await c.req.json());
  const instance = getOrCreateCouncil();
  
  wsManager.broadcast({
    type: 'log',
    payload: { level: 'info', message: `Debate started: ${task.description}`, timestamp: Date.now(), source: 'council' },
    timestamp: Date.now(),
  });
  
  try {
    // @ts-expect-error -- type mismatch from inferred schema types
    const decision = await instance.debate(task);
    
    wsManager.broadcast({
      type: 'council_decision',
      payload: decision,
      timestamp: Date.now(),
    });
    
    wsManager.broadcast({
      type: 'log',
      payload: { 
        level: decision.approved ? 'info' : 'warn', 
        message: `Debate completed: ${decision.approved ? 'APPROVED' : 'REJECTED'} (${Math.round(decision.consensus * 100)}% consensus)`,
        timestamp: Date.now(),
        source: 'council'
      },
      timestamp: Date.now(),
    });
    
    return c.json<ApiResponse<CouncilDecision>>({ success: true, data: decision });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error';
    
    wsManager.notifyError(`Debate failed: ${message}`);
    
    return c.json<ApiResponse<CouncilDecision>>({ 
      success: false, 
      error: `Debate failed: ${message}` 
    }, 500);
  }
});

council.post('/toggle', apiRateLimit(), apiKeyAuth, (c) => {
  config.enabled = !config.enabled;
  
  wsManager.broadcast({
    type: 'log',
    payload: { level: 'info', message: `Council ${config.enabled ? 'enabled' : 'disabled'}`, timestamp: Date.now() },
    timestamp: Date.now(),
  });
  
  return c.json<ApiResponse<{ enabled: boolean }>>({ success: true, data: { enabled: config.enabled } });
});

council.post('/add-mock', apiRateLimit(), apiKeyAuth, (c) => {
  const instance = getOrCreateCouncil();
  const mockName = `MockSupervisor-${Date.now()}`;
  instance.addSupervisor(createMockSupervisor(mockName));
  
  wsManager.broadcast({
    type: 'log',
    payload: { level: 'info', message: `Mock supervisor added: ${mockName}`, timestamp: Date.now() },
    timestamp: Date.now(),
  });
  
  return c.json<ApiResponse<{ added: string }>>({ 
    success: true, 
    data: { added: mockName } 
  });
});

council.get('/consensus-modes', apiRateLimit(), (c) => {
  return c.json<ApiResponse<{ modes: ConsensusMode[]; current: ConsensusMode }>>({
    success: true,
    data: { modes: consensusModes, current: config.consensusMode ?? 'weighted' },
  });
});

council.post('/consensus-mode', apiRateLimit(), apiKeyAuth, async (c) => {
  const { mode } = z.object({ mode: z.enum(['simple-majority', 'supermajority', 'unanimous', 'weighted', 'ceo-override', 'ceo-veto', 'hybrid-ceo-majority', 'ranked-choice']) }).parse(await c.req.json());
  config.consensusMode = mode;
  
  const instance = getOrCreateCouncil();
  instance.setConsensusMode(mode);
  
  wsManager.broadcast({
    type: 'log',
    payload: { level: 'info', message: `Consensus mode changed to: ${mode}`, timestamp: Date.now() },
    timestamp: Date.now(),
  });
  
  return c.json<ApiResponse<{ mode: ConsensusMode }>>({ success: true, data: { mode } });
});

council.post('/lead-supervisor', apiRateLimit(), apiKeyAuth, async (c) => {
  const { name } = z.object({ name: z.string() }).parse(await c.req.json());
  config.leadSupervisor = name;
  
  const instance = getOrCreateCouncil();
  instance.setLeadSupervisor(name);
  
  wsManager.broadcast({
    type: 'log',
    payload: { level: 'info', message: `Lead supervisor set to: ${name}`, timestamp: Date.now() },
    timestamp: Date.now(),
  });
  
  return c.json<ApiResponse<{ leadSupervisor: string }>>({ success: true, data: { leadSupervisor: name } });
});

council.post('/fallback-chain', apiRateLimit(), apiKeyAuth, async (c) => {
  const { supervisors } = z.object({ supervisors: z.array(z.string()) }).parse(await c.req.json());
  config.fallbackSupervisors = supervisors;
  
  const instance = getOrCreateCouncil();
  instance.setFallbackChain(supervisors);
  
  wsManager.broadcast({
    type: 'log',
    payload: { level: 'info', message: `Fallback chain set: ${supervisors.join(' → ')}`, timestamp: Date.now() },
    timestamp: Date.now(),
  });
  
  return c.json<ApiResponse<{ fallbackChain: string[] }>>({ success: true, data: { fallbackChain: supervisors } });
});

council.post('/supervisor-weight', apiRateLimit(), apiKeyAuth, async (c) => {
  const { name, weight } = z.object({ name: z.string(), weight: z.number().min(0).max(2) }).parse(await c.req.json());
  
  const instance = getOrCreateCouncil();
  instance.setSupervisorWeight(name, weight);
  
  wsManager.broadcast({
    type: 'log',
    payload: { level: 'info', message: `Supervisor ${name} weight set to: ${weight}`, timestamp: Date.now() },
    timestamp: Date.now(),
  });
  
  return c.json<ApiResponse<{ name: string; weight: number }>>({ success: true, data: { name, weight } });
});

export { council as councilRoutes };
