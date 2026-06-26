import { EventEmitter } from 'events';
import type { Message, Supervisor, SupervisorConfig, TaskType } from './types.js';

// ============ Types ============

export type FineTuneProvider = 'openai' | 'anthropic' | 'custom';

export type FineTuneStatus = 
  | 'pending'      // Waiting to start
  | 'preparing'    // Preparing training data
  | 'training'     // Training in progress
  | 'completed'    // Training completed successfully
  | 'failed'       // Training failed
  | 'cancelled';   // Training was cancelled

export type ModelDeploymentStatus = 
  | 'inactive'     // Not deployed
  | 'deploying'    // Being deployed
  | 'active'       // Ready for use
  | 'deprecated'   // Still works but shouldn't be used
  | 'retired';     // No longer available

export interface TrainingDataset {
  id: string;
  name: string;
  description?: string;
  taskTypes: TaskType[];
  exampleCount: number;
  format: 'jsonl' | 'csv' | 'parquet';
  filePath?: string;
  fileId?: string; // Provider's file ID after upload
  createdAt: number;
  validationSplit?: number; // 0-1, portion for validation
  metadata?: Record<string, unknown>;
}

export interface TrainingExample {
  messages: Message[];
  taskType?: TaskType;
  expectedOutput?: string;
  weight?: number; // Importance weight
}

export interface FineTuneJobConfig {
  baseModel: string;
  datasetId: string;
  provider: FineTuneProvider;
  hyperparameters?: {
    epochs?: number;
    batchSize?: number;
    learningRateMultiplier?: number;
    promptLossWeight?: number;
  };
  suffix?: string; // Custom suffix for model name
  validationDatasetId?: string;
}

export interface FineTuneJob {
  id: string;
  provider: FineTuneProvider;
  providerJobId?: string; // ID from the provider
  baseModel: string;
  datasetId: string;
  status: FineTuneStatus;
  config: FineTuneJobConfig;
  resultModelId?: string; // The fine-tuned model ID
  createdAt: number;
  startedAt?: number;
  completedAt?: number;
  error?: string;
  metrics?: TrainingMetrics;
  events: FineTuneEvent[];
}

export interface FineTuneEvent {
  timestamp: number;
  type: 'created' | 'started' | 'progress' | 'completed' | 'failed' | 'cancelled';
  message: string;
  data?: Record<string, unknown>;
}

export interface TrainingMetrics {
  trainingLoss?: number;
  validationLoss?: number;
  trainingTokens?: number;
  epochs?: number;
  steps?: number;
  duration?: number; // ms
}

export interface FineTunedModel {
  id: string;
  name: string;
  description?: string;
  provider: FineTuneProvider;
  providerModelId: string; // The actual model ID to use in API calls
  baseModel: string;
  jobId: string;
  taskTypes: TaskType[]; // What this model is good at
  deploymentStatus: ModelDeploymentStatus;
  version: string;
  createdAt: number;
  deployedAt?: number;
  retiredAt?: number;
  performance?: ModelPerformance;
  config?: Partial<SupervisorConfig>;
  tags?: string[];
}

export interface ModelPerformance {
  totalDebates: number;
  approvalRate: number;
  averageConfidence: number;
  consensusAlignment: number; // How often it aligns with final decision
  responseTimeMs: number;
  errorRate: number;
  comparisonToBase?: {
    approvalRateDelta: number;
    confidenceDelta: number;
    alignmentDelta: number;
  };
  byTaskType?: Record<TaskType, {
    debates: number;
    approvalRate: number;
    confidence: number;
    alignment: number;
  }>;
}

export interface ModelComparison {
  modelA: string;
  modelB: string;
  taskType?: TaskType;
  metrics: {
    approvalRateDelta: number;
    confidenceDelta: number;
    alignmentDelta: number;
    responseTimeDelta: number;
  };
  recommendation: 'prefer-a' | 'prefer-b' | 'equivalent';
  reasoning: string;
}

export interface FineTunedSupervisor extends Supervisor {
  name: string;
  provider: string;
  modelId: string;
  version: string;
  taskTypes: TaskType[];
  chat(messages: Message[]): Promise<string>;
  isAvailable(): Promise<boolean>;
}

