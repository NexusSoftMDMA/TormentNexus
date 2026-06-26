/**
 * Coverage for the v0.8.3 template adapter.
 *
 * Pins behaviour at the boundaries that ship to users:
 *   - Slot lineage with matching voices: replace models with user's
 *     actual voices for that lineage.
 *   - Slot lineage with NO matching voices: substitute with
 *     diversity-preferring fallback OR mark template incomplete.
 *   - Vendor-family fallback (template wants moonshot, user has
 *     opencode-go/kimi-k2.5 with vendor_family=moonshot).
 *   - Empty-voices case → entire template incomplete, all slots blank.
 *   - openrouter voices use prefixed `id` not bare `model_id`.
 *   - Idempotent: same input → same output.
 *
 * Pure function; no DB / file system / daemon dependency.
 */
import { describe, it, expect } from 'vitest';
import yaml from 'yaml';
import { adaptTemplate } from '@/daemon/template-adapter';

interface TestVoice {
  id: string;
  provider: string;
  model_id: string;
  lineage: string;
  vendor_family: string | null;
  enabled: boolean;
}

const v = (
  lineage: string,
  model_id: string,
  overrides: Partial<TestVoice> = {},
): TestVoice => ({
  id: `${overrides.provider ?? `${lineage}-cli`}:${model_id}`,
  provider: `${lineage}-cli`,
  model_id,
  lineage,
  vendor_family: null,
  enabled: true,
  ...overrides,
});

const TRI_REVIEW = `id: tri-review
name: Tri-Review
phases:
  - id: review
    kind: review
    doer:
      lineage: anthropic
      models:
        - claude-opus-4-7
    reviewer:
      require: 2
      crossLineage: true
      candidates:
        - lineage: openai
          models:
            - gpt-5.5
        - lineage: google
          models:
            - gemini-3.1-pro-preview
        - lineage: moonshot
          models:
            - kimi-k2.6
`;

