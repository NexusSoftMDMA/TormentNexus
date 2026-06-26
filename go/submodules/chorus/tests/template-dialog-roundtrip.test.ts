/**
 * Round-trip regression: a template with OpenRouter reviewers must
 * survive emit → parse without losing the candidate. The bug was a
 * missing `openrouter` entry in DAEMON_TO_COCKPIT — parse silently
 * dropped every openrouter candidate, the form rendered 0 reviewers,
 * and save failed validation. Locked in here so it can't regress.
 */
import { describe, expect, it } from 'vitest';
import { parseYamlToForm } from '@/components/template-dialog/parse';

describe('template-dialog parse — openrouter lineage round-trip', () => {
  it('preserves an openrouter reviewer candidate on parse', () => {
    const yaml = `
id: review-only
name: Review Only
description: External review of an artifact.
author: chorus
agreementThreshold: 0.66
onThresholdMet: ask
maxRounds: 3
yoloDefault: false
phases:
  - id: review
    kind: review_only
    title: External Review
    description: One round of reviewers.
    reviewer:
      require: 1
      crossLineage: true
      candidates:
        - lineage: anthropic
          models:
            - claude-opus-4-7
        - lineage: openrouter
          models:
            - "openrouter:x-ai/grok-3"
    inputs:
      include: []
      exclude: []
    artifact:
      label: Artifact to review
      hint: Paste an artifact.
      maxBytes: 1048576
`;
    const result = parseYamlToForm(yaml, 'review-only');
    const phase = result.form.phases[0];
    expect(phase.reviewer.candidates).toContain('claude');
    expect(phase.reviewer.candidates).toContain('openrouter');
    expect(phase.reviewer.candidateModels?.openrouter).toEqual([
      'openrouter:x-ai/grok-3',
    ]);
  });

  it('preserves an openrouter fallback reviewer on parse', () => {
    const yaml = `
id: review-only
name: Review Only
description: x
author: chorus
phases:
  - id: review
    kind: review_only
    title: Review
    description: x
    reviewer:
      require: 1
      candidates:
        - lineage: anthropic
          models: [claude-opus-4-7]
    inputs: { include: [], exclude: [] }
    artifact: { label: a, hint: b, maxBytes: 1024 }
fallback:
  reviewer:
    - lineage: openrouter
      models:
        - "openrouter:x-ai/grok-3"
`;
    const result = parseYamlToForm(yaml, 'review-only');
    expect(result.form.fallbackReviewer).toEqual([
      { lineage: 'openrouter', model: 'openrouter:x-ai/grok-3' },
    ]);
  });
});
