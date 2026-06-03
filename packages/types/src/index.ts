// @tormentnexus/types -- schemas for tRPC routers
import { z } from 'zod';

export const getRecentObservationsInputSchema = z.object({
  limit: z.number().min(1).max(1000).default(50),
  namespace: z.string().optional(),
  type: z.string().optional(),
});

export const searchObservationsInputSchema = z.object({
  query: z.string(),
  limit: z.number().min(1).max(1000).default(20),
  namespace: z.string().optional(),
  type: z.string().optional(),
});

export const getRecentUserPromptsInputSchema = z.object({
  role: z.string().optional(),
  limit: z.number().min(1).max(1000).default(50),
});

export const searchUserPromptsInputSchema = z.object({
  query: z.string(),
  role: z.string().optional(),
  limit: z.number().min(1).max(1000).default(20),
});

export const observationTypeSchema = z.string();
export const userPromptRoleSchema = z.string();
export const memoryInterchangeFormatSchema = z.any();
export const structuredObservationSchema = z.any();
export const structuredUserPromptSchema = z.any();
export const searchMemoryPivotInputSchema = z.object({ query: z.string(), limit: z.number().default(20) });
export const getMemoryTimelineWindowInputSchema = z.object({ from: z.number(), to: z.number(), limit: z.number().default(50) });
export const getCrossSessionMemoryLinksInputSchema = z.object({ sessionId: z.string(), limit: z.number().default(20) });
