"""
BORG FEATURE SPEC — Derived from 13,503 bookmark intelligence reports
The bookmark database is RESEARCH INPUT about the AI engineering ecosystem.
Borg is a SEPARATE system that should be INFORMED BY those trends.

Method: Ecosystem saturation analysis + enrichment ratios + gap detection
"""

BORG_SPEC = """
========================================================================
  BORG: THE SYSTEM THE DATA SAYS SHOULD EXIST
========================================================================

  ECOSYSTEM SATURATION (what's already built):
    Protocol:      6,406 systems (66%)  — MCP servers, bridges, gateways
    Agent Runtime: 5,059 systems (52%)  — orchestrators, frameworks, workflows  
    UX:            5,009 systems (52%)  — CLIs, IDEs, dashboards
    Intelligence:  3,516 systems (36%)  — RAG, search, extraction
    Context:       3,328 systems (34%)  — chunking, indexing, token mgmt
    Tools:         2,424 systems (25%)  — MCP tools, skill registries
    Infra:         2,067 systems (21%)  — sandboxes, CI/CD, hosting
    Verification:  1,729 systems (18%)  — review, testing, quality gates
    Self-Mod:      1,506 systems (15%)  — evolution, learning, recursion
    Memory:        1,445 systems (15%)  — persistence, recall, knowledge

  SATURATION RATIOS (relative to Agent Runtime = 1.0):
    Protocol:     1.27x  OVERBUILT
    Agent Runtime:1.00x  (baseline)
    UX:           0.99x  OVERBUILT
    Intelligence: 0.69x  adequate
    Context:      0.66x  adequate
    Tools:        0.48x  UNDERBUILT  <-- BORG OPPORTUNITY
    Infra:        0.41x  UNDERBUILT  <-- BORG OPPORTUNITY
    Verification: 0.34x  UNDERBUILT  <-- BORG OPPORTUNITY
    Self-Mod:     0.30x  UNDERBUILT  <-- BORG OPPORTUNITY
    Memory:       0.29x  UNDERBUILT  <-- BORG OPPORTUNITY

  MISSING COMBINATIONS (deficit vs expected co-occurrence):
    Self-Mod + Tools:   2.7% deficit  (should be 3.9%, only 1.2%)
    Memory + Tools:     1.7% deficit  (should be 3.7%, only 2.1%)
    Infra + Memory:     1.2% deficit  (should be 3.2%, only 2.0%)

  BORG IS NOT:
    Another agent framework (52% saturated)
    Another MCP server (66% saturated)
    Another CLI/dashboard (52% saturated)

  BORG IS:
    A Memory + Self-Modification + Verification system
    that CONNECTS to the saturated layers via Protocol.
    The operating system for agent intelligence.

========================================================================
  BORG'S 7 CORE CAPABILITIES (ranked by data evidence)
========================================================================

  1. KNOWLEDGE GRAPH BACKBONE
     Evidence: Most cross-layer references, highest zero-co-occurrence
     The connective tissue no other system has. Maps relationships
     between tools, skills, memory, and decisions.
     Enables: blast-radius analysis, semantic skill search, failure tracing
     Unique: Graph Memory + HITL Gates (signal: 1,984) has ZERO implementations

  2. TIERED MEMORY CORE
     Evidence: 14x enrichment, 29% saturation = massive gap
     L1 Hot (session) / L2 Warm (recent, heat-ranked) / L3 Cold (archive).
     With adaptive forgetting and heat-based promotion.
     Enables: every other feature (skills need memory, healing needs history)
     REFS: LangMem, MemoryOS, Meta/Harvard hierarchical, mcp-titan

  3. SKILL EVOLUTION ENGINE
     Evidence: 37x enrichment — STRONGEST SIGNAL IN ENTIRE DATASET
     SKILL.md templates, /evolve command, DSPy optimization.
     Learns which strategies work for which contexts.
     Enables: automatic prompt optimization, cross-agent skill transfer
     REFS: Anthropic SKILL.md, everything-claude-code, VoltAgent 5k skills

  4. CONTEXT BUDGET MANAGER
     Evidence: 20x + 10x enrichment combined (re-injection + token reduction)
     Token reduction + context re-injection + progressive schemas.
     Manages the LLM's context window as a scarce resource.
     Enables: running on smaller models, longer sessions, more tools
     REFS: Claude Code Hooks, MCP-Zero, Cloudflare Code Mode, ToolRAG

  5. SELF-HEALING VERIFIER
     Evidence: 15x enrichment
     Build -> fail -> error-to-LLM -> fix -> retry.
     Stop hooks, promise checking, idle-state self-healing.
     Enables: autonomous operation without human babysitting
     REFS: Ralph Wiggum, Post-Agentic Forges, agtx, Auggie Wiggum

  6. MULTI-MODEL VERIFICATION
     Evidence: 5-10x enrichment across sub-mechanisms
     Cross-validation, challenge logic, HITL gates.
     Uses 2-3 models to verify high-stakes decisions.
     Enables: confidence scoring, risk-based escalation
     REFS: Verdent AI 3-model, GSD planner-checker, CodeRabbit

  7. CROSS-SESSION PERSISTENCE
     Evidence: 4.5x enrichment, foundational for Memory layer
     Durable state that survives restarts.
     Execution checkpoints, architectural decisions, tool preferences.
     Enables: everything above persists between sessions
     REFS: Letta stateful agents, Penfield MCP, Lily Memory

========================================================================
  THE DATA-PROVEN UNIQUE VALUE
========================================================================

  No system in the 13,503-bookmark corpus combines:
    Graph Memory + Verification + Self-Modification

  That's what Borg is. The graph sees relationships, verification
  checks decisions, and self-modification improves from outcomes.

  The closest existing systems span 3-4 layers but miss the triangle:
    - Cloudflare Agents: 8 layers but no self-modification
    - Mimir: 7 layers but no verification
    - Hindsight: 7 layers but no tools layer
  
  Borg is the triangle: Memory + Self-Mod + Verify, connected via Protocol.

========================================================================
  BUILD ORDER
========================================================================

  Phase 1: MEMORY (foundation everything else needs)
    - Knowledge Graph (Neo4j or SQLite + adjacency lists)
    - Tiered Memory (L1/L2/L3 with heat promotion)
    - Cross-session persistence

  Phase 2: SKILL ENGINE (37x signal — build early)
    - SKILL.md template system
    - /evolve command with win-rate tracking
    - DSPy-style prompt optimization

  Phase 3: CONTEXT MANAGER (solves the budget problem)
    - Progressive schema disclosure
    - Context re-injection hooks
    - Token budget allocation per tool

  Phase 4: VERIFICATION (makes it safe to run autonomously)
    - Self-healing build-fix loop
    - Multi-model cross-validation
    - HITL gates with graph-based blast radius analysis

  Phase 5: THE RECURSIVE LOOP (connects everything)
    - Self-modification informed by graph memory
    - Verification triggers skill evolution
    - Memory improves from verification outcomes
    - The system that learns how to learn
========================================================================
"""