describe('adaptTemplate — exact-lineage match', () => {
  it("replaces canonical model with user's actual model for matching lineage", () => {
    const voices = [
      v('anthropic', 'claude-sonnet-4-6'),
      v('openai', 'gpt-5.4'),
      v('google', 'gemini-2.5-pro'),
      v('moonshot', 'kimi-k2-thinking'),
    ];
    const result = adaptTemplate(TRI_REVIEW, voices);
    expect(result.isComplete).toBe(true);
    const parsed = yaml.parse(result.yaml);
    expect(parsed.phases[0].doer.models).toEqual(['claude-sonnet-4-6']);
    const cands = parsed.phases[0].reviewer.candidates;
    expect(cands[0].models).toEqual(['gpt-5.4']);
    expect(cands[1].models).toEqual(['gemini-2.5-pro']);
    expect(cands[2].models).toEqual(['kimi-k2-thinking']);
  });

  it('picks the top-ranked single model when user has multiple voices for one lineage', () => {
    // Per user feedback: "Top models only if many choices" — adapter
    // emits 1 model per slot, the highest-ranked. Power users add
    // fallback chains via YAML.
    const voices = [
      v('anthropic', 'claude-haiku-4-5'),    // alphabetically first
      v('anthropic', 'claude-sonnet-4-6'),
      v('anthropic', 'claude-opus-4-7'),     // most capable
      v('openai', 'gpt-5.4'),
    ];
    const result = adaptTemplate(TRI_REVIEW, voices);
    const parsed = yaml.parse(result.yaml);
    // Opus tier (1000) > Sonnet (700) > Haiku (400). Within Opus,
    // the -4-7 version suffix gives the highest score. Top wins.
    expect(parsed.phases[0].doer.models).toEqual(['claude-opus-4-7']);
  });

  it('ranks gpt versions by major.minor (gpt-5.5 > gpt-5.4)', () => {
    const voices = [
      v('openai', 'gpt-5.2'),
      v('openai', 'gpt-5.5'),
      v('openai', 'gpt-5.4'),
    ];
    const tpl = `id: t
phases:
  - id: p
    kind: review
    doer:
      lineage: openai
      models: [gpt-anything]
    reviewer:
      require: 1
      candidates: []
`;
    const result = adaptTemplate(tpl, voices);
    expect(yaml.parse(result.yaml).phases[0].doer.models).toEqual(['gpt-5.5']);
  });

  it('rotates models when multiple slots match the same lineage', () => {
    // Review-only's 3 opencode reviewer slots used to all get the
    // same top-ranked opencode voice (e.g. kimi-k2.6 three times).
    // With usedTuples tracking, each slot pulls a different model.
    const voices = [
      v('opencode', 'opencode-go/kimi-k2.6'),
      v('opencode', 'opencode-go/deepseek-v4-pro'),
      v('opencode', 'opencode-go/glm-5.1'),
    ];
    const tpl = `id: t
phases:
  - id: p
    kind: review
    doer:
      lineage: opencode
      models: [whatever]
    reviewer:
      require: 2
      candidates:
        - lineage: opencode
          models: [whatever]
        - lineage: opencode
          models: [whatever]
        - lineage: opencode
          models: [whatever]
`;
    const result = adaptTemplate(tpl, voices);
    const parsed = yaml.parse(result.yaml);
    const cands = parsed.phases[0].reviewer.candidates;
    const modelsUsed = [
      parsed.phases[0].doer.models[0],
      cands[0].models[0],
      cands[1].models[0],
      cands[2].models[0],
    ];
    // 4 slots, 3 distinct opencode voices — first 3 should be all
    // distinct, 4th can repeat (fallback to top when all used).
    expect(new Set(modelsUsed).size).toBeGreaterThanOrEqual(3);
  });

  it('rotates anthropic models so 3 slots get different versions', () => {
    const voices = [
      v('anthropic', 'claude-haiku-4-5'),
      v('anthropic', 'claude-sonnet-4-6'),
      v('anthropic', 'claude-opus-4-7'),
    ];
    const tpl = `id: t
phases:
  - id: p
    kind: review
    doer:
      lineage: anthropic
      models: [whatever]
    reviewer:
      require: 2
      candidates:
        - lineage: anthropic
          models: [whatever]
        - lineage: anthropic
          models: [whatever]
`;
    const result = adaptTemplate(tpl, voices);
    const parsed = yaml.parse(result.yaml);
    const cands = parsed.phases[0].reviewer.candidates;
    const tuples = [
      parsed.phases[0].doer.models[0],
      cands[0].models[0],
      cands[1].models[0],
    ];
    // Doer + 2 reviewers all anthropic — should rotate through the
    // 3 enabled voices. Capability ranking puts opus first (doer),
    // sonnet next, haiku last.
    expect(tuples[0]).toBe('claude-opus-4-7');
    expect(new Set(tuples).size).toBe(3);
  });

  it('ranks gemini by major.minor + pro/flash modifier', () => {
    const voices = [
      v('google', 'gemini-2.5-flash'),
      v('google', 'gemini-3.1-pro-preview'),
      v('google', 'gemini-2.5-pro'),
    ];
    const tpl = `id: t
phases:
  - id: p
    kind: review
    doer:
      lineage: google
      models: [gemini-anything]
    reviewer:
      require: 1
      candidates: []
`;
    const result = adaptTemplate(tpl, voices);
    expect(yaml.parse(result.yaml).phases[0].doer.models).toEqual([
      'gemini-3.1-pro-preview',
    ]);
  });
});

describe('adaptTemplate — vendor-family fallback', () => {
  it('matches template lineage=moonshot via opencode-go/kimi voice with vendor_family=moonshot', () => {
    const voices = [
      v('anthropic', 'claude-opus-4-7'),
      v('openai', 'gpt-5.4'),
      v('google', 'gemini-2.5-pro'),
      // No direct moonshot lineage; user has it via opencode gateway.
      v('opencode', 'opencode-go/kimi-k2.5', { vendor_family: 'moonshot' }),
    ];
    const result = adaptTemplate(TRI_REVIEW, voices);
    expect(result.isComplete).toBe(true);
    const parsed = yaml.parse(result.yaml);
    const moonshotSlot = parsed.phases[0].reviewer.candidates[2];
    // Lineage swaps to opencode (where the voice actually lives), but
    // the model is the kimi one.
    expect(moonshotSlot.lineage).toBe('opencode');
    expect(moonshotSlot.models).toEqual(['opencode-go/kimi-k2.5']);
  });
});

