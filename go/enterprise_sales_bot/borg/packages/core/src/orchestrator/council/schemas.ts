import { z } from 'zod';

export const supervisorConfigSchema = z.object({
  name: z.string().min(1),
  provider: z.enum(['openai', 'anthropic', 'google', 'xai', 'moonshot', 'deepseek', 'qwen', 'custom', 'gemini', 'grok', 'kimi']),
  apiKey: z.string().optional(),
  model: z.string().optional(),
  temperature: z.number().min(0).max(2).optional(),
  baseURL: z.string().url().optional(),
  systemPrompt: z.string().optional(),
  weight: z.number().min(0).max(2).optional(),
});

export const debateRequestSchema = z.object({
  task: z.object({
    id: z.string().min(1),
    description: z.string().min(1),
    context: z.string(),
    files: z.array(z.string()),
    timestamp: z.number().optional(),
  }),
});

export const smartPilotConfigSchema = z.object({
  autoApproveThreshold: z.number().min(0).max(1).optional(),
  maxAutoApprovals: z.number().min(0).optional(),
  requireUnanimous: z.boolean().optional(),
});

export const hookRegisterSchema = z.object({
  id: z.string().min(1),
  phase: z.enum(['pre-debate', 'post-debate', 'pre-guidance', 'post-guidance', 'on-error']),
  webhookUrl: z.string().url(),
  priority: z.number().int().min(0).max(100).optional(),
});

export const hookUnregisterSchema = z.object({
  id: z.string().min(1),
});

export type SupervisorConfigInput = z.infer<typeof supervisorConfigSchema>;
export type DebateRequestInput = z.infer<typeof debateRequestSchema>;
export type SmartPilotConfigInput = z.infer<typeof smartPilotConfigSchema>;
export type HookRegisterInput = z.infer<typeof hookRegisterSchema>;