// ============ Fine-Tuned Model Manager Service ============

export class FineTunedModelManagerService extends EventEmitter {
  private datasets: Map<string, TrainingDataset> = new Map();
  private jobs: Map<string, FineTuneJob> = new Map();
  private models: Map<string, FineTunedModel> = new Map();
  private supervisors: Map<string, FineTunedSupervisor> = new Map();
  private performanceHistory: Map<string, Array<{ timestamp: number; metrics: Partial<ModelPerformance> }>> = new Map();
  
  // Provider API keys (should be set via environment or config)
  private providerConfigs: Map<FineTuneProvider, { apiKey?: string; baseURL?: string }> = new Map();

  constructor() {
    super();
  }

  // ============ Dataset Management ============

  createDataset(params: {
    name: string;
    description?: string;
    taskTypes: TaskType[];
    format?: 'jsonl' | 'csv' | 'parquet';
    validationSplit?: number;
    metadata?: Record<string, unknown>;
  }): TrainingDataset {
    const id = `dataset_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    const dataset: TrainingDataset = {
      id,
      name: params.name,
      description: params.description,
      taskTypes: params.taskTypes,
      exampleCount: 0,
      format: params.format ?? 'jsonl',
      createdAt: Date.now(),
      validationSplit: params.validationSplit ?? 0.1,
      metadata: params.metadata,
    };

    this.datasets.set(id, dataset);
    this.emit('dataset:created', dataset);
    return dataset;
  }

  getDataset(id: string): TrainingDataset | undefined {
    return this.datasets.get(id);
  }

  listDatasets(filter?: { taskType?: TaskType }): TrainingDataset[] {
    let datasets = Array.from(this.datasets.values());
    
    if (filter?.taskType) {
      datasets = datasets.filter(d => d.taskTypes.includes(filter.taskType!));
    }
    
    return datasets.sort((a, b) => b.createdAt - a.createdAt);
  }

  updateDatasetExampleCount(id: string, count: number): void {
    const dataset = this.datasets.get(id);
    if (dataset) {
      dataset.exampleCount = count;
      this.emit('dataset:updated', dataset);
    }
  }

  setDatasetFileId(id: string, fileId: string): void {
    const dataset = this.datasets.get(id);
    if (dataset) {
      dataset.fileId = fileId;
      this.emit('dataset:updated', dataset);
    }
  }

  deleteDataset(id: string): boolean {
    const dataset = this.datasets.get(id);
    if (!dataset) return false;

    const jobsUsingDataset = Array.from(this.jobs.values())
      .filter(j => j.datasetId === id && !['completed', 'failed', 'cancelled'].includes(j.status));
    
    if (jobsUsingDataset.length > 0) {
      throw new Error(`Dataset is in use by ${jobsUsingDataset.length} active job(s)`);
    }

    this.datasets.delete(id);
    this.emit('dataset:deleted', { id });
    return true;
  }

  // ============ Training Examples ============

  formatExamplesForProvider(examples: TrainingExample[], provider: FineTuneProvider): string {
    switch (provider) {
      case 'openai':
        return examples.map(ex => JSON.stringify({
          messages: ex.messages.map(m => ({
            role: m.role,
            content: m.content,
          })),
        })).join('\n');
      
      case 'anthropic':
        return examples.map(ex => JSON.stringify({
          prompt: ex.messages.filter(m => m.role !== 'assistant').map(m => m.content).join('\n'),
          completion: ex.messages.find(m => m.role === 'assistant')?.content ?? '',
        })).join('\n');
      
      case 'custom':
        return examples.map(ex => JSON.stringify(ex)).join('\n');
      
      default:
        return examples.map(ex => JSON.stringify(ex)).join('\n');
    }
  }

  validateTrainingData(examples: TrainingExample[]): { valid: boolean; errors: string[] } {
    const errors: string[] = [];

    if (examples.length < 10) {
      errors.push('Minimum 10 examples required for fine-tuning');
    }

    for (let i = 0; i < examples.length; i++) {
      const ex = examples[i];
      if (!ex.messages || ex.messages.length === 0) {
        errors.push(`Example ${i}: No messages provided`);
      }
      
      const hasAssistant = ex.messages.some(m => m.role === 'assistant');
      if (!hasAssistant) {
        errors.push(`Example ${i}: Must include at least one assistant message`);
      }
    }

    return { valid: errors.length === 0, errors };
  }

  // ============ Fine-Tune Job Management ============

  createFineTuneJob(config: FineTuneJobConfig): FineTuneJob {
    const dataset = this.datasets.get(config.datasetId);
    if (!dataset) {
      throw new Error(`Dataset not found: ${config.datasetId}`);
    }

    if (dataset.exampleCount < 10) {
      throw new Error('Dataset must have at least 10 examples');
    }

    const id = `ftjob_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    const job: FineTuneJob = {
      id,
      provider: config.provider,
      baseModel: config.baseModel,
      datasetId: config.datasetId,
      status: 'pending',
      config,
      createdAt: Date.now(),
      events: [{
        timestamp: Date.now(),
        type: 'created',
        message: `Fine-tune job created for ${config.baseModel}`,
      }],
    };

    this.jobs.set(id, job);
    this.emit('job:created', job);
    return job;
  }

