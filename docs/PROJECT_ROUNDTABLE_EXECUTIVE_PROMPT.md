# tormentnexus Roundtable Executive Prompt

_Last updated: 2026-03-19_

Use this shorter prompt when you want fast, high-signal feedback from multiple models without sending the full long-form debate prompt.

## Prompt

You are reviewing **tormentnexus**, a local AI operations control plane for builders.

tormentnexus is **not** supposed to be a general chatbot product. Its credible kernel is:

- MCP router / aggregation / tool inspection
- provider fallback and billing/operator truthfulness
- session supervision for external CLI runtimes
- operator dashboard for truthful state and control
- memory/context storage and retrieval

Current reality:

- tormentnexus already has a broad real dashboard and a substantial TypeScript monorepo runtime.
- Startup/readiness, MCP search/inspection, session supervision, billing/provider, and memory surfaces are all materially implemented.
- The biggest risks are documentation drift, scope inflation, uneven maturity, and too many visible surfaces relative to the 1.0 trust story.
- The two most important still-open release-sensitive areas are:
  1. MCP dashboard runtime/import robustness
  2. Session supervisor worktree/attach reliability

Important: do **not** recommend full parity with every adjacent tool before 1.0.

## Your task

Give the best realistic recommendation for how tormentnexus should evolve from here into a dependable, shippable product.

Optimize for:

- believable 1.0 scope
- truthful operator UX
- strong sequencing
- architectural coherence
- boring reliability over breadth theater

## Answer these questions

1. What is the smallest compelling 1.0 for tormentnexus?
2. What should tormentnexus protect as its kernel?
3. What should tormentnexus stop pretending to be right now?
4. What should move to 1.5 or 2.0?
5. What are the next 3 to 6 slices?
6. What are the top risks from here?

## Required output

Use exactly these sections:

### Verdict

### Kernel

### Stop doing

### Roadmap

### Next slices

### Risks

### Final recommendation
