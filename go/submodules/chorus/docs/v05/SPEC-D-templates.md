# Agent D — Built-in Templates

You own: `templates/*.yaml`, `templates/README.md`, `src/lib/template-schema.ts` (new — Zod schema).

## Build

4 built-in YAML templates that ship with v0.5. Each is a `Template` per the locked phase schema in `chorus_phase_schema.md` (memory file).

### Template schema (Zod)

```ts
const PhaseSchema = z.object({
  id: z.string(),                              // 'plan', 'implement', 'review', etc
  kind: z.enum(['plan','spec','tests','implement','review','verify','divergence']),
  title: z.string(),
  doer: z.object({
    lineage: z.enum(['anthropic','openai','google','xai','any']),
    models: z.array(z.string()).optional(),    // narrow to specific models
  }),
  reviewer: z.object({
    require: z.number().int().min(0).default(1),
    crossLineage: z.boolean().default(true),
    candidates: z.array(z.object({
      lineage: z.enum(['anthropic','openai','google','xai']),
      models: z.array(z.string()).optional(),
    })),
  }).optional(),                                // some phases (final-review) may have no reviewer
  inputs: z.object({
    include: z.array(z.string()).default([]),  // file globs or named refs
    exclude: z.array(z.string()).default([]),  // info asymmetry — e.g. exclude tests on impl phase
  }).default({ include: [], exclude: [] }),
  iterate: z.object({
    maxRounds: z.number().int().min(1).default(2),
    onDisagreement: z.enum(['continue','escalate','accept-doer']).default('continue'),
  }).default({ maxRounds: 2, onDisagreement: 'continue' }),
});

const TemplateSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string(),
  author: z.string().default('chorus'),
  agreementThreshold: z.number().min(0).max(1).default(0.66),
  onThresholdMet: z.enum(['merge','ask','review']).default('ask'),
  maxRounds: z.number().int().min(1).default(3),
  yoloDefault: z.boolean().default(false),
  phases: z.array(PhaseSchema).min(1),
});
```

Export types via `z.infer`. Place this in `src/lib/template-schema.ts` so daemon and UI both consume it.

### The 4 templates

#### 1. `templates/code-review.yaml`

The bread-and-butter template — 80% of usage. Single review phase.

- **Phase 1 (review)**: doer = the user's existing implementation (no doer, just submission). Reviewer = 3 cross-lineage (Anthropic, OpenAI, Google). `agreementThreshold: 0.66`, `onThresholdMet: ask`.

#### 2. `templates/bug-diagnose.yaml`

Single doer + reviewer, adversarial. For "what's broken and why?"

- **Phase 1 (plan)**: doer claude-opus-4-7 produces a hypothesis. Reviewer codex (gpt-5) challenges it. `iterate.maxRounds: 2`, `onDisagreement: continue`.

#### 3. `templates/architect-review.yaml`

Decision-before-coding template. 2 phases: plan + review.

- **Phase 1 (plan)**: doer claude-opus-4-7 drafts an architecture proposal.
- **Phase 2 (review)**: 3 reviewers cross-lineage critique it. `agreementThreshold: 0.5` (lower bar — we want disagreement surfaced, not consensus rubber-stamped).

#### 4. `templates/red-green.yaml`

TheDailyClaude's 7-phase adversarial flow. **Critical:** the implement phase has `inputs.exclude: ['tests']` so the implementer can't see the tests it must pass.

Phases (id, kind, brief):
1. `plan` (plan) — claude-opus drafts approach
2. `spec` (spec) — codex writes API contract from plan
3. `tests` (tests) — gemini writes failing tests against spec
4. `implement` (implement) — claude-opus implements; `inputs.exclude: ['tests']` (key info asymmetry — must reason from spec alone)
5. `verify` (verify) — deterministic test runner; pass/fail names only, no test bodies
6. `final-review` (review) — 3 reviewers cross-lineage final pass
7. `divergence` (divergence) — if final-review disagrees, loops back to `spec` with the disagreement summary

Set `iterate.maxRounds` carefully: implement should retry up to 3 times reading test failure NAMES; if still failing, `onDisagreement: escalate`.

Mark `author: '@TheDailyClaude (reddit)'`.

### `templates/README.md`

User-facing doc explaining what each template does and when to use it. Keep it under 100 lines.

## Don't touch

- `src/daemon/`, `src/mcp/`, `src/app/` — other agents
- `package.json`, `bin/` — foundation

## Acceptance

```bash
cd /home/ubuntu/dev/chorus
# Each YAML file parses against the Zod schema:
node -e '
  const { TemplateSchema } = await import("./src/lib/template-schema.ts");
  const yaml = require("yaml");
  const fs = require("fs");
  for (const f of fs.readdirSync("templates").filter(x=>x.endsWith(".yaml"))) {
    const parsed = TemplateSchema.parse(yaml.parse(fs.readFileSync(`templates/${f}`,"utf8")));
    console.log(f, "✓", parsed.id);
  }
'
pnpm typecheck                # passes
```

## Reference

Memory: `chorus_phase_schema.md`, `feedback_redgreen_adversarial_pattern.md`, `feedback_consensus_threshold_design.md`.
The existing prototype's `t-red-green` mock template (in `src/lib/mock-data.ts`) is a starting reference — but write the YAML fresh from the spec, don't just stringify the JSON.
