# Chorus Built-in Templates

4 production templates bundled with Chorus v0.5. Each YAML file defines a workflow phase sequence, reviewer strategy, and agreement policy.

## Quick Start

All templates use the same YAML schema (see `src/lib/template-schema.ts`). Load with:

```typescript
import yaml from 'yaml'
import { TemplateSchema } from '../src/lib/template-schema.ts'

const template = TemplateSchema.parse(yaml.parse(yamlContent))
```

## The 4 Templates

### 1. Code Review (`code-review.yaml`)

**Use for:** Quick peer review of an existing implementation.

- Single phase: you submit code; 3 reviewers from different LLM lineages critique it.
- Threshold: 2 of 3 must agree (66%) to pass.
- Best for: PRs, quick feedback loops, when you want broad coverage.

### 2. Bug Diagnosis (`bug-diagnose.yaml`)

**Use for:** Debugging complex issues where multiple perspectives help.

- Single phase: Anthropic hypothesizes what's broken; OpenAI challenges it.
- Adversarial: lower 50% threshold; disagreement surfaces root causes.
- Best for: Mysterious bugs, post-mortems, when divergence is useful.

### 3. Architecture Review (`architect-review.yaml`)

**Use for:** Decision-before-coding. Validate a design before implementation starts.

- Two phases: Anthropic drafts a proposal; 3 reviewers critique it.
- Threshold: 50% agreement needed (surfaces disagreement, doesn't rubber-stamp).
- Best for: Major refactors, API design, infrastructure decisions, risk mitigation upfront.

### 4. Red / Green Adversarial (`red-green.yaml`)

**Use for:** End-to-end code production with adversarial review at every phase.

- 7 phases: plan → spec → tests → implement → verify → final-review → divergence.
- **Critical:** implementer is blind to tests (no `inputs.exclude: [tests]`). Prevents overfitting to test literals; forces spec-driven discipline.
- Cross-lineage pairs at every phase (Anthropic ↔ OpenAI ↔ Xai rotation).
- Deterministic verify loop: test failures feed back as name-only (no bodies), driving iteration without information leakage.
- Best for: High-stakes code, safety-critical features, when you want multiple LLMs to catch what one misses. TheDailyClaude (r/ClaudeAI) designed this flow.

## Schema Highlights

Each phase has:
- **doer**: the LLM producing an artifact
- **reviewer**: optional cross-lineage peer(s) that gate the phase
- **inputs**: control what the doer sees (`include`/`exclude` for info asymmetry)
- **iterate**: retry policy on disagreement

All templates support:
- `agreementThreshold`: 0.0–1.0 (or named: `unanimous` / `majority` / `any`)
- `onThresholdMet`: what to do when reviewers agree (`merge` / `ask` / `review`)
- `maxRounds`: how many revision loops before escalating

## Customization

User-authored templates arrive in v0.6. For now, these 4 cover 80% of use cases.

---

**See also:** `src/lib/template-schema.ts` (Zod schema), `src/lib/mock-data.ts` (runtime references).
