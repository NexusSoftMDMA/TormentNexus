import { Hono } from 'hono';
import { fineTunedModelManager } from '../services/fine-tuned-model-manager.js';
import type { FineTuneProvider, FineTuneStatus, ModelDeploymentStatus } from '../services/fine-tuned-model-manager.js';
import type { TaskType } from './types.js';

const app = new Hono();

// ============ Dataset Routes ============

app.post('/datasets', async (c) => {
  const body = await c.req.json();
  const dataset = fineTunedModelManager.createDataset({
    name: body.name,
    description: body.description,
    taskTypes: body.taskTypes ?? ['general'],
    format: body.format,
    validationSplit: body.validationSplit,
    metadata: body.metadata,
  });
  return c.json({ success: true, data: dataset });
});

app.get('/datasets', async (c) => {
  const taskType = c.req.query('taskType') as TaskType | undefined;
  const datasets = fineTunedModelManager.listDatasets({ taskType });
  return c.json({ success: true, data: datasets });
});

app.get('/datasets/:id', async (c) => {
  const dataset = fineTunedModelManager.getDataset(c.req.param('id'));
  if (!dataset) {
    return c.json({ success: false, error: 'Dataset not found' }, 404);
  }
  return c.json({ success: true, data: dataset });
});

app.patch('/datasets/:id/example-count', async (c) => {
  const body = await c.req.json();
  fineTunedModelManager.updateDatasetExampleCount(c.req.param('id'), body.count);
  const dataset = fineTunedModelManager.getDataset(c.req.param('id'));
  return c.json({ success: true, data: dataset });
});

app.patch('/datasets/:id/file-id', async (c) => {
  const body = await c.req.json();
  fineTunedModelManager.setDatasetFileId(c.req.param('id'), body.fileId);
  const dataset = fineTunedModelManager.getDataset(c.req.param('id'));
  return c.json({ success: true, data: dataset });
});

