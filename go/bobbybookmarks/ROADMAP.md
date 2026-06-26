# Borg Intelligence — Definitive Feature Roadmap

## Phase 1: MEMORY (Foundation)
- Knowledge Graph mapping between skills, tools, and past session history.
- Tiered Memory implementation (L1 Hot / L2 Warm / L3 Cold with heat-based promotion).
- Cross-session persistence for durable state management.

## Phase 2: SKILL ENGINE (Highest Value)
- `SKILL.md` template systems for extraction and task resolution strategies.
- Tracking of execution win-rates with auto-retirement for bad skills.
- DSPy-style prompt optimization built into the workflow.

## Phase 3: CONTEXT MANAGER (Budget Control)
- Progressive schema disclosure (lazy-loading tools on demand).
- Context re-injection hooks after compaction events to save tokens.
- Dynamic token budget allocation assigned per active tool.

## Phase 4: VERIFICATION (Safety & Autonomy)
- Closed-loop self-healing (Build -> Fail -> LLM Error Resolution -> Retry).
- Multi-model cross-validation pipelines (e.g. 3-model verification for critical tasks).
- Human-in-the-loop (HITL) gates backed by graph blast-radius analysis.

## Phase 5: THE RECURSIVE LOOP (Continuous Improvement)
- Self-modification informed by historical graph memory.
- Triggers linking verification failure/success directly to skill evolution.
- System that dynamically updates its own models based on verification outcomes.