  getJob(id: string): FineTuneJob | undefined {
    return this.jobs.get(id);
  }

  listJobs(filter?: { status?: FineTuneStatus; provider?: FineTuneProvider }): FineTuneJob[] {
    let jobs = Array.from(this.jobs.values());
    
    if (filter?.status) {
      jobs = jobs.filter(j => j.status === filter.status);
    }
    if (filter?.provider) {
      jobs = jobs.filter(j => j.provider === filter.provider);
    }
    
    return jobs.sort((a, b) => b.createdAt - a.createdAt);
  }

  async startJob(id: string): Promise<FineTuneJob> {
    const job = this.jobs.get(id);
    if (!job) {
      throw new Error(`Job not found: ${id}`);
    }

    if (job.status !== 'pending') {
      throw new Error(`Job is not in pending state: ${job.status}`);
    }

    job.status = 'preparing';
    job.startedAt = Date.now();
    job.events.push({
      timestamp: Date.now(),
      type: 'started',
      message: 'Job started, preparing training data',
    });

    this.emit('job:started', job);
    setTimeout(() => this.simulateJobProgress(id), 1000);

    return job;
  }

  private simulateJobProgress(id: string): void {
    const job = this.jobs.get(id);
    if (!job || job.status === 'cancelled') return;

    if (job.status === 'preparing') {
      job.status = 'training';
      job.providerJobId = `ft-${Math.random().toString(36).substr(2, 16)}`;
      job.events.push({
        timestamp: Date.now(),
        type: 'progress',
        message: 'Training started',
        data: { providerJobId: job.providerJobId },
      });
      this.emit('job:progress', job);
      setTimeout(() => this.simulateJobProgress(id), 2000);
    } else if (job.status === 'training') {
      job.status = 'completed';
      job.completedAt = Date.now();
      job.resultModelId = `ft:${job.baseModel}:${job.config.suffix || 'custom'}:${Math.random().toString(36).substr(2, 8)}`;
      job.metrics = {
        trainingLoss: 0.1 + Math.random() * 0.2,
        validationLoss: 0.15 + Math.random() * 0.25,
        trainingTokens: 50000 + Math.floor(Math.random() * 100000),
        epochs: job.config.hyperparameters?.epochs ?? 3,
        steps: 1000 + Math.floor(Math.random() * 2000),
        duration: Date.now() - (job.startedAt ?? job.createdAt),
      };
      job.events.push({
        timestamp: Date.now(),
        type: 'completed',
        message: `Training completed. Model: ${job.resultModelId}`,
        data: { metrics: job.metrics },
      });
      this.emit('job:completed', job);
      this.registerModelFromJob(job);
    }
  }

  cancelJob(id: string): FineTuneJob {
    const job = this.jobs.get(id);
    if (!job) {
      throw new Error(`Job not found: ${id}`);
    }

    if (['completed', 'failed', 'cancelled'].includes(job.status)) {
      throw new Error(`Cannot cancel job in ${job.status} state`);
    }

    job.status = 'cancelled';
    job.events.push({
      timestamp: Date.now(),
      type: 'cancelled',
      message: 'Job cancelled by user',
    });

    this.emit('job:cancelled', job);
    return job;
  }

