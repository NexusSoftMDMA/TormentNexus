import { describe, expect, it } from 'vitest';

import {
  getBillingUsageSummary,
  getDefaultRoutingStrategy,
  getFallbackTaskType,
  normalizeBillingPricingModels,
  normalizeBillingQuotaRows,
  normalizeFallbackChain,
  normalizeTaskRoutingRules,
} from './billing-page-normalizers';

describe('billing page normalizers', () => {
  it('falls back to safe numeric usage defaults when payload is malformed', () => {
    expect(getBillingUsageSummary({ usage: { currentMonth: 'bad', limit: null, breakdown: { bad: true } } })).toEqual({
      currentMonth: 0,
      limit: 0,
      breakdown: [],
    });
  });

  it('normalizes fallback chain rows and task type safely', () => {
    const fallback = {
      selectedTaskType: 'supervisor',
      chain: [
        null,
        { provider: 'openai', priority: 'oops', reason: '' },
        { provider: 'anthropic', priority: 2, model: 'claude', reason: 'quota_headroom' },
      ],
    };

    expect(getFallbackTaskType(fallback, 'general')).toBe('supervisor');
    expect(getFallbackTaskType({ selectedTaskType: 'bad' }, 'general')).toBe('general');

    expect(normalizeFallbackChain(fallback)).toEqual([
      {
        priority: 2,
        provider: 'openai',
        reason: 'ranked fallback',
      },
      {
        priority: 2,
        provider: 'anthropic',
        model: 'claude',
        reason: 'quota_headroom',
      },
    ]);
  });

  it('normalizes routing strategy and rules from unknown payload shapes', () => {
    expect(getDefaultRoutingStrategy({ defaultStrategy: 'cheapest' })).toBe('cheapest');
    expect(getDefaultRoutingStrategy({ defaultStrategy: 'invalid' })).toBe('best');

    expect(normalizeTaskRoutingRules({
      rules: [
        {
          taskType: 'coding',
          strategy: 'round-robin',
          fallbackPreview: [{ provider: 'openai', reason: 'quality' }, { provider: 1 }],
        },
        {
          taskType: 'invalid',
          strategy: 'invalid',
          fallbackPreview: 'bad',
        },
      ],
    })).toEqual([
      {
        taskType: 'coding',
        strategy: 'round-robin',
        fallbackPreview: [
          { provider: 'openai', reason: 'quality' },
          { provider: 'provider-2' },
        ],
      },
      {
        taskType: 'general',
        strategy: 'best',
        fallbackPreview: [],
      },
    ]);
  });

  it('normalizes provider quota and pricing rows used by billing tables', () => {
    expect(normalizeBillingQuotaRows([
      null,
      {
        provider: 'openai',
        name: '',
        configured: true,
        authenticated: false,
        authMethod: null,
        tier: 123,
        limit: 'bad',
        used: 'bad',
        rateLimitRpm: null,
        availability: null,
        lastError: 42,
      },
    ])).toEqual([
      {
        provider: 'openai',
        name: 'openai',
        configured: true,
        authenticated: false,
        authMethod: 'none',
        tier: 'standard',
        limit: 0,
        used: 0,
        rateLimitRpm: null,
        availability: 'unknown',
        lastError: null,
      },
    ]);

    expect(normalizeBillingPricingModels({
      models: [
        null,
        {
          id: 'gpt-4.1',
          contextWindow: 'bad',
          inputPrice: null,
          inputPricePer1k: 'bad',
          outputPricePer1k: 1.23,
          recommended: 'bad',
        },
      ],
    })).toEqual([
      {
        id: 'gpt-4.1',
        contextWindow: 0,
        inputPrice: null,
        inputPricePer1k: 0,
        outputPricePer1k: 1.23,
        recommended: false,
      },
    ]);
  });
});