app.delete('/datasets/:id', async (c) => {
  try {
    const deleted = fineTunedModelManager.deleteDataset(c.req.param('id'));
    if (!deleted) {
      return c.json({ success: false, error: 'Dataset not found' }, 404);
    }
    return c.json({ success: true });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/datasets/validate', async (c) => {
  const body = await c.req.json();
  const result = fineTunedModelManager.validateTrainingData(body.examples ?? []);
  return c.json({ success: true, data: result });
});

app.post('/datasets/format', async (c) => {
  const body = await c.req.json();
  const formatted = fineTunedModelManager.formatExamplesForProvider(
    body.examples ?? [],
    body.provider ?? 'openai'
  );
  return c.json({ success: true, data: { formatted } });
});

// ============ Job Routes ============

app.post('/jobs', async (c) => {
  try {
    const body = await c.req.json();
    const job = fineTunedModelManager.createFineTuneJob({
      baseModel: body.baseModel,
      datasetId: body.datasetId,
      provider: body.provider ?? 'openai',
      hyperparameters: body.hyperparameters,
      suffix: body.suffix,
      validationDatasetId: body.validationDatasetId,
    });
    return c.json({ success: true, data: job });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.get('/jobs', async (c) => {
  const status = c.req.query('status') as FineTuneStatus | undefined;
  const provider = c.req.query('provider') as FineTuneProvider | undefined;
  const jobs = fineTunedModelManager.listJobs({ status, provider });
  return c.json({ success: true, data: jobs });
});

app.get('/jobs/:id', async (c) => {
  const job = fineTunedModelManager.getJob(c.req.param('id'));
  if (!job) {
    return c.json({ success: false, error: 'Job not found' }, 404);
  }
  return c.json({ success: true, data: job });
});

app.post('/jobs/:id/start', async (c) => {
  try {
    const job = await fineTunedModelManager.startJob(c.req.param('id'));
    return c.json({ success: true, data: job });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/jobs/:id/cancel', async (c) => {
  try {
    const job = fineTunedModelManager.cancelJob(c.req.param('id'));
    return c.json({ success: true, data: job });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

// ============ Model Routes ============

app.post('/models', async (c) => {
  try {
    const body = await c.req.json();
    const model = fineTunedModelManager.registerModel({
      name: body.name,
      description: body.description,
      provider: body.provider ?? 'openai',
      providerModelId: body.providerModelId,
      baseModel: body.baseModel,
      jobId: body.jobId,
      taskTypes: body.taskTypes ?? ['general'],
      version: body.version ?? '1.0.0',
      config: body.config,
      tags: body.tags,
    });
    return c.json({ success: true, data: model });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.get('/models', async (c) => {
  const status = c.req.query('status') as ModelDeploymentStatus | undefined;
  const provider = c.req.query('provider') as FineTuneProvider | undefined;
  const taskType = c.req.query('taskType') as TaskType | undefined;
  const tag = c.req.query('tag');
  const models = fineTunedModelManager.listModels({ status, provider, taskType, tag });
  return c.json({ success: true, data: models });
});

app.get('/models/best/:taskType', async (c) => {
  const model = fineTunedModelManager.getBestModelForTask(c.req.param('taskType') as TaskType);
  if (!model) {
    return c.json({ success: false, error: 'No suitable model found' }, 404);
  }
  return c.json({ success: true, data: model });
});

app.get('/models/:id', async (c) => {
  const model = fineTunedModelManager.getModel(c.req.param('id'));
  if (!model) {
    return c.json({ success: false, error: 'Model not found' }, 404);
  }
  return c.json({ success: true, data: model });
});

app.post('/models/:id/deploy', async (c) => {
  try {
    const model = fineTunedModelManager.deployModel(c.req.param('id'));
    return c.json({ success: true, data: model });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/models/:id/deprecate', async (c) => {
  try {
    const model = fineTunedModelManager.deprecateModel(c.req.param('id'));
    return c.json({ success: true, data: model });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/models/:id/retire', async (c) => {
  try {
    const model = fineTunedModelManager.retireModel(c.req.param('id'));
    return c.json({ success: true, data: model });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.delete('/models/:id', async (c) => {
  try {
    const deleted = fineTunedModelManager.deleteModel(c.req.param('id'));
    if (!deleted) {
      return c.json({ success: false, error: 'Model not found' }, 404);
    }
    return c.json({ success: true });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.get('/models/:id/export', async (c) => {
  try {
    const data = fineTunedModelManager.exportModel(c.req.param('id'));
    return c.json({ success: true, data });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/models/import', async (c) => {
  try {
    const body = await c.req.json();
    const model = fineTunedModelManager.importModel(body);
    return c.json({ success: true, data: model });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/models/compare', async (c) => {
  try {
    const body = await c.req.json();
    const comparison = fineTunedModelManager.compareModels(
      body.modelA,
      body.modelB,
      body.taskType
    );
    return c.json({ success: true, data: comparison });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

// ============ Supervisor Routes ============

app.post('/models/:id/chat', async (c) => {
  try {
    const body = await c.req.json();
    const modelId = c.req.param('id');
    let supervisor = fineTunedModelManager.getSupervisor(modelId);

    if (!supervisor) {
      // Try to auto-create if model is active
      const model = fineTunedModelManager.getModel(modelId);
      if (model && model.deploymentStatus === 'active') {
        supervisor = fineTunedModelManager.createSupervisorFromModel(modelId);
      } else {
        return c.json({ success: false, error: 'Supervisor not found or model not active' }, 404);
      }
    }

    const messages = body.messages || [];
    const response = await supervisor.chat(messages);

    return c.json({ success: true, data: { response } });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.post('/models/:id/supervisor', async (c) => {
  try {
    const body = await c.req.json();
    const supervisor = fineTunedModelManager.createSupervisorFromModel(
      c.req.param('id'),
      body.overrides
    );
    return c.json({ 
      success: true, 
      data: { 
        name: supervisor.name, 
        modelId: supervisor.modelId,
        version: supervisor.version,
        taskTypes: supervisor.taskTypes,
      } 
    });
  } catch (error) {
    return c.json({ success: false, error: (error as Error).message }, 400);
  }
});

app.get('/supervisors', async (c) => {
  const supervisors = fineTunedModelManager.listSupervisors().map(s => ({
    name: s.name,
    modelId: s.modelId,
    version: s.version,
    taskTypes: s.taskTypes,
  }));
  return c.json({ success: true, data: supervisors });
});

app.get('/supervisors/:modelId', async (c) => {
  const supervisor = fineTunedModelManager.getSupervisor(c.req.param('modelId'));
  if (!supervisor) {
    return c.json({ success: false, error: 'Supervisor not found' }, 404);
  }
  return c.json({ 
    success: true, 
    data: {
      name: supervisor.name,
      modelId: supervisor.modelId,
      version: supervisor.version,
      taskTypes: supervisor.taskTypes,
    }
  });
});

app.get('/supervisors/:modelId/config', async (c) => {
  const config = fineTunedModelManager.getSupervisorConfig(c.req.param('modelId'));
  if (!config) {
    return c.json({ success: false, error: 'Supervisor config not found' }, 404);
  }
  const safeConfig = { ...config };
  delete safeConfig.apiKey;
  return c.json({ success: true, data: safeConfig });
});

// ============ Performance Routes ============

app.post('/models/:id/record-debate', async (c) => {
  const body = await c.req.json();
  fineTunedModelManager.recordDebateResult(c.req.param('id'), {
    approved: body.approved,
    confidence: body.confidence,
    alignedWithConsensus: body.alignedWithConsensus,
    taskType: body.taskType,
  });
  return c.json({ success: true });
});

app.get('/models/:id/performance', async (c) => {
  const model = fineTunedModelManager.getModel(c.req.param('id'));
  if (!model) {
    return c.json({ success: false, error: 'Model not found' }, 404);
  }
  return c.json({ success: true, data: model.performance });
});

app.get('/models/:id/performance/history', async (c) => {
  const since = c.req.query('since') ? parseInt(c.req.query('since')!) : undefined;
  const history = fineTunedModelManager.getPerformanceHistory(c.req.param('id'), since);
  return c.json({ success: true, data: history });
});

// ============ Provider Config Routes ============

app.post('/providers/:provider/config', async (c) => {
  const body = await c.req.json();
  fineTunedModelManager.setProviderConfig(
    c.req.param('provider') as FineTuneProvider,
    { apiKey: body.apiKey, baseURL: body.baseURL }
  );
  return c.json({ success: true });
});

app.get('/providers/:provider/config', async (c) => {
  const config = fineTunedModelManager.getProviderConfig(c.req.param('provider') as FineTuneProvider);
  if (!config) {
    return c.json({ success: true, data: {} });
  }
  return c.json({ success: true, data: { baseURL: config.baseURL, hasApiKey: !!config.apiKey } });
});

// ============ Statistics ============

app.get('/stats', async (c) => {
  const stats = fineTunedModelManager.getStatistics();
  return c.json({ success: true, data: stats });
});

export default app;