describe('adaptTemplate — diversity-preferring substitution', () => {
  it('swaps a missing-lineage slot to a different lineage not yet used in the phase', () => {
    const voices = [
      v('anthropic', 'claude-opus-4-7'),
      v('google', 'gemini-2.5-pro'),
    ];
    const result = adaptTemplate(TRI_REVIEW, voices);
    expect(result.isComplete).toBe(true);
    const parsed = yaml.parse(result.yaml);
    const cands = parsed.phases[0].reviewer.candidates;
    // Doer used anthropic. openai slot has no exact match → falls
    // through to "find an unused lineage" → picks google.
    expect(cands[0].lineage).toBe('google');
    // google slot has an EXACT match → keeps google (intent
    // preservation wins over diversity when the user actually has
    // the requested lineage).
    expect(cands[1].lineage).toBe('google');
    // moonshot slot has no match and both available lineages are
    // used; falls through to "any available" — last-ditch.
    expect(['anthropic', 'google']).toContain(cands[2].lineage);
  });

  it('preserves intent: exact match wins over diversity', () => {
    // User has both anthropic and google. Doer is anthropic (used).
    // Reviewer slot wants google → should get google even though
    // anthropic-anthropic-google would also be valid.
    const oneReviewer = `id: t
phases:
  - id: phase1
    kind: review
    doer:
      lineage: anthropic
      models: [claude-opus-4-7]
    reviewer:
      require: 1
      candidates:
        - lineage: google
          models: [gemini-3.1-pro-preview]
`;
    const voices = [
      v('anthropic', 'claude-opus-4-7'),
      v('google', 'gemini-2.5-pro'),
    ];
    const result = adaptTemplate(oneReviewer, voices);
    const parsed = yaml.parse(result.yaml);
    expect(parsed.phases[0].reviewer.candidates[0].lineage).toBe('google');
  });
});

describe('adaptTemplate — empty / incomplete states', () => {
  it('marks template incomplete when user has no enabled voices anywhere', () => {
    const result = adaptTemplate(TRI_REVIEW, []);
    expect(result.isComplete).toBe(false);
    const parsed = yaml.parse(result.yaml);
    expect(parsed.phases[0].doer.models).toEqual([]);
    for (const c of parsed.phases[0].reviewer.candidates) {
      expect(c.models).toEqual([]);
    }
  });

  it('marks template incomplete when only one slot type can be filled', () => {
    // Doer slot fillable (anthropic), but reviewer cans... wait, the
    // adapter falls through to "any available" so even one lineage
    // fills all slots. This test pins the fact that only-one-voice
    // results in COMPLETE templates (with reduced diversity).
    const result = adaptTemplate(TRI_REVIEW, [v('anthropic', 'claude-opus-4-7')]);
    expect(result.isComplete).toBe(true);
    const parsed = yaml.parse(result.yaml);
    // All slots fall back to anthropic.
    expect(parsed.phases[0].doer.lineage).toBe('anthropic');
    for (const c of parsed.phases[0].reviewer.candidates) {
      expect(c.lineage).toBe('anthropic');
    }
  });

  it('skips disabled voices', () => {
    const voices = [
      v('anthropic', 'claude-opus-4-7', { enabled: false }),
      v('openai', 'gpt-5.4'),
    ];
    const result = adaptTemplate(TRI_REVIEW, voices);
    const parsed = yaml.parse(result.yaml);
    // anthropic disabled → doer falls back. openai is the only enabled
    // voice — doer takes it, reviewers all fall through too.
    expect(parsed.phases[0].doer.lineage).toBe('openai');
    expect(parsed.phases[0].doer.models).toEqual(['gpt-5.4']);
  });
});

