import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../../lib/trpc-core.js';
import { fineTunedModelManager } from '../../orchestrator/council/services/fine-tuned-model-manager.js';

const datasetSchema = z.object({
  name: z.string(),
  description: z.string().optional(),
  taskTypes: z.array(z.string()).optional(),
  format: z.enum(['alpaca', 'sharegpt', 'openai', 'messages']).optional(),
  validationSplit: z.number().optional(),
  metadata: z.any().optional(),
});

const jobSchema = z.object({
  baseModel: z.string(),
  datasetId: z.string(),
  provider: z.enum(['openai', 'anthropic', 'google', 'custom']).optional(),
  hyperparameters: z.any().optional(),
  suffix: z.string().optional(),
  validationDatasetId: z.string().optional(),
});

const modelSchema = z.object({
  name: z.string(),
  description: z.string().optional(),
  provider: z.string(),
  providerModelId: z.string(),
  baseModel: z.string().optional(),
  jobId: z.string().optional(),
  taskTypes: z.array(z.string()).optional(),
  version: z.string().optional(),
  config: z.any().optional(),
  tags: z.array(z.string()).optional(),
});

export const fineTuneRouter = t.router({
  // Datasets
  createDataset: adminProcedure.input(datasetSchema).mutation(async ({ input }) => {
    return fineTunedModelManager.createDataset(input as any);
  }),

  listDatasets: publicProcedure.input(z.object({ taskType: z.string().optional() }).optional()).query(async ({ input }) => {
    return fineTunedModelManager.listDatasets(input as any);
  }),

  getDataset: publicProcedure.input(z.object({ id: z.string() })).query(async ({ input }) => {
    const dataset = fineTunedModelManager.getDataset(input.id);
    if (!dataset) throw new Error('Dataset not found');
    return dataset;
  }),

  // Jobs
  createJob: adminProcedure.input(jobSchema).mutation(async ({ input }) => {
    return fineTunedModelManager.createFineTuneJob(input as any);
  }),

  listJobs: publicProcedure.input(z.any().optional()).query(async ({ input }) => {
    return fineTunedModelManager.listJobs(input);
  }),

  startJob: adminProcedure.input(z.object({ id: z.string() })).mutation(async ({ input }) => {
    return fineTunedModelManager.startJob(input.id);
  }),

  // Models
  registerModel: adminProcedure.input(modelSchema).mutation(async ({ input }) => {
    return fineTunedModelManager.registerModel(input as any);
  }),

  listModels: publicProcedure.input(z.any().optional()).query(async ({ input }) => {
    return fineTunedModelManager.listModels(input);
  }),

  deployModel: adminProcedure.input(z.object({ id: z.string() })).mutation(async ({ input }) => {
    return fineTunedModelManager.deployModel(input.id);
  }),

  chat: publicProcedure.input(z.object({
    id: z.string(),
    messages: z.array(z.any()),
  })).mutation(async ({ input }) => {
    let supervisor = fineTunedModelManager.getSupervisor(input.id);
    if (!supervisor) {
      const model = fineTunedModelManager.getModel(input.id);
      if (model && model.deploymentStatus === 'active') {
        supervisor = fineTunedModelManager.createSupervisorFromModel(input.id);
      } else {
        throw new Error('Supervisor not found or model not active');
      }
    }
    const response = await supervisor.chat(input.messages);
    return { response };
  }),

  stats: publicProcedure.query(async () => {
    return fineTunedModelManager.getStatistics();
  }),
});
