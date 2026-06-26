import { describe, it, expect, beforeEach } from 'vitest';
import { FineTunedModelManagerService } from '../fine-tuned-model-manager.js';

describe('FineTunedModelManagerService', () => {
  let manager: FineTunedModelManagerService;

  beforeEach(() => {
    manager = new FineTunedModelManagerService();
  });

  describe('Dataset Management', () => {
    it('should create a dataset', () => {
      const dataset = manager.createDataset({
        name: 'test-dataset',
        description: 'Test description',
        taskTypes: ['code-review', 'security-audit'],
      });

      expect(dataset.id).toMatch(/^dataset_/);
      expect(dataset.name).toBe('test-dataset');
      expect(dataset.taskTypes).toEqual(['code-review', 'security-audit']);
      expect(dataset.exampleCount).toBe(0);
      expect(dataset.format).toBe('jsonl');
    });

    it('should list datasets with filter', () => {
      manager.createDataset({ name: 'ds1', taskTypes: ['code-review'] });
      manager.createDataset({ name: 'ds2', taskTypes: ['security-audit'] });
      manager.createDataset({ name: 'ds3', taskTypes: ['code-review', 'security-audit'] });

      const codeReviewDatasets = manager.listDatasets({ taskType: 'code-review' });
      expect(codeReviewDatasets.length).toBe(2);
    });

    it('should update example count', () => {
      const dataset = manager.createDataset({ name: 'test', taskTypes: ['general'] });
      manager.updateDatasetExampleCount(dataset.id, 100);
      
      const updated = manager.getDataset(dataset.id);
      expect(updated?.exampleCount).toBe(100);
    });

    it('should set file ID', () => {
      const dataset = manager.createDataset({ name: 'test', taskTypes: ['general'] });
      manager.setDatasetFileId(dataset.id, 'file-abc123');
      
      const updated = manager.getDataset(dataset.id);
      expect(updated?.fileId).toBe('file-abc123');
    });

    it('should delete dataset', () => {
      const dataset = manager.createDataset({ name: 'test', taskTypes: ['general'] });
      const deleted = manager.deleteDataset(dataset.id);
      
      expect(deleted).toBe(true);
      expect(manager.getDataset(dataset.id)).toBeUndefined();
    });
  });

  describe('Training Data Validation', () => {
    it('should validate training data with errors', () => {
      const result = manager.validateTrainingData([
        { messages: [] },
        { messages: [{ role: 'user', content: 'hello' }] },
      ]);

      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
    });

    it('should validate valid training data', () => {
      const examples = Array.from({ length: 15 }, (_, i) => ({
        messages: [
          { role: 'user' as const, content: `Question ${i}` },
          { role: 'assistant' as const, content: `Answer ${i}` },
        ],
      }));

      const result = manager.validateTrainingData(examples);
      expect(result.valid).toBe(true);
      expect(result.errors.length).toBe(0);
    });

    it('should format examples for OpenAI', () => {
      const examples = [
        {
          messages: [
            { role: 'user' as const, content: 'Hello' },
            { role: 'assistant' as const, content: 'Hi there!' },
          ],
        },
      ];

      const formatted = manager.formatExamplesForProvider(examples, 'openai');
      const parsed = JSON.parse(formatted);
      expect(parsed.messages).toHaveLength(2);
    });
  });

  describe('Fine-Tune Job Management', () => {
    it('should create a job', () => {
      const dataset = manager.createDataset({ name: 'test', taskTypes: ['general'] });
      manager.updateDatasetExampleCount(dataset.id, 50);

      const job = manager.createFineTuneJob({
        baseModel: 'gpt-4o-mini',
        datasetId: dataset.id,
        provider: 'openai',
        suffix: 'test-model',
      });

      expect(job.id).toMatch(/^ftjob_/);
      expect(job.status).toBe('pending');
      expect(job.baseModel).toBe('gpt-4o-mini');
    });

    it('should reject job with insufficient examples', () => {
      const dataset = manager.createDataset({ name: 'test', taskTypes: ['general'] });

      expect(() => {
        manager.createFineTuneJob({
          baseModel: 'gpt-4o-mini',
          datasetId: dataset.id,
          provider: 'openai',
        });
      }).toThrow('at least 10 examples');
    });

    it('should list jobs with filters', () => {
      const dataset = manager.createDataset({ name: 'test', taskTypes: ['general'] });
      manager.updateDatasetExampleCount(dataset.id, 50);

      manager.createFineTuneJob({ baseModel: 'gpt-4o', datasetId: dataset.id, provider: 'openai' });
      manager.createFineTuneJob({ baseModel: 'claude-3', datasetId: dataset.id, provider: 'anthropic' });

      const openaiJobs = manager.listJobs({ provider: 'openai' });
      expect(openaiJobs.length).toBe(1);
      expect(openaiJobs[0].provider).toBe('openai');
    });

    it('should start and progress a job', async () => {
      const dataset = manager.createDataset({ name: 'test', taskTypes: ['general'] });
      manager.updateDatasetExampleCount(dataset.id, 50);
      const job = manager.createFineTuneJob({
        baseModel: 'gpt-4o-mini',
        datasetId: dataset.id,
        provider: 'openai',
      });

      const started = await manager.startJob(job.id);
      expect(started.status).toBe('preparing');
      expect(started.startedAt).toBeDefined();
    });

    it('should cancel a job', async () => {
      const dataset = manager.createDataset({ name: 'test', taskTypes: ['general'] });
      manager.updateDatasetExampleCount(dataset.id, 50);
      const job = manager.createFineTuneJob({
        baseModel: 'gpt-4o-mini',
        datasetId: dataset.id,
        provider: 'openai',
      });

      await manager.startJob(job.id);
      const cancelled = manager.cancelJob(job.id);
      expect(cancelled.status).toBe('cancelled');
    });
  });

  describe('Model Management', () => {
    it('should register a model', () => {
      const model = manager.registerModel({
        name: 'my-fine-tuned-model',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:org:suffix:abc123',
        baseModel: 'gpt-4o',
        taskTypes: ['code-review'],
        version: '1.0.0',
      });

      expect(model.id).toMatch(/^ftmodel_/);
      expect(model.name).toBe('my-fine-tuned-model');
      expect(model.deploymentStatus).toBe('inactive');
    });

    it('should deploy a model', () => {
      const model = manager.registerModel({
        name: 'test-model',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:test',
        baseModel: 'gpt-4o',
        taskTypes: ['general'],
        version: '1.0.0',
      });

      const deployed = manager.deployModel(model.id);
      expect(deployed.deploymentStatus).toBe('active');
      expect(deployed.deployedAt).toBeDefined();
    });

    it('should deprecate a model', () => {
      const model = manager.registerModel({
        name: 'test-model',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:test',
        baseModel: 'gpt-4o',
        taskTypes: ['general'],
        version: '1.0.0',
      });

      manager.deployModel(model.id);
      const deprecated = manager.deprecateModel(model.id);
      expect(deprecated.deploymentStatus).toBe('deprecated');
    });

    it('should retire a model', () => {
      const model = manager.registerModel({
        name: 'test-model',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:test',
        baseModel: 'gpt-4o',
        taskTypes: ['general'],
        version: '1.0.0',
      });

      const retired = manager.retireModel(model.id);
      expect(retired.deploymentStatus).toBe('retired');
      expect(retired.retiredAt).toBeDefined();
    });

    it('should list models with filters', () => {
      manager.registerModel({
        name: 'model1',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:m1',
        baseModel: 'gpt-4o',
        taskTypes: ['code-review'],
        version: '1.0.0',
        tags: ['production'],
      });
      manager.registerModel({
        name: 'model2',
        provider: 'anthropic',
        providerModelId: 'ft:claude:m2',
        baseModel: 'claude-3',
        taskTypes: ['security-audit'],
        version: '1.0.0',
        tags: ['staging'],
      });

      expect(manager.listModels({ provider: 'openai' }).length).toBe(1);
      expect(manager.listModels({ taskType: 'security-audit' }).length).toBe(1);
      expect(manager.listModels({ tag: 'production' }).length).toBe(1);
    });

    it('should prevent deleting active model', () => {
      const model = manager.registerModel({
        name: 'test-model',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:test',
        baseModel: 'gpt-4o',
        taskTypes: ['general'],
        version: '1.0.0',
      });

      manager.deployModel(model.id);

      expect(() => manager.deleteModel(model.id)).toThrow('Cannot delete active model');
    });
  });

  describe('Supervisor Integration', () => {
    it('should create supervisor from deployed model', async () => {
      const model = manager.registerModel({
        name: 'test-supervisor',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:test',
        baseModel: 'gpt-4o',
        taskTypes: ['code-review'],
        version: '1.0.0',
      });

      manager.deployModel(model.id);
      const supervisor = manager.createSupervisorFromModel(model.id);

      expect(supervisor.name).toBe('test-supervisor');
      expect(supervisor.modelId).toBe(model.id);
      expect(supervisor.taskTypes).toEqual(['code-review']);
      expect(await supervisor.isAvailable()).toBe(true);
    });

    it('should reject creating supervisor from inactive model', () => {
      const model = manager.registerModel({
        name: 'test-model',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:test',
        baseModel: 'gpt-4o',
        taskTypes: ['general'],
        version: '1.0.0',
      });

      expect(() => manager.createSupervisorFromModel(model.id)).toThrow('not active');
    });

    it('should chat with supervisor', async () => {
      const model = manager.registerModel({
        name: 'test-supervisor',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:test',
        baseModel: 'gpt-4o',
        taskTypes: ['code-review'],
        version: '1.0.0',
      });

      manager.deployModel(model.id);
      const supervisor = manager.createSupervisorFromModel(model.id);

      const response = await supervisor.chat([
        { role: 'user', content: 'Review this code' },
      ]);

      expect(response).toContain('VOTE:');
      expect(response).toContain('CONFIDENCE:');
    });

    it('should list supervisors', () => {
      const model = manager.registerModel({
        name: 'test-supervisor',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:test',
        baseModel: 'gpt-4o',
        taskTypes: ['code-review'],
        version: '1.0.0',
      });

      manager.deployModel(model.id);
      manager.createSupervisorFromModel(model.id);

      const supervisors = manager.listSupervisors();
      expect(supervisors.length).toBe(1);
    });
  });

  describe('Performance Tracking', () => {
    it('should record debate results', () => {
      const model = manager.registerModel({
        name: 'test-model',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:test',
        baseModel: 'gpt-4o',
        taskTypes: ['code-review'],
        version: '1.0.0',
      });

      manager.recordDebateResult(model.id, {
        approved: true,
        confidence: 0.85,
        alignedWithConsensus: true,
        taskType: 'code-review',
      });

      const updated = manager.getModel(model.id);
      expect(updated?.performance?.totalDebates).toBe(1);
      expect(updated?.performance?.approvalRate).toBe(1);
      expect(updated?.performance?.averageConfidence).toBe(0.85);
    });

    it('should track performance by task type', () => {
      const model = manager.registerModel({
        name: 'test-model',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:test',
        baseModel: 'gpt-4o',
        taskTypes: ['code-review', 'security-audit'],
        version: '1.0.0',
      });

      manager.recordDebateResult(model.id, {
        approved: true,
        confidence: 0.9,
        alignedWithConsensus: true,
        taskType: 'code-review',
      });
      manager.recordDebateResult(model.id, {
        approved: false,
        confidence: 0.7,
        alignedWithConsensus: false,
        taskType: 'security-audit',
      });

      const updated = manager.getModel(model.id);
      expect(updated?.performance?.byTaskType?.['code-review']?.debates).toBe(1);
      expect(updated?.performance?.byTaskType?.['security-audit']?.debates).toBe(1);
    });

    it('should get performance history', () => {
      const model = manager.registerModel({
        name: 'test-model',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:test',
        baseModel: 'gpt-4o',
        taskTypes: ['general'],
        version: '1.0.0',
      });

      manager.recordDebateResult(model.id, { approved: true, confidence: 0.8, alignedWithConsensus: true });
      manager.recordDebateResult(model.id, { approved: true, confidence: 0.9, alignedWithConsensus: true });

      const history = manager.getPerformanceHistory(model.id);
      expect(history.length).toBe(2);
    });
  });

  describe('Model Comparison', () => {
    it('should compare two models', () => {
      const modelA = manager.registerModel({
        name: 'model-a',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:a',
        baseModel: 'gpt-4o',
        taskTypes: ['general'],
        version: '1.0.0',
      });
      const modelB = manager.registerModel({
        name: 'model-b',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:b',
        baseModel: 'gpt-4o',
        taskTypes: ['general'],
        version: '1.0.0',
      });

      for (let i = 0; i < 5; i++) {
        manager.recordDebateResult(modelA.id, { approved: true, confidence: 0.9, alignedWithConsensus: true });
        manager.recordDebateResult(modelB.id, { approved: true, confidence: 0.7, alignedWithConsensus: false });
      }

      const comparison = manager.compareModels(modelA.id, modelB.id);
      expect(comparison.recommendation).toBe('prefer-a');
      expect(comparison.metrics.confidenceDelta).toBeGreaterThan(0);
    });
  });

  describe('Best Model Selection', () => {
    it('should find best model for task type', () => {
      const model1 = manager.registerModel({
        name: 'model1',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:m1',
        baseModel: 'gpt-4o',
        taskTypes: ['code-review'],
        version: '1.0.0',
      });
      const model2 = manager.registerModel({
        name: 'model2',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:m2',
        baseModel: 'gpt-4o',
        taskTypes: ['code-review'],
        version: '1.0.0',
      });

      manager.deployModel(model1.id);
      manager.deployModel(model2.id);

      for (let i = 0; i < 10; i++) {
        manager.recordDebateResult(model1.id, { approved: true, confidence: 0.7, alignedWithConsensus: true, taskType: 'code-review' });
        manager.recordDebateResult(model2.id, { approved: true, confidence: 0.9, alignedWithConsensus: true, taskType: 'code-review' });
      }

      const best = manager.getBestModelForTask('code-review');
      expect(best?.id).toBe(model2.id);
    });

    it('should return undefined when no models match', () => {
      const best = manager.getBestModelForTask('security-audit');
      expect(best).toBeUndefined();
    });
  });

  describe('Export/Import', () => {
    it('should export model with performance history', () => {
      const model = manager.registerModel({
        name: 'export-test',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:export',
        baseModel: 'gpt-4o',
        taskTypes: ['general'],
        version: '1.0.0',
      });

      manager.recordDebateResult(model.id, { approved: true, confidence: 0.8, alignedWithConsensus: true });

      const exported = manager.exportModel(model.id);
      expect(exported.model.name).toBe('export-test');
      expect(exported.performance.length).toBe(1);
    });

    it('should import model', () => {
      const model = manager.registerModel({
        name: 'original',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:orig',
        baseModel: 'gpt-4o',
        taskTypes: ['general'],
        version: '1.0.0',
      });

      const exported = manager.exportModel(model.id);
      const imported = manager.importModel(exported);

      expect(imported.id).not.toBe(model.id);
      expect(imported.name).toBe('original');
      expect(imported.deploymentStatus).toBe('inactive');
    });
  });

  describe('Provider Configuration', () => {
    it('should set and get provider config', () => {
      manager.setProviderConfig('openai', { apiKey: 'sk-test', baseURL: 'https://api.openai.com' });
      
      const config = manager.getProviderConfig('openai');
      expect(config?.apiKey).toBe('sk-test');
      expect(config?.baseURL).toBe('https://api.openai.com');
    });
  });

  describe('Statistics', () => {
    it('should return correct statistics', () => {
      manager.createDataset({ name: 'ds1', taskTypes: ['general'] });
      manager.createDataset({ name: 'ds2', taskTypes: ['general'] });

      const ds = manager.createDataset({ name: 'ds3', taskTypes: ['general'] });
      manager.updateDatasetExampleCount(ds.id, 50);
      manager.createFineTuneJob({ baseModel: 'gpt-4o', datasetId: ds.id, provider: 'openai' });

      const model = manager.registerModel({
        name: 'stats-model',
        provider: 'openai',
        providerModelId: 'ft:gpt-4o:stats',
        baseModel: 'gpt-4o',
        taskTypes: ['general'],
        version: '1.0.0',
      });
      manager.deployModel(model.id);
      manager.createSupervisorFromModel(model.id);

      const stats = manager.getStatistics();
      expect(stats.totalDatasets).toBe(3);
      expect(stats.totalJobs).toBe(1);
      expect(stats.totalModels).toBe(1);
      expect(stats.activeSupervisors).toBe(1);
    });
  });
});
