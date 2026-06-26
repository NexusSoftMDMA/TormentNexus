/**
 * Tests for the shared template-yaml validator.
 *
 * Pins the contracts the cockpit and the daemon both rely on:
 *   - Empty input is invalid with a friendly message
 *   - YAML syntax errors come through with line info when available
 *   - Structural (zod) errors flatten to {path, message} pairs that name
 *     the offending field exactly
 *   - A real builtin template (review-only) round-trips through the
 *     validator cleanly
 */
import { describe, it, expect } from 'vitest';
import { validateTemplateYaml } from '../src/lib/template-validation';

describe('validateTemplateYaml', () => {
  it('rejects empty input with an actionable message', () => {
    const result = validateTemplateYaml('');
    expect(result.valid).toBe(false);
    expect(result.issues).toHaveLength(1);
    expect(result.issues[0].path).toBe('<yaml>');
    expect(result.issues[0].message).toMatch(/empty/i);
  });

  it('rejects whitespace-only input', () => {
    const result = validateTemplateYaml('   \n  \n');
    expect(result.valid).toBe(false);
    expect(result.issues[0].path).toBe('<yaml>');
  });

  it('surfaces YAML syntax errors at the parser stage', () => {
    // Block scalar with mismatched indent — yaml package rejects this.
    const bad = 'name: foo\n  description:\n    - bad';
    const result = validateTemplateYaml(bad);
    expect(result.valid).toBe(false);
    expect(result.issues[0].path).toBe('<yaml>');
  });

  it('surfaces zod structural errors with field paths', () => {
    // Parses as YAML but fails the schema (missing required fields).
    const yamlText = 'id: x\nname: x\n';
    const result = validateTemplateYaml(yamlText);
    expect(result.valid).toBe(false);
    // At least one issue should reference `description` or `phases` —
    // both are required at the top level.
    const paths = result.issues.map((i) => i.path);
    expect(paths.some((p) => p.includes('description') || p.includes('phases'))).toBe(
      true,
    );
  });

  it('surfaces nested phase errors with dotted paths', () => {
    const yamlText = `
id: bad-template
name: bad-template
description: bad
phases:
  - id: review
    kind: review
    title: Review
    doer:
      lineage: not-a-real-lineage
      models: []
    inputs:
      include: []
      exclude: []
    iterate:
      maxRounds: 3
      onDisagreement: continue
      shareSessionAcrossRounds: true
      shareSessionAcrossPhases: false
`;
    const result = validateTemplateYaml(yamlText);
    expect(result.valid).toBe(false);
    // The bad lineage should produce a path that includes 'phases.0.doer.lineage'.
    const paths = result.issues.map((i) => i.path);
    expect(paths.some((p) => p.startsWith('phases.0.doer'))).toBe(true);
  });

  it('accepts a minimal valid review-only template', () => {
    const yamlText = `
id: review-only
name: Review Only
description: Single-pass review of a runtime artifact.
agreementThreshold: 0.66
onThresholdMet: ask
maxRounds: 1
yoloDefault: false
phases:
  - id: review
    kind: review_only
    title: External Review
    description: Three lineages critique the supplied artifact.
    reviewer:
      require: 1
      crossLineage: true
      candidates:
        - lineage: anthropic
          models: [claude-opus-4-7]
        - lineage: openai
          models: [gpt-5.5]
        - lineage: google
          models: [gemini-3.1-pro-preview]
    inputs:
      include: []
      exclude: []
`;
    const result = validateTemplateYaml(yamlText);
    expect(result.valid).toBe(true);
    expect(result.issues).toHaveLength(0);
  });

  it('rejects review_only mixed with other phase kinds (hybrid templates)', () => {
    const yamlText = `
id: hybrid
name: hybrid
description: should not be allowed
phases:
  - id: a
    kind: review_only
    title: A
    reviewer:
      require: 1
      crossLineage: true
      candidates:
        - lineage: anthropic
          models: [claude-opus-4-7]
    inputs:
      include: []
      exclude: []
  - id: b
    kind: review
    title: B
    doer:
      lineage: anthropic
      models: [claude-opus-4-7]
    inputs:
      include: []
      exclude: []
    iterate:
      maxRounds: 1
      onDisagreement: continue
      shareSessionAcrossRounds: true
      shareSessionAcrossPhases: false
`;
    const result = validateTemplateYaml(yamlText);
    expect(result.valid).toBe(false);
    // The refine() guard fires at the phases array level.
    expect(
      result.issues.some((i) =>
        /review_only.*cannot be mixed/i.test(i.message),
      ),
    ).toBe(true);
  });
});
