# tormentnexus Frontier-Model Roundtable Debate Prompt

_Last updated: 2026-03-19_

Use this prompt with multiple frontier models **after** providing `docs/PROJECT_ROUNDTABLE_BRIEF.md`.

## Prompt

You are participating in a structured architecture and product roundtable about **tormentnexus**, a local AI operations control plane for builders.

You are not here to hype the repo, and you are not here to recommend “build everything.”

You are reviewing a real codebase with substantial implemented breadth, meaningful recent validation work, and equally meaningful risks around documentation drift, scope inflation, and uneven maturity.

Assume the following are true unless you can justify a correction:

- tormentnexus is fundamentally a **local AI control plane**, not a chatbot app.
- The most credible kernel today is:
  1. **MCP Router / Aggregator**
  2. **Provider Router / Fallback Engine**
  3. **Session Supervisor**
  4. **Operator Dashboard**
  5. **Memory / Context Layer**
- tormentnexus already has real implementations for startup/readiness, MCP inspection, session supervision, billing/provider surfaces, and memory search/timeline/pivot flows.
- tormentnexus’s largest near-term risks are:
  - documentation drift
  - scope inflation
  - product truth drift
  - parity theater
  - monorepo/reference noise
- Current active release-sensitive work is still centered on:
  - MCP dashboard runtime/import robustness
  - session supervisor worktree/attach reliability
  - startup truthfulness regression prevention

Also assume the following adjacent systems contain useful ideas, but tormentnexus does **not** need full parity with them before 1.0:

- **TormentNexus / MCP routers** — search, progressive disclosure, middleware, lifecycle ideas
- **tormentnexus / memory systems** — capture, compression, context harvesting
- **MCP-SuperAssistant / browser integrations** — browser automation and web-chat bridging
- **Jules / cloud-dev / autonomy systems** — session portability, replay, keeper loops, operator-safe autonomy

## Your objective

Produce the **best realistic roadmap from this point forward**.

Optimize for:

1. a believable tormentnexus 1.0
2. truthful operator UX over breadth theater
3. correct sequencing of hardening work vs expansion work
4. selective assimilation, not uncontrolled scope absorption
5. architectural coherence and maintainability

## Required questions

1. What is tormentnexus’s smallest compelling 1.0?
2. Which current subsystems are already kernel-grade and should be protected?
3. Which surfaces are visible but should be demoted, quarantined, or explicitly deferred?
4. Which current open blocker matters more right now: MCP runtime robustness or session attach/worktree reliability?
5. Which parts of the broad vision belong in 1.5 rather than 1.0?
6. Which 3 to 6 next slices should the maintainer prioritize next?
7. What are the top ways tormentnexus could still fail from here?

## Hard constraints

- Do **not** recommend “build everything.”
- Do **not** assume every existing route should remain a 1.0 surface.
- Do **not** optimize for impressive demo breadth over runtime truthfulness.
- Prefer boring, inspectable infrastructure over speculative autonomy.
- If you recommend a feature, explain why it strengthens tormentnexus as a control plane.
- If you reject or defer a feature, explain the distraction/risk it introduces.

## Required output format

Use exactly these sections:

### 1. Verdict
Give a blunt 1-paragraph judgment on tormentnexus’s current state.

### 2. Kernel
List the 3 to 6 core things tormentnexus fundamentally should be.

### 3. Stop doing
List what tormentnexus should demote, quarantine, or stop pretending is product-complete.

### 4. Kernel vs ornament
Provide a 2-column table:
- **Kernel: protect and deepen**
- **Ornament: defer, shrink, or quarantine**

### 5. Proposed roadmap
Describe tormentnexus **1.0**, **1.5**, and **2.0** from this point forward.

### 6. Next slices
List the next 3 to 6 implementation slices in order, with a short reason for each.

### 7. Risks
List the top 5 risks.

### 8. Final recommendation
End with the single most important next move.

## Preferred tone

- skeptical
- specific
- architecture-aware
- product-aware
- comfortable saying **not yet**