  // ============ Model Management ============

  private registerModelFromJob(job: FineTuneJob): FineTunedModel | undefined {
    if (!job.resultModelId) return undefined;

    const dataset = this.datasets.get(job.datasetId);
    
    return this.registerModel({
      name: job.config.suffix || `${job.baseModel}-finetuned`,
      provider: job.provider,
      providerModelId: job.resultModelId,
      baseModel: job.baseModel,
      jobId: job.id,
      taskTypes: dataset?.taskTypes ?? ['general'],
      version: '1.0.0',
    });
  }

  registerModel(params: {
    name: string;
    description?: string;
    provider: FineTuneProvider;
    providerModelId: string;
    baseModel: string;
    jobId?: string;
    taskTypes: TaskType[];
    version: string;
    config?: Partial<SupervisorConfig>;
    tags?: string[];
  }): FineTunedModel {
    const id = `ftmodel_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    const model: FineTunedModel = {
      id,
      name: params.name,
      description: params.description,
      provider: params.provider,
      providerModelId: params.providerModelId,
      baseModel: params.baseModel,
      jobId: params.jobId ?? 'external',
      taskTypes: params.taskTypes,
      deploymentStatus: 'inactive',
      version: params.version,
      createdAt: Date.now(),
      config: params.config,
      tags: params.tags,
      performance: {
        totalDebates: 0,
        approvalRate: 0,
        averageConfidence: 0,
        consensusAlignment: 0,
        responseTimeMs: 0,
        errorRate: 0,
      },
    };

    this.models.set(id, model);
    this.emit('model:registered', model);
    return model;
  }

  getModel(id: string): FineTunedModel | undefined {
    return this.models.get(id);
  }

  getModelByProviderModelId(providerModelId: string): FineTunedModel | undefined {
    return Array.from(this.models.values()).find(m => m.providerModelId === providerModelId);
  }

  listModels(filter?: { 
    status?: ModelDeploymentStatus; 
    provider?: FineTuneProvider;
    taskType?: TaskType;
    tag?: string;
  }): FineTunedModel[] {
    let models = Array.from(this.models.values());
    
    if (filter?.status) {
      models = models.filter(m => m.deploymentStatus === filter.status);
    }
    if (filter?.provider) {
      models = models.filter(m => m.provider === filter.provider);
    }
    if (filter?.taskType) {
      models = models.filter(m => m.taskTypes.includes(filter.taskType!));
    }
    if (filter?.tag) {
      models = models.filter(m => m.tags?.includes(filter.tag!));
    }
    
    return models.sort((a, b) => b.createdAt - a.createdAt);
  }

  deployModel(id: string): FineTunedModel {
    const model = this.models.get(id);
    if (!model) {
      throw new Error(`Model not found: ${id}`);
    }

    if (model.deploymentStatus === 'active') {
      return model;
    }

    model.deploymentStatus = 'deploying';
    this.emit('model:deploying', model);
    model.deploymentStatus = 'active';
    model.deployedAt = Date.now();
    this.emit('model:deployed', model);

    return model;
  }

  deprecateModel(id: string): FineTunedModel {
    const model = this.models.get(id);
    if (!model) {
      throw new Error(`Model not found: ${id}`);
    }

    model.deploymentStatus = 'deprecated';
    this.emit('model:deprecated', model);
    return model;
  }

  retireModel(id: string): FineTunedModel {
    const model = this.models.get(id);
    if (!model) {
      throw new Error(`Model not found: ${id}`);
    }

    model.deploymentStatus = 'retired';
    model.retiredAt = Date.now();
    this.supervisors.delete(id);
    
    this.emit('model:retired', model);
    return model;
  }

  deleteModel(id: string): boolean {
    const model = this.models.get(id);
    if (!model) return false;

    if (model.deploymentStatus === 'active') {
      throw new Error('Cannot delete active model. Retire it first.');
    }

    this.supervisors.delete(id);
    this.models.delete(id);
    this.performanceHistory.delete(id);
    this.emit('model:deleted', { id });
    return true;
  }

  // ============ Supervisor Integration ============

  createSupervisorFromModel(modelId: string, overrides?: Partial<SupervisorConfig>): FineTunedSupervisor {
    const model = this.models.get(modelId);
    if (!model) {
      throw new Error(`Model not found: ${modelId}`);
    }

    if (model.deploymentStatus !== 'active') {
      throw new Error(`Model is not active: ${model.deploymentStatus}`);
    }

    const providerConfig = this.providerConfigs.get(model.provider);
    
    const supervisor: FineTunedSupervisor = {
      name: overrides?.name ?? model.name,
      provider: model.provider,
      modelId: model.id,
      version: model.version,
      taskTypes: model.taskTypes,
      
      chat: async (messages: Message[]): Promise<string> => {
        const startTime = Date.now();
        try {
          const response = await this.simulateModelChat(model, messages);
          this.recordModelUsage(modelId, {
            responseTimeMs: Date.now() - startTime,
            success: true,
          });
          
          return response;
        } catch (error) {
          this.recordModelUsage(modelId, {
            responseTimeMs: Date.now() - startTime,
            success: false,
          });
          throw error;
        }
      },
      
      isAvailable: async (): Promise<boolean> => {
        return model.deploymentStatus === 'active';
      },
    };

    this.supervisors.set(modelId, supervisor);
    this.emit('supervisor:created', { modelId, supervisor: { name: supervisor.name } });
    return supervisor;
  }

  private async simulateModelChat(model: FineTunedModel, messages: Message[]): Promise<string> {
    await new Promise(resolve => setTimeout(resolve, 100 + Math.random() * 200));
    const taskContext = messages.map(m => m.content).join('\n');
    const isApproval = Math.random() > 0.3;
    const confidence = 0.7 + Math.random() * 0.3;
    
    return `Based on my analysis as ${model.name} (fine-tuned from ${model.baseModel}):

VOTE: ${isApproval ? 'APPROVE' : 'REJECT'}
CONFIDENCE: ${confidence.toFixed(2)}
REASONING: This task ${isApproval ? 'aligns well with' : 'has concerns regarding'} best practices for ${model.taskTypes.join(', ')} tasks.`;
  }

  getSupervisor(modelId: string): FineTunedSupervisor | undefined {
    return this.supervisors.get(modelId);
  }

  listSupervisors(): FineTunedSupervisor[] {
    return Array.from(this.supervisors.values());
  }

  getSupervisorConfig(modelId: string): SupervisorConfig | undefined {
    const model = this.models.get(modelId);
    const supervisor = this.supervisors.get(modelId);
    if (!model || !supervisor) return undefined;

    const providerConfig = this.providerConfigs.get(model.provider);

    return {
      name: supervisor.name,
      provider: model.provider === 'custom' ? 'custom' : model.provider as SupervisorConfig['provider'],
      model: model.providerModelId,
      apiKey: providerConfig?.apiKey,
      baseURL: providerConfig?.baseURL,
      ...model.config,
    };
  }

  // ============ Performance Tracking ============

  private recordModelUsage(modelId: string, usage: { responseTimeMs: number; success: boolean }): void {
    const model = this.models.get(modelId);
    if (!model || !model.performance) return;
    const perf = model.performance;
    const n = perf.totalDebates;
    
    perf.responseTimeMs = (perf.responseTimeMs * n + usage.responseTimeMs) / (n + 1);
    if (!usage.success) {
      perf.errorRate = (perf.errorRate * n + 1) / (n + 1);
    } else {
      perf.errorRate = (perf.errorRate * n) / (n + 1);
    }
  }

  recordDebateResult(modelId: string, result: {
    approved: boolean;
    confidence: number;
    alignedWithConsensus: boolean;
    taskType?: TaskType;
  }): void {
    const model = this.models.get(modelId);
    if (!model || !model.performance) return;

    const perf = model.performance;
    const n = perf.totalDebates;
    
    perf.totalDebates++;
    perf.approvalRate = (perf.approvalRate * n + (result.approved ? 1 : 0)) / (n + 1);
    perf.averageConfidence = (perf.averageConfidence * n + result.confidence) / (n + 1);
    perf.consensusAlignment = (perf.consensusAlignment * n + (result.alignedWithConsensus ? 1 : 0)) / (n + 1);

    if (result.taskType) {
      if (!perf.byTaskType) {
        perf.byTaskType = {} as Record<TaskType, { debates: number; approvalRate: number; confidence: number; alignment: number }>;
      }
      
      const taskPerf = perf.byTaskType[result.taskType] ?? { debates: 0, approvalRate: 0, confidence: 0, alignment: 0 };
      const tn = taskPerf.debates;
      
      taskPerf.debates++;
      taskPerf.approvalRate = (taskPerf.approvalRate * tn + (result.approved ? 1 : 0)) / (tn + 1);
      taskPerf.confidence = (taskPerf.confidence * tn + result.confidence) / (tn + 1);
      taskPerf.alignment = (taskPerf.alignment * tn + (result.alignedWithConsensus ? 1 : 0)) / (tn + 1);
      
      perf.byTaskType[result.taskType] = taskPerf;
    }

    const history = this.performanceHistory.get(modelId) ?? [];
    history.push({ timestamp: Date.now(), metrics: { ...perf } });
    if (history.length > 1000) {
      history.shift();
    }
    this.performanceHistory.set(modelId, history);

    this.emit('performance:updated', { modelId, performance: perf });
  }

  getPerformanceHistory(modelId: string, since?: number): Array<{ timestamp: number; metrics: Partial<ModelPerformance> }> {
    const history = this.performanceHistory.get(modelId) ?? [];
    if (since) {
      return history.filter(h => h.timestamp >= since);
    }
    return history;
  }

  compareModels(modelAId: string, modelBId: string, taskType?: TaskType): ModelComparison {
    const modelA = this.models.get(modelAId);
    const modelB = this.models.get(modelBId);
    
    if (!modelA || !modelB) {
      throw new Error('One or both models not found');
    }

    const perfA = taskType && modelA.performance?.byTaskType?.[taskType] 
      ? modelA.performance.byTaskType[taskType]
      : modelA.performance;
    
    const perfB = taskType && modelB.performance?.byTaskType?.[taskType]
      ? modelB.performance.byTaskType[taskType]
      : modelB.performance;

    if (!perfA || !perfB) {
      throw new Error('Performance data not available for comparison');
    }

    const metrics = {
      approvalRateDelta: (perfA.approvalRate ?? 0) - (perfB.approvalRate ?? 0),
      confidenceDelta: ((perfA as ModelPerformance).averageConfidence ?? (perfA as { confidence: number }).confidence ?? 0) - 
                       ((perfB as ModelPerformance).averageConfidence ?? (perfB as { confidence: number }).confidence ?? 0),
      alignmentDelta: ((perfA as ModelPerformance).consensusAlignment ?? (perfA as { alignment: number }).alignment ?? 0) - 
                      ((perfB as ModelPerformance).consensusAlignment ?? (perfB as { alignment: number }).alignment ?? 0),
      responseTimeDelta: ((perfA as ModelPerformance).responseTimeMs ?? 0) - ((perfB as ModelPerformance).responseTimeMs ?? 0),
    };

    let score = 0;
    if (metrics.alignmentDelta > 0.05) score++;
    if (metrics.alignmentDelta < -0.05) score--;
    if (metrics.confidenceDelta > 0.05) score++;
    if (metrics.confidenceDelta < -0.05) score--;
    if (metrics.responseTimeDelta < -100) score++;
    if (metrics.responseTimeDelta > 100) score--;

    let recommendation: 'prefer-a' | 'prefer-b' | 'equivalent';
    let reasoning: string;
    
    if (score > 0) {
      recommendation = 'prefer-a';
      reasoning = `${modelA.name} shows better overall performance`;
    } else if (score < 0) {
      recommendation = 'prefer-b';
      reasoning = `${modelB.name} shows better overall performance`;
    } else {
      recommendation = 'equivalent';
      reasoning = 'Both models perform similarly';
    }

    if (taskType) {
      reasoning += ` for ${taskType} tasks`;
    }

    return {
      modelA: modelAId,
      modelB: modelBId,
      taskType,
      metrics,
      recommendation,
      reasoning,
    };
  }

  // ============ Provider Configuration ============

  setProviderConfig(provider: FineTuneProvider, config: { apiKey?: string; baseURL?: string }): void {
    this.providerConfigs.set(provider, config);
  }

  getProviderConfig(provider: FineTuneProvider): { apiKey?: string; baseURL?: string } | undefined {
    return this.providerConfigs.get(provider);
  }

  // ============ Best Model Selection ============

  getBestModelForTask(taskType: TaskType): FineTunedModel | undefined {
    const activeModels = this.listModels({ status: 'active', taskType });
    
    if (activeModels.length === 0) return undefined;

    return activeModels.sort((a, b) => {
      const perfA = a.performance?.byTaskType?.[taskType] ?? a.performance;
      const perfB = b.performance?.byTaskType?.[taskType] ?? b.performance;
      
      const alignA = (perfA as ModelPerformance)?.consensusAlignment ?? (perfA as { alignment?: number })?.alignment ?? 0;
      const alignB = (perfB as ModelPerformance)?.consensusAlignment ?? (perfB as { alignment?: number })?.alignment ?? 0;
      
      if (Math.abs(alignA - alignB) > 0.05) {
        return alignB - alignA;
      }
      
      const confA = (perfA as ModelPerformance)?.averageConfidence ?? (perfA as { confidence?: number })?.confidence ?? 0;
      const confB = (perfB as ModelPerformance)?.averageConfidence ?? (perfB as { confidence?: number })?.confidence ?? 0;
      
      return confB - confA;
    })[0];
  }

  // ============ Export/Import ============

  exportModel(id: string): { model: FineTunedModel; performance: Array<{ timestamp: number; metrics: Partial<ModelPerformance> }> } {
    const model = this.models.get(id);
    if (!model) {
      throw new Error(`Model not found: ${id}`);
    }

    return {
      model: { ...model },
      performance: this.getPerformanceHistory(id),
    };
  }

  importModel(data: { model: FineTunedModel; performance?: Array<{ timestamp: number; metrics: Partial<ModelPerformance> }> }): FineTunedModel {
    const id = `ftmodel_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    const model: FineTunedModel = {
      ...data.model,
      id,
      createdAt: Date.now(),
      deploymentStatus: 'inactive',
    };

    this.models.set(id, model);
    
    if (data.performance) {
      this.performanceHistory.set(id, data.performance);
    }

    this.emit('model:imported', model);
    return model;
  }

