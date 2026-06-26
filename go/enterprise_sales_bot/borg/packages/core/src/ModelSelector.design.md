# ModelSelector Design

## Purpose
Select the optimal AI model provider based on user quota, task complexity, and cost.

## Interface
```typescript
export interface ModelSelectionRequest {
  provider?: string; // Optional user preference
  taskComplexity?: 'low' | 'medium' | 'high';
}

export interface SelectedModel {
  provider: string; // 'anthropic' | 'openrouter' | 'google'
  modelId: string; // 'claude-3-5-sonnet' | 'gemini-1.5-pro'
  reason: string; // 'QUOTA_OK' | 'FALLBACK_FREE'
}

export class ModelSelector {
  async selectModel(req: ModelSelectionRequest): Promise<SelectedModel> {
    // Logic:
    // 1. Check if specific provider requested and has quota.
    // 2. If no quota, fallback to cheaper model.
    // 3. If complexity low, prefer cheaper model.
    return { provider: 'google', modelId: 'gemini-pro', reason: 'DEFAULT' };
  }
}
```
