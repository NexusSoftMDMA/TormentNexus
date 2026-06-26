import { describe, it, expect } from 'vitest';
import { getParsedTemplate } from '../src/daemon/index';

const SAMPLE_YAML = `
id: test
name: Test Template
description: smoke
phases:
  - id: phase-a
    kind: review
    title: Phase A
    doer:
      lineage: anthropic
      models: [claude-opus-4-7]
    reviewer:
      require: 1
      crossLineage: true
      candidates:
        - lineage: openai
          models: [gpt-5.5]
    inputs:
      include: []
      exclude: []
    iterate:
      maxRounds: 1
      onDisagreement: continue
`;

describe('getParsedTemplate', () => {
  it('returns same parsed object on cache hit (same stamp)', () => {
    const a = getParsedTemplate('cache-hit', SAMPLE_YAML, 1);
    const b = getParsedTemplate('cache-hit', SAMPLE_YAML, 1);
    expect(a).toBe(b);
  });

  it('reparses on cache miss (different stamp)', () => {
    const a = getParsedTemplate('cache-miss', SAMPLE_YAML, 1);
    const b = getParsedTemplate('cache-miss', SAMPLE_YAML, 2);
    expect(a).not.toBe(b);
    expect(a).toEqual(b);
  });

  it('throws on invalid yaml', () => {
    expect(() =>
      getParsedTemplate('bad', '::: not valid yaml :::', 1),
    ).toThrow();
  });
});