  // ============ Statistics ============

  getStatistics(): {
    totalDatasets: number;
    totalJobs: number;
    jobsByStatus: Record<FineTuneStatus, number>;
    totalModels: number;
    modelsByStatus: Record<ModelDeploymentStatus, number>;
    activeSupervisors: number;
    totalDebatesTracked: number;
    averageAlignment: number;
  } {
    const jobs = Array.from(this.jobs.values());
    const models = Array.from(this.models.values());
    
    const jobsByStatus = jobs.reduce((acc, j) => {
      acc[j.status] = (acc[j.status] || 0) + 1;
      return acc;
    }, {} as Record<FineTuneStatus, number>);

    const modelsByStatus = models.reduce((acc, m) => {
      acc[m.deploymentStatus] = (acc[m.deploymentStatus] || 0) + 1;
      return acc;
    }, {} as Record<ModelDeploymentStatus, number>);

    const activeModels = models.filter(m => m.deploymentStatus === 'active' && m.performance);
    const totalDebates = activeModels.reduce((sum, m) => sum + (m.performance?.totalDebates ?? 0), 0);
    const avgAlignment = activeModels.length > 0
      ? activeModels.reduce((sum, m) => sum + (m.performance?.consensusAlignment ?? 0), 0) / activeModels.length
      : 0;

    return {
      totalDatasets: this.datasets.size,
      totalJobs: this.jobs.size,
      jobsByStatus,
      totalModels: this.models.size,
      modelsByStatus,
      activeSupervisors: this.supervisors.size,
      totalDebatesTracked: totalDebates,
      averageAlignment: avgAlignment,
    };
  }
}

// Export singleton instance
export const fineTunedModelManager = new FineTunedModelManagerService();
