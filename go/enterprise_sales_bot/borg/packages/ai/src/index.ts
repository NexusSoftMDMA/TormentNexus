// @tormentnexus/ai — type-safe stub package
// Provides type declarations for consumers while deferring real implementations
// to the TypeScript core runtime.

// ── Type declarations ──────────────────────────────────────────────────────

export interface ChatMessage {
  role: 'system' | 'user' | 'assistant' | 'tool';
  content: string;
  name?: string;
  toolCallId?: string;
  toolCalls?: any[];
}

export interface ModelSelectionRequest {
  taskComplexity?: 'low' | 'medium' | 'high';
  taskType?: string;
  routingTaskType?: string;
  provider?: string;
}

export interface SelectedModel {
  provider: string;
  modelId: string;
  reason?: string;
}

export interface QuotaConfig {
  providers?: Record<string, any>;
  dailyBudgetUsd?: number;
  monthlyBudgetUsd?: number;
  providerLimits?: Record<string, any>;
  preEmptiveSwitchThreshold?: number;
}

export interface IAgent {
  id: string;
  name: string;
  role: string;
  start(): Promise<void>;
  stop(): Promise<void>;
  getStatus?(): any;
  getCapabilities?(): string[];
  handleMessage?(message: any): Promise<any>;
}

// ── Class stubs ────────────────────────────────────────────────────────────

export class ModelSelector {
  async selectModel(_req?: Partial<ModelSelectionRequest & Record<string, any>>): Promise<SelectedModel> {
    return { provider: 'stub', modelId: 'stub-model', reason: '@tormentnexus/ai stub' };
  }
}

export class LLMService {
  modelSelector: ModelSelector = new ModelSelector();
  constructor(modelSelector?: ModelSelector, _opts?: any) {
    if (modelSelector) this.modelSelector = modelSelector;
  }
  async generateText(
    _provider: string,
    _modelId: string,
    _systemPrompt: string,
    _userPrompt: string,
    _opts?: any,
  ): Promise<{ content: string; usage?: { inputTokens: number; outputTokens: number } }> {
    return { content: '[stub response from @tormentnexus/ai]' };
  }
}

export class QuotaService {
  protected configState: QuotaConfig = {
    providers: {},
    dailyBudgetUsd: 5,
    monthlyBudgetUsd: 100,
    providerLimits: {},
  };
  setConfig(config: QuotaConfig) { Object.assign(this.configState, config); }
  getConfig(): QuotaConfig { return this.configState; }
  markAuthRevoked(_provider: string, _reason?: string): void {}
}

export const DEFAULT_OPENROUTER_FREE_MODEL = 'google/gemma-3-27b-it:free';
