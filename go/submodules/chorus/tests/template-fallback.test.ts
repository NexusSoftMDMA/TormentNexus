/**
 * Tests the template-level fallback chain composition.
 *
 * The runner doesn't need its own retry loop — `runWithChainFallback`
 * walks a (lineage, model) chain and falls through on null.
 * `buildSlotFallbackChain`'s job is to compose that chain correctly:
 * append template-level fallbacks (same- AND cross-lineage as of v0.8)
 * onto the slot's per-slot chain, deduped against every active
 * (lineage, model) in the phase.
 *
 * Critical case from the user spec (2026-05-03):
 *   reviewers=[kimi, deepseek] (both opencode lineage)
 *   template.fallback=[kimi]
 *   When deepseek fails → must NOT spawn a second kimi reviewer.
 */
import { describe, it, expect } from 'vitest';
import { buildSlotFallbackChain } from '../src/daemon/runner/template-fallback';

describe('buildSlotFallbackChain', () => {
  it('returns slot.models as same-lineage entries when no template fallback exists', () => {
    const slot = { lineage: 'opencode', models: ['kimi-k2.6'] };
    const chain = buildSlotFallbackChain(slot, [slot], undefined);
    expect(chain).toEqual([{ lineage: 'opencode', model: 'kimi-k2.6' }]);
  });

  it('returns slot.models unchanged when template fallback is empty array', () => {
    const slot = { lineage: 'opencode', models: ['kimi-k2.6'] };
    const chain = buildSlotFallbackChain(slot, [slot], []);
    expect(chain).toEqual([{ lineage: 'opencode', model: 'kimi-k2.6' }]);
  });

  it('emits a single undefined-model entry when slot has no models', () => {
    // Slot with empty models gets one attempt with the lineage default.
    const slot = { lineage: 'anthropic', models: [] };
    const chain = buildSlotFallbackChain(slot, [slot], undefined);
    expect(chain).toEqual([{ lineage: 'anthropic', model: undefined }]);
  });

  it('appends same-lineage template fallbacks onto the chain', () => {
    const slot = { lineage: 'opencode', models: ['deepseek-v4-pro'] };
    const fallback = [
      { lineage: 'opencode', models: ['kimi-k2.6'] },
      { lineage: 'opencode', models: ['glm-5.1'] },
    ];
    const chain = buildSlotFallbackChain(slot, [slot], fallback);
    expect(chain).toEqual([
      { lineage: 'opencode', model: 'deepseek-v4-pro' },
      { lineage: 'opencode', model: 'kimi-k2.6' },
      { lineage: 'opencode', model: 'glm-5.1' },
    ]);
  });

  it('appends cross-lineage template fallbacks, diversity-first', () => {
    // Codex slot, fallback chain mixes same-lineage and cross-lineage.
    // Diversity-first ordering puts the lineage NOT in active slots
    // (anthropic, count=0) before the same-lineage continuation
    // (openai, count=1). User can override by manually re-ordering the
    // fallback list — within a single lineage, declared order wins.
    const slot = { lineage: 'openai', models: ['gpt-5.5'] };
    const fallback = [
      { lineage: 'openai', models: ['gpt-5.4'] },
      { lineage: 'anthropic', models: ['claude-opus-4-7'] },
    ];
    const chain = buildSlotFallbackChain(slot, [slot], fallback);
    expect(chain).toEqual([
      { lineage: 'openai', model: 'gpt-5.5' },
      // anthropic absent from active slots → tried first.
      { lineage: 'anthropic', model: 'claude-opus-4-7' },
      // openai already represented by the failing slot → tried last.
      { lineage: 'openai', model: 'gpt-5.4' },
    ]);
  });

  it('diversity-first: respects user order WITHIN the same lineage', () => {
    // Two anthropic fallbacks, one openai. anthropic count=1 (one
    // reviewer), openai count=1. Tie on counts → declared order
    // preserved: opus before sonnet, both before openai/gpt-5.4 only
    // when openai count is higher. Here openai matches anthropic so
    // declared order wins outright.
    const slot = { lineage: 'anthropic', models: ['claude-opus-4-7'] };
    const fallback = [
      { lineage: 'anthropic', models: ['claude-sonnet-4-6'] },
      { lineage: 'anthropic', models: ['claude-haiku-4-5'] },
    ];
    const chain = buildSlotFallbackChain(slot, [slot], fallback);
    expect(chain).toEqual([
      { lineage: 'anthropic', model: 'claude-opus-4-7' },
      { lineage: 'anthropic', model: 'claude-sonnet-4-6' },
      { lineage: 'anthropic', model: 'claude-haiku-4-5' },
    ]);
  });

  it('v0.8: cross-lineage fallback is the only entry when slot is exhausted', () => {
    // Slot with one model + one cross-lineage fallback — chain has two
    // entries, one per lineage.
    const slot = { lineage: 'openai', models: ['gpt-5.5'] };
    const fallback = [{ lineage: 'anthropic', models: ['claude-opus-4-7'] }];
    const chain = buildSlotFallbackChain(slot, [slot], fallback);
    expect(chain).toEqual([
      { lineage: 'openai', model: 'gpt-5.5' },
      { lineage: 'anthropic', model: 'claude-opus-4-7' },
    ]);
  });

  it('dedups against the slot itself — never appends the slot model again', () => {
    const slot = { lineage: 'opencode', models: ['kimi-k2.6'] };
    const fallback = [{ lineage: 'opencode', models: ['kimi-k2.6'] }];
    const chain = buildSlotFallbackChain(slot, [slot], fallback);
    expect(chain).toEqual([{ lineage: 'opencode', model: 'kimi-k2.6' }]);
  });

  it('user spec: reviewers=[kimi, deepseek] + fallback=[kimi] does not spawn duplicate kimi', () => {
    const kimiSlot = { lineage: 'opencode', models: ['kimi-k2.6'] };
    const deepseekSlot = { lineage: 'opencode', models: ['deepseek-v4-pro'] };
    const fallback = [{ lineage: 'opencode', models: ['kimi-k2.6'] }];

    const deepseekChain = buildSlotFallbackChain(
      deepseekSlot,
      [kimiSlot, deepseekSlot],
      fallback,
    );
    expect(deepseekChain).toEqual([{ lineage: 'opencode', model: 'deepseek-v4-pro' }]);

    const kimiChain = buildSlotFallbackChain(
      kimiSlot,
      [kimiSlot, deepseekSlot],
      fallback,
    );
    expect(kimiChain).toEqual([{ lineage: 'opencode', model: 'kimi-k2.6' }]);
  });

  it('extended user spec: fallback=[kimi, glm-5.1] with [kimi, deep] → both slots get glm-5.1', () => {
    const kimiSlot = { lineage: 'opencode', models: ['kimi-k2.6'] };
    const deepseekSlot = { lineage: 'opencode', models: ['deepseek-v4-pro'] };
    const fallback = [
      { lineage: 'opencode', models: ['kimi-k2.6'] },
      { lineage: 'opencode', models: ['glm-5.1'] },
    ];

    const deepChain = buildSlotFallbackChain(
      deepseekSlot,
      [kimiSlot, deepseekSlot],
      fallback,
    );
    expect(deepChain).toEqual([
      { lineage: 'opencode', model: 'deepseek-v4-pro' },
      { lineage: 'opencode', model: 'glm-5.1' },
    ]);

    const kimiChain = buildSlotFallbackChain(
      kimiSlot,
      [kimiSlot, deepseekSlot],
      fallback,
    );
    expect(kimiChain).toEqual([
      { lineage: 'opencode', model: 'kimi-k2.6' },
      { lineage: 'opencode', model: 'glm-5.1' },
    ]);
  });

  it('flattens multi-model fallback rows in priority order', () => {
    const slot = { lineage: 'opencode', models: ['deepseek-v4-pro'] };
    const fallback = [
      { lineage: 'opencode', models: ['kimi-k2.6', 'glm-5.1', 'qwen3.6-plus'] },
    ];
    const chain = buildSlotFallbackChain(slot, [slot], fallback);
    expect(chain).toEqual([
      { lineage: 'opencode', model: 'deepseek-v4-pro' },
      { lineage: 'opencode', model: 'kimi-k2.6' },
      { lineage: 'opencode', model: 'glm-5.1' },
      { lineage: 'opencode', model: 'qwen3.6-plus' },
    ]);
  });

  it('dedups within the template fallback list itself (no double-append)', () => {
    const slot = { lineage: 'opencode', models: ['deepseek-v4-pro'] };
    const fallback = [
      { lineage: 'opencode', models: ['kimi-k2.6'] },
      { lineage: 'opencode', models: ['kimi-k2.6'] }, // duplicate row
    ];
    const chain = buildSlotFallbackChain(slot, [slot], fallback);
    expect(chain).toEqual([
      { lineage: 'opencode', model: 'deepseek-v4-pro' },
      { lineage: 'opencode', model: 'kimi-k2.6' },
    ]);
  });

  it('handles a slot with multiple per-slot models (chains both before fallback)', () => {
    const slot = {
      lineage: 'anthropic',
      models: ['claude-opus-4-7', 'claude-sonnet-4-6'],
    };
    const fallback = [
      { lineage: 'anthropic', models: ['claude-haiku-4-5'] },
    ];
    const chain = buildSlotFallbackChain(slot, [slot], fallback);
    expect(chain).toEqual([
      { lineage: 'anthropic', model: 'claude-opus-4-7' },
      { lineage: 'anthropic', model: 'claude-sonnet-4-6' },
      { lineage: 'anthropic', model: 'claude-haiku-4-5' },
    ]);
  });

  it('treats per-slot fallbacks as already-active — template fallback skips them', () => {
    const slot = {
      lineage: 'anthropic',
      models: ['claude-opus-4-7', 'claude-sonnet-4-6'],
    };
    const fallback = [
      { lineage: 'anthropic', models: ['claude-sonnet-4-6'] },
    ];
    const chain = buildSlotFallbackChain(slot, [slot], fallback);
    expect(chain).toEqual([
      { lineage: 'anthropic', model: 'claude-opus-4-7' },
      { lineage: 'anthropic', model: 'claude-sonnet-4-6' },
    ]);
  });

  it('diversity-first: prefers absent lineages over already-represented ones', () => {
    // Reviewers: openai + google + anthropic (1 each).
    // Fallbacks: anthropic/haiku, moonshot/kimi.
    // Failing slot is openai. moonshot count=0, anthropic count=1 →
    // kimi runs FIRST as the more diverse choice.
    const failingSlot = { lineage: 'openai', models: ['gpt-5.5'] };
    const activeSlots = [
      failingSlot,
      { lineage: 'google', models: ['gemini-3.1-pro-preview'] },
      { lineage: 'anthropic', models: ['claude-opus-4-7'] },
    ];
    const fallback = [
      { lineage: 'anthropic', models: ['claude-haiku-4-5'] },
      { lineage: 'moonshot', models: ['kimi-k2.6'] },
    ];
    const chain = buildSlotFallbackChain(failingSlot, activeSlots, fallback);
    expect(chain).toEqual([
      { lineage: 'openai', model: 'gpt-5.5' },
      // moonshot absent from active slots → tried before haiku.
      { lineage: 'moonshot', model: 'kimi-k2.6' },
      { lineage: 'anthropic', model: 'claude-haiku-4-5' },
    ]);
  });

  it('cross-lineage dedup uses (lineage, model) tuple — same model name on different lineages is allowed', () => {
    // Highly unusual but valid: a model name shared across lineages stays
    // distinct because the dedup key is (lineage, model).
    const slot = { lineage: 'openai', models: ['shared-model'] };
    const fallback = [{ lineage: 'anthropic', models: ['shared-model'] }];
    const chain = buildSlotFallbackChain(slot, [slot], fallback);
    expect(chain).toEqual([
      { lineage: 'openai', model: 'shared-model' },
      { lineage: 'anthropic', model: 'shared-model' },
    ]);
  });
});
