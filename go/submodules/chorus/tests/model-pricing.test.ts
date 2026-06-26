/**
 * Tests for the OpenRouter pricing lookup. Mocks the in-memory cache via
 * the _testing seam so we don't hit network from CI. Covers the three
 * blockers caught in the chorus-self-review of v0.8.24:
 *   1. Gateway-prefixed ids fall back to bare-id lookup
 *   2. cachedInputTokens is included in cost (not silently dropped)
 *   3. ID normalization is dot/dash insensitive
 */
import { afterEach, describe, expect, it } from 'vitest';
import {
  _testing,
  getModelPricing,
  synthesizeCostUsd,
} from '@/lib/model-pricing';

const FIXTURE = {
  fetchedAt: Date.now(),
  prices: {
    // Bare keys — what fetchOpenRouterCatalog stores after stripping the
    // vendor prefix.
    'gemini-2-5-pro': { inputCostPerToken: 0.000002, outputCostPerToken: 0.000012 },
    'claude-opus-4-7': { inputCostPerToken: 0.000005, outputCostPerToken: 0.000025 },
    'kimi-k2-6': { inputCostPerToken: 0.0000005, outputCostPerToken: 0.000002 },
    'gpt-5-5': { inputCostPerToken: 0.000005, outputCostPerToken: 0.00003 },
    // Vendor-prefixed keys — what gets stored in addition to the bare form.
    'google/gemini-2-5-pro': { inputCostPerToken: 0.000002, outputCostPerToken: 0.000012 },
    'anthropic/claude-opus-4-7': { inputCostPerToken: 0.000005, outputCostPerToken: 0.000025 },
    'moonshotai/kimi-k2-6': { inputCostPerToken: 0.0000005, outputCostPerToken: 0.000002 },
  },
};

afterEach(() => {
  _testing.setMemoryCache(null);
});

describe('getModelPricing — id normalization', () => {
  it('matches dot-form id against dash-form catalog (claude-opus-4.7 → claude-opus-4-7)', async () => {
    _testing.setMemoryCache(FIXTURE);
    const price = await getModelPricing('claude-opus-4.7');
    expect(price?.inputCostPerToken).toBe(0.000005);
  });

  it('matches dash-form id against same dash-form catalog (chorus default)', async () => {
    _testing.setMemoryCache(FIXTURE);
    const price = await getModelPricing('claude-opus-4-7');
    expect(price?.inputCostPerToken).toBe(0.000005);
  });

  it('strips openrouter: prefix before lookup', async () => {
    _testing.setMemoryCache(FIXTURE);
    const price = await getModelPricing('openrouter:anthropic/claude-opus-4.7');
    expect(price?.inputCostPerToken).toBe(0.000005);
  });
});

describe('getModelPricing — gateway-prefix fallback (chorus-self-review blocker)', () => {
  it('finds bare-id pricing when full gateway-prefixed form misses', async () => {
    // opencode-go/kimi-k2.6 → normalized opencode-go/kimi-k2-6
    // (no exact key) → fallback to bare suffix kimi-k2-6 → hit.
    _testing.setMemoryCache(FIXTURE);
    const price = await getModelPricing('opencode-go/kimi-k2.6');
    expect(price?.inputCostPerToken).toBe(0.0000005);
  });

  it('also resolves the same model via the explicit moonshotai/ vendor id', async () => {
    _testing.setMemoryCache(FIXTURE);
    const price = await getModelPricing('moonshotai/kimi-k2.6');
    expect(price?.inputCostPerToken).toBe(0.0000005);
  });

  it('returns null when neither full nor bare form is in the catalog', async () => {
    _testing.setMemoryCache(FIXTURE);
    const price = await getModelPricing('made-up/unknown-model-99');
    expect(price).toBeNull();
  });
});

describe('synthesizeCostUsd — cached-input contract (chorus-self-review blocker)', () => {
  it('sums inputTokens + cachedInputTokens at the full input rate', async () => {
    _testing.setMemoryCache(FIXTURE);
    const cost = await synthesizeCostUsd('claude-opus-4-7', {
      inputTokens: 1000,
      cachedInputTokens: 9000,
      outputTokens: 500,
    });
    // (1000 + 9000) * 0.000005 + 500 * 0.000025 = 0.05 + 0.0125 = 0.0625
    expect(cost).toBeCloseTo(0.0625, 6);
  });

  it('treats cachedInputTokens as full-priced even when inputTokens is zero', async () => {
    _testing.setMemoryCache(FIXTURE);
    const cost = await synthesizeCostUsd('claude-opus-4-7', {
      cachedInputTokens: 1_000_000,
    });
    // 1_000_000 * 0.000005 = 5.0
    expect(cost).toBeCloseTo(5.0, 6);
  });

  it('returns undefined for unknown model (graceful no-cost rather than fake $0)', async () => {
    _testing.setMemoryCache(FIXTURE);
    const cost = await synthesizeCostUsd('unknown-model', {
      inputTokens: 100,
      outputTokens: 50,
    });
    expect(cost).toBeUndefined();
  });

  it('returns undefined for zero-token usage', async () => {
    _testing.setMemoryCache(FIXTURE);
    const cost = await synthesizeCostUsd('claude-opus-4-7', {});
    expect(cost).toBeUndefined();
  });

  it('returns undefined for empty model id', async () => {
    _testing.setMemoryCache(FIXTURE);
    const cost = await synthesizeCostUsd(undefined, {
      inputTokens: 100,
      outputTokens: 50,
    });
    expect(cost).toBeUndefined();
  });
});

describe('_testing.normalize', () => {
  it('lowercases and converts dots to dashes', () => {
    expect(_testing.normalize('Claude-Opus-4.7')).toBe('claude-opus-4-7');
  });
  it('idempotent on already-normalized form', () => {
    expect(_testing.normalize('claude-opus-4-7')).toBe('claude-opus-4-7');
  });
});