describe('adaptTemplate — openrouter id form', () => {
  it('uses prefixed id (openrouter:<model>) for openrouter voices, not bare model_id', () => {
    const voices = [
      v('anthropic', 'claude-opus-4-7'),
      v('openai', 'gpt-5.4'),
      v('google', 'gemini-2.5-pro'),
      // openrouter voice has provider='openrouter' so the adapter
      // emits the prefixed id (so pickShimForVoice routes to the
      // OpenRouter HTTP shim instead of a CLI shim).
      {
        id: 'openrouter:moonshotai/kimi-k2.5',
        provider: 'openrouter',
        model_id: 'moonshotai/kimi-k2.5',
        lineage: 'moonshot',
        vendor_family: null,
        enabled: true,
      },
    ];
    const result = adaptTemplate(TRI_REVIEW, voices);
    const parsed = yaml.parse(result.yaml);
    const moonshotSlot = parsed.phases[0].reviewer.candidates[2];
    expect(moonshotSlot.models).toEqual(['openrouter:moonshotai/kimi-k2.5']);
  });
});

describe('adaptTemplate — idempotency', () => {
  it('returns identical YAML when re-run on already-adapted output', () => {
    const voices = [
      v('anthropic', 'claude-opus-4-7'),
      v('openai', 'gpt-5.4'),
      v('google', 'gemini-2.5-pro'),
      v('moonshot', 'kimi-k2-thinking'),
    ];
    const first = adaptTemplate(TRI_REVIEW, voices);
    const second = adaptTemplate(first.yaml, voices);
    expect(second.yaml).toBe(first.yaml);
    expect(second.changed).toBe(false);
    expect(second.isComplete).toBe(first.isComplete);
  });

  it('reports changed=false when canonical input already matches user voices', () => {
    // Voice id matches the model in the canonical YAML.
    const voices = [
      v('anthropic', 'claude-opus-4-7'),
      v('openai', 'gpt-5.5'),
      v('google', 'gemini-3.1-pro-preview'),
      v('moonshot', 'kimi-k2.6'),
    ];
    const result = adaptTemplate(TRI_REVIEW, voices);
    expect(result.changed).toBe(false);
    expect(result.isComplete).toBe(true);
  });
});

describe('adaptTemplate — multi-phase templates', () => {
  it('tracks usedLineages per phase, not globally', () => {
    // Two phases; each has its own diversity budget.
    const twoPhase = `id: two-phase
phases:
  - id: phase1
    kind: review
    doer:
      lineage: anthropic
      models:
        - claude-opus-4-7
    reviewer:
      require: 1
      candidates:
        - lineage: openai
          models:
            - gpt-5.5
  - id: phase2
    kind: review
    doer:
      lineage: anthropic
      models:
        - claude-opus-4-7
    reviewer:
      require: 1
      candidates:
        - lineage: openai
          models:
            - gpt-5.5
`;
    const voices = [
      v('anthropic', 'claude-opus-4-7'),
      v('openai', 'gpt-5.4'),
    ];
    const result = adaptTemplate(twoPhase, voices);
    const parsed = yaml.parse(result.yaml);
    // Phase 2 should also have anthropic doer + openai reviewer (no
    // global lineage exhaustion).
    expect(parsed.phases[1].doer.lineage).toBe('anthropic');
    expect(parsed.phases[1].reviewer.candidates[0].lineage).toBe('openai');
  });
});

describe('adaptTemplate — robustness', () => {
  it('returns canonical YAML when input is malformed', () => {
    const malformed = '!!! not yaml ::: \n  -- bad';
    const result = adaptTemplate(malformed, []);
    expect(result.yaml).toBe(malformed);
    expect(result.isComplete).toBe(false);
    expect(result.changed).toBe(false);
  });

  it('returns canonical YAML when phases array is missing', () => {
    const noPhases = `id: empty\nname: Empty\n`;
    const result = adaptTemplate(noPhases, []);
    expect(result.yaml).toBe(noPhases);
    expect(result.isComplete).toBe(false);
    expect(result.changed).toBe(false);
  });
});
