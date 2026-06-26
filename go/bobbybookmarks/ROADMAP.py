"""
Borg Intelligence — Definitive Feature Roadmap
Data source: 13,503 clean bookmarks | 1,967 elite (score 9-10)
Method: Mechanism enrichment analysis + co-occurrence + gap detection

ENRICHMENT SCORE = how much more likely a mechanism appears in
score-9-10 systems vs score-7-8. This measures what SEPARATES
the best from the rest, not just what's common.
"""

ROADMAP = """
========================================================================
  BORG INTELLIGENCE — DEFINITIVE FEATURE ROADMAP
  13,503 clean bookmarks | 1,967 elite (score 9-10) | 20 mechanisms
========================================================================

TIER 1: PROVEN DIFFERENTIATORS (Enrichment > 10x)
These are what elite systems do that average systems DON'T.

 1. SKILL EVOLUTION ENGINE                   37.3x enrichment
    - SKILL.md template system for extraction strategies
    - /evolve command: promote instinct -> learned skill
    - DSPy-style programmatic prompt optimization
    - Cross-agent skill synchronization
    - Win-rate tracking with auto-retirement of bad skills
    REFS: Anthropic SKILL.md, everything-claude-code /evolve,
          VoltAgent 5k+ SKILL.md runbooks, InstaVM DSPy conversion

 2. REPO-SCALE CONTEXT INDEXING              23.5x enrichment
    - Tree-sitter AST parsing for code repos
    - 20M+ token repository indexing with .cgc bundles
    - Pre-indexed repository context cards
    - Symbol-level graph querying (callers/callees)
    REFS: Plandex 2M token + tree-sitter, Chunkhound cAST

 3. CONTEXT RE-INJECTION AFTER COMPACTION    20.3x enrichment
    - PreToolUse/PostToolUse lifecycle hooks
    - Automatic context re-injection after compaction
    - Context bloat prevention via token budgets
    - Progressive schema disclosure
    REFS: Claude Code Hooks, ToolRAG bloat prevention,
          Meridian pre-compaction injection, MCP-Zero

 4. LIVE VISUALIZATION / SPATIAL NAV         19.6x enrichment
    - Spatial mindmap for pipeline navigation
    - 5-second heartbeat refresh
    - Strategic zoom: macro fleet -> micro extraction
    REFS: Claude Flow v3, Mission Control 32 panels

 5. SELF-HEALING BUILD-FIX                   14.9x enrichment
    - Execute -> exit code -> error-to-LLM -> fix -> retry
    - Stop Hook: intercept exit signals, check promises
    - Idle-state self-healing when agent is idle
    - Closed-loop: build -> test -> auto-fix (max 3)
    REFS: Ralph Wiggum, Post-Agentic Forges, agtx, Auggie Wiggum

 6. THREE-TIER MEMORY                        13.7x enrichment
    - L1 Hot: session working (in-process)
    - L2 Warm: recent facts (SQLite, heat-ranked)
    - L3 Cold: compressed archive (zlib, access promotion)
    - Heat-based promotion + adaptive forgetting
    REFS: LangMem, Meta/Harvard hierarchical, MemoryOS, mcp-titan

 7. PLANNER-CHECKER-REVISE LOOP              12.7x enrichment
    - Plan Mode: premium model creates PLAN.md
    - Checker: second model validates plan
    - Revise: loop back with error context if rejected
    - Architect/Implementer bifurcation
    REFS: Verdent AI, GSD Framework, Gemini Conductor, Jules

 8. TOKEN REDUCTION ENGINE                   10.5x enrichment
    - Progressive schema disclosure (load on demand)
    - Fact distillation: store facts, not raw chunks
    - Just-in-time schema injection
    - Constant context footprint via batch execution
    REFS: Anthropic MCP 98% reduction, Cloudflare 99.9%,
          Claude Tool Search 90%, MCP-Zero 98%

TIER 2: STRONG SIGNALS (5-10x)

 9.  CHALLENGE/ADVERSARIAL VERIFICATION      10.2x
10.  MECE CODEBASE PARTITIONING              10.2x
11.  MAP-ELITES EVOLUTION SEARCH             10.2x
12.  GIT WORKTREE ISOLATION                  9.2x
13.  DIFF LENS / CUMULATIVE DIFF REVIEW      6.5x
14.  HITL GATES                              6.0x
15.  CROSS-VALIDATION (3-MODEL VOTING)        5.1x
16.  REACTION ENGINE (CI/PR LOG INJECTION)   5.1x
17.  TMUX SESSION MANAGEMENT                  5.0x

TIER 3: EMERGING (2-5x)

18.  GRAPH MEMORY / KNOWLEDGE GRAPH           3.0x
19.  E2B/DOCKER SANDBOXING                    2.9x
20.  A2A PROTOCOL BRIDGE                      2.3x
21.  KANBAN BOARD ORCHESTRATION               2.4x
22.  SELF-IMPROVEMENT LOOPS                   3.9x
23.  AGENT TEAMS (PEER MESSAGING)             3.2x
24.  TIERED MODEL ROUTING                     3.5x
25.  MEMORY RECALL (CROSS-SESSION)            4.5x
26.  MCP TOOL DISCOVERY                       1.9x
27.  TELEMETRY/OBSERVABILITY                  1.9x

========================================================================
  THE 5 UNDISCOVERED COMBINATIONS
  High individual frequency + ZERO co-occurrence = predicted next-gen
========================================================================

  A.  GRAPH MEMORY + HITL GATES              Signal: 1,984
      No system combines knowledge graph with approval gates.
      BUILD: Graph-based impact analysis before approval decisions.
      Auto-approve low-impact, escalate high-impact.

  B.  GRAPH MEMORY + MCP TOOL DISCOVERY      Signal: 1,472
      No system uses a knowledge graph to inform tool discovery.
      BUILD: Semantic graph of past tool usage suggests which tools
      to discover for new domains.

  C.  GIT WORKTREE + GRAPH MEMORY            Signal: 1,344
      No system tracks worktree state in a knowledge graph.
      BUILD: Graph-visualized worktree dependency map. Detect
      conflicts BEFORE they happen via dependency tracing.

  D.  GRAPH MEMORY + SELF-HEALING            Signal: 1,216
      No system uses knowledge graph for self-healing decisions.
      BUILD: On build failure, graph shows which recent changes
      caused similar failures -> targeted rollback.

  E.  GRAPH MEMORY + SKILL EVOLUTION          Signal: 1,216
      No system connects skill learning to a knowledge graph.
      BUILD: Skills indexed by semantic similarity. When one
      evolves, graph propagation updates all related skills.

========================================================================
  BUILD ORDER (by evidence strength)
========================================================================

  Phase 3A — SKILL EVOLUTION (37x)
    The single strongest differentiator. SKILL.md templates per domain
    + /evolve command that tracks win rates and auto-optimizes prompts.

  Phase 3B — CONTEXT RE-INJECTION (20x) + TOKEN REDUCTION (10x)
    Lifecycle hooks + progressive schema disclosure. Together they
    solve the context window budget problem.

  Phase 3C — SELF-HEALING + PLANNER-CHECKER (15x + 13x)
    Build -> fail -> error-to-LLM -> fix -> retry.
    Plus Plan Mode: premium model strategizes, cheap model executes.

  Phase 3D — THREE-TIER MEMORY (14x)
    Hot/Warm/Cold with heat-based promotion. Makes everything above
    learn over time instead of starting fresh each session.

  Phase 3E — THE UNDISCOVERED COMBINATION: GRAPH + HITL GATES
    The strongest zero-co-occurrence pair. No existing system has this.
    Knowledge graph maps dependencies between extracted facts, then
    uses it for smarter approval/rejection decisions.
========================================================================
"""
