# TORMENTNEXUS FEATURE ASSESSMENT & NEXT PRIORITIES
*Derived from: (1) TormentNexus codebase audit (2) 13,503 bookmark ecosystem intelligence*
*Date: 2026-05-08 | Version: 1.0.0-alpha.53*

---

## WHAT TORMENTNEXUS IS (Actual, Not Aspirational)

TormentNexus is a **local-first cognitive control plane** — a Go+TypeScript modular monolith that coordinates multi-agent LLM workflows, MCP tool routing, provider failover, memory persistence, and operator observability. It currently runs as two cooperating processes:

- **Go Sidecar** (port 4300): 232 Go files, 446 HTTP handler methods, 18K-line server
- **TypeScript Control Plane** (port 4100): 583 TS files, 4343-line MCPServer.ts core
- **Next.js Dashboard** (port 3000): 91 pages, 709 TS/TSX files
- **SQLite storage**: tormentnexus.db (35MB), .tormentnexus/agent_memory/ (14,726 memories), .tormentnexus/memory/ (13,478 contexts)

---

## EXISTING FUNCTIONALITY — What's Actually Built

### FULLY OPERATIONAL (Stable)

| Feature | Go | TS | Dashboard | Details |
|---------|:--:|:--:|:---------:|---------|
| **Progressive MCP Tool Routing** | ✅ | ✅ | ✅ | 6 meta-tools, ranked search, auto-load, LRU eviction, profiles |
| **MCP Decision System** | ✅ | ✅ | ✅ | Multi-signal scoring, silent high-confidence auto-load, deferred binary startup |
| **LLM Waterfall** | ✅ | ✅ | ✅ | NVIDIA → OpenRouter → LM Studio/Ollama cascade on 429/5xx |
| **Provider Catalog** | ✅ | ✅ | ✅ | Google, Anthropic, OpenAI, DeepSeek, OpenRouter, GitHub Copilot routing |
| **MCP Config Management** | ✅ | ✅ | ✅ | CRUD for mcp.jsonc + DB tools, sync targets, export/import |
| **MCP Catalog Ingestion** | ✅ | ✅ | ✅ | 5 adapters: Glama, Smithery, MCP.run, npm, GitHub Topics |
| **MCP Server Pool** | — | ✅ | ✅ | Multi-process supervision with PID tracking |
| **Session Import/Export** | ✅ | ✅ | ✅ | Claude Code, Cursor, Aider, Windsurf, Copilot format detection |
| **CLI (31 commands)** | — | ✅ | — | Full lifecycle for MCP, sessions, providers, knowledge, swarm |
| **Dashboard (91 pages)** | — | — | ✅ | Health, inspector, tools, catalog, memory, swarm, council, etc. |

### BUILT BUT IMMATURE (Beta / Experimental)

| Feature | Go | TS | Maturity | Gap |
|---------|:--:|:--:|:--------:|-----|
| **Tiered Memory (L1/L2)** | ✅ | ✅ | Beta | L1 Scratchpad + L2 Vault work, but **no heat-based promotion**, no adaptive forgetting, no L3 cold archive. 14,726 memories but all `long_term/project` — no real tiering in practice |
| **Knowledge Graph** | ✅ | ✅ | Stub | `@tormentnexus/memory` exports `GraphNode`/`GraphEdge` **interfaces only** — actual implementations are `undefined`. `RepoGraphService` builds file-level import graphs but NOT semantic entity graphs |
| **Context Harvester** | ✅ | ✅ | Beta | Harvest/prune/compact/rerank pipeline works. **Groomer is rudimentary** (token estimation, system message preservation only). No semantic compaction via LLM |
| **Healer (Self-Healing)** | ✅ | ✅ | Beta | Error → LLM diagnosis → fix suggestion pipeline works. **No closed-loop** (execute fix → verify → retry). No stop hooks. No idle-state healing |
| **Swarm Controller** | ✅ | ✅ | Experimental | Role rotation (Planner/Implementer/Tester/Critic) and shared transcript. **No real consensus**, no task bidding, no completion detection by ExpertSupervisor in production |
| **Pair Orchestrator** | ✅ | — | Experimental | State machine: Planner→Reviewer→Implementer→Reviewer→Critic. Works in Go. **Not wired to actual agent sessions** |
| **A2A Broker** | ✅ | ✅ | Experimental | Message routing + heartbeat + query pattern + audit logging. **No real multi-process agents** — everything is in-process |
| **Council/Debate** | ✅ | ✅ | Experimental | Collaborative debate manager, rotation room, human veto service. **Self-evolution only adjusts weights** — no prompt evolution, no skill evolution |
| **Skill Registry** | ✅ | ✅ | Beta | SKILL.md discovery + CRUD + search. **No /evolve command**, no win-rate tracking, no auto-retirement. 0 skills in `.tormentnexus/skills/` |
| **Skill Assimilation** | — | ✅ | Experimental | Researches topic → generates SKILL.md via LLM. **No feedback loop** — skills never improve from usage |
| **Darwin (Self-Modification)** | ✅ | ✅ | Experimental | Prompt mutation + A/B testing. **No integration with skills or memory**. Runs in isolation |
| **Code Executor** | ✅ | ✅ | Beta | Multi-language sandbox with timeouts. WASM sandbox scaffolded but **uses exec.Command fallback** |
| **Repo Graph** | ✅ | ✅ | Beta | File-level import dependency graph. **No symbol-level graph**, no callers/callees, no type hierarchy |
| **Deep Research** | — | ✅ | Beta | Web search → LLM synthesis pipeline. Works but **no BobbyBookmarks integration** for ecosystem intelligence |
| **BobbyBookmarks Sync** | ✅ | ✅ | — | LinkCrawler + high-value ingestor. **Not feeding back into tool ranking or skill decisions** |

### NOT BUILT (Vision / TODO Only)

| Feature | Status | Notes |
|---------|--------|-------|
| **Progressive Skill Disclosure** | Listed in TODO | Same LRU/ranking architecture as tools, but never applied to skills |
| **Supervisor Tool Prediction** | Partial | ToolPredictor exists, but **not wired into active chat prompt chain** |
| **Browser Extension** | Stub package | MemoryCaptureService stubbed, no real injection into web chats |
| **Free-Tier Fallback Chain** | Partial | Only Gemini 2.5 Flash. No OpenRouter Free, no Google AI Studio native |
| **OAuth Login** | — | No Claude Max/Pro, Copilot Premium, ChatGPT Plus OAuth |
| **Cursor/Windsurf/Kiro Parity** | — | Only Claude Code, Codex, Gemini CLI, OpenCode parity done |
| **WASM Sandbox** | Stub | `wasm_sandbox.go` scaffolded, falls back to `exec.Command` |
| **Graph Memory + HITL Gates** | — | Zero implementations anywhere in ecosystem |
| **Context Re-Injection** | — | No PostToolUse hooks, no pre-compaction injection |
| **Token Budget Manager** | — | ContextHarvester has maxTokenBudget but no per-tool allocation |
| **Cross-Session Persistence** | Partial | 14,726 memories survive restarts, but **no execution checkpoints, no architectural decisions stored** |
| **Memory Consolidation** | — | No working→long-term promotion with LLM summarization |

---

## ECOSYSTEM GAP ANALYSIS (From 13,503 Bookmarks)

The BobbyBookmarks database shows the AI engineering ecosystem is lopsided:

| Stack Layer | Systems | Saturation | TormentNexus Coverage |
|-------------|---------|-----------|---------------|
| Protocol (MCP, bridges) | 6,406 | 66% OVERBUILT | ✅ Strong |
| Agent Runtime | 5,059 | 52% OVERBUILT | ✅ Strong |
| Developer UX | 5,009 | 52% OVERBUILT | ✅ Strong (91 pages) |
| Intelligence/RAG | 3,516 | 36% adequate | ✅ Has DeepResearch |
| Context Engine | 3,328 | 34% adequate | 🟡 Harvester exists but shallow |
| Tools | 2,424 | 25% UNDERBUILT | ✅ Strong (MCP routing) |
| Infrastructure | 2,067 | 21% UNDERBUILT | 🟡 Code executor exists, WASM stub |
| **Verification** | **1,729** | **18% UNDERBUILT** | 🔴 Healer is beta, no closed loop |
| **Self-Modification** | **1,506** | **15% UNDERBUILT** | 🔴 Darwin is isolated, no feedback |
| **Memory** | **1,445** | **15% UNDERBUILT** | 🔴 Tiering is flat, graph is stub |

**Missing Combinations** (deficit vs expected co-occurrence):
- Self-Mod + Tools: 2.7% deficit → TormentNexus has tools but self-mod is disconnected
- Memory + Tools: 1.7% deficit → TormentNexus's memory doesn't inform tool selection
- Infra + Memory: 1.2% deficit → No execution state in memory

---

## NEXT MOST VIABLE FEATURES

Ranked by: (1) evidence from ecosystem data, (2) TormentNexus's existing foundation to build on, (3) leverage — how much it unlocks downstream

### TIER 1: HIGH EVIDENCE + EXISTING FOUNDATION = DO NOW

#### 1. REAL TIERED MEMORY WITH HEAT PROMOTION
**Evidence:** 14x enrichment, 0.29x saturation (massive gap), 37x for skill evolution which depends on this
**TormentNexus foundation:** L1 Scratchpad + L2 Vault already work. 14,726 memories already stored.
**Gap:** All memories are `long_term/project` — no real tiering. No heat scoring. No promotion/demotion.
**What to build:**
- Add `heat_score` (0-100) and `last_accessed_at` to every memory entry
- Heat increases on access, decays over time (exponential, configurable half-life)
- Promote: working memories with heat > 80 → long_term automatically
- Demote: long_term memories with heat < 20 → archive (L3)
- Consolidation: when working memories exceed cap, LLM summarizes → single long_term entry
- **Why first:** Every feature below (skill evolution, self-healing, graph memory) needs real memory tiering to work

#### 2. CLOSE THE SELF-HEALING LOOP
**Evidence:** 15x enrichment, 0.34x saturation
**TormentNexus foundation:** HealerService (TS) + healer.go (Go) already do error→diagnosis→fix suggestion
**Gap:** No execute-fix-verify-retry cycle. No stop hooks. No idle-state healing.
**What to build:**
- `heal_and_verify(error, file)`: diagnose → apply fix → run test → if fail, re-diagnose with error context (max 3 loops)
- `StopHook`: intercept before session ends, check if promises from earlier in session are fulfilled
- `IdleHealer`: when no active task, scan recent errors and attempt background fixes
- Wire healer output back into memory (store what failed, what fixed it)
- **Why second:** Makes the control plane actually autonomous. Currently it suggests fixes but can't apply them.

#### 3. PROGRESSIVE SKILL DISCOVERY (Apply MCP Routing to Skills)
**Evidence:** 37x enrichment (skill evolution), listed in TormentNexus's own TODO
**TormentNexus foundation:** MCP Decision System already has the exact architecture (ranked search, auto-load, LRU eviction, profiles). SkillRegistry already discovers SKILL.md files.
**Gap:** Skills have no ranking, no working set, no eviction, no profile-based boosting. All-or-nothing.
**What to build:**
- `SkillDecisionSystem` — exact mirror of `MCPDecisionSystem` but for skills
- 5-6 permanent meta-skills (`search_skills`, `load_skill`, `list_loaded_skills`, `unload_skill`)
- Ranked skill search with multi-signal scoring (same algorithm as toolSearchRanking.ts)
- Working set with LRU eviction (default: 8 loaded skills)
- Profile boosting: `repo-coding` profile boosts Git/Docker skills, `web-research` boosts search/scrape skills
- **Why third:** Directly addresses the 37x signal. Architecture already exists in tool routing — just apply it to skills.

### TIER 2: MEDIUM EVIDENCE + MODERATE FOUNDATION = DO NEXT

#### 4. CONTEXT RE-INJECTION AFTER COMPACTION
**Evidence:** 20x enrichment
**TormentNexus foundation:** ContextHarvester + Groomer already compact context. Memory hydration already injects L2 into L1 on session start.
**Gap:** No re-injection AFTER compaction. No PreToolUse/PostToolUse hooks.
**What to build:**
- `CompactionHook`: when Groomer compacts, immediately re-inject key facts from L2 Vault
- `PreToolUse` / `PostToolUse` lifecycle hooks: inject tool-relevant context before each tool call
- `TokenBudgetManager`: allocate context budget per tool (e.g., bash gets 2K, search gets 4K)
- Progressive schema disclosure: load full tool schema only on first use, keep stub descriptions otherwise
- **Why fourth:** Solves the context window problem that gets worse as more tools/skills are loaded.

#### 5. PLANNER-CHECKER-REVISE LOOP
**Evidence:** 12.7x enrichment
**TormentNexus foundation:** PairOrchestrator has the state machine. Council has debate/consensus. HumanVetoService exists.
**Gap:** PairOrchestrator isn't wired to real sessions. Council debates don't produce actionable plans. No "Plan Mode" where premium model strategizes before cheap model executes.
**What to build:**
- `PlanMode`: premium model (e.g., Claude Opus) creates PLAN.md, budget model (e.g., Gemini Flash) executes
- `CheckerAgent`: second model validates plan against codebase constraints before execution
- `ReviseLoop`: if checker rejects, loop back with rejection reason as context
- Wire PairOrchestrator state machine into actual agent sessions (currently standalone)
- **Why fifth:** Dramatically improves code quality without increasing cost (premium model used only for planning).

#### 6. MEMORY → TOOL SELECTION FEEDBACK LOOP
**Evidence:** Memory + Tools has 1.7% deficit in ecosystem — almost no system connects them
**TormentNexus foundation:** Both memory and tool selection are mature systems. They just don't talk.
**What to build:**
- `MemoryInformedRanking`: when searching tools, boost scores for tools that succeeded in similar past contexts
- `ToolOutcomeMemory`: after every tool call, store outcome (success/fail, latency, relevance) in memory
- `RankingFeedbackLoop`: periodically re-weight tool search signals based on accumulated outcomes
- This is the **Memory + Tools combination** the data says is missing from the entire ecosystem
- **Why sixth:** Low effort (both systems exist), high impact (no competitor has this).

### TIER 3: HIGH EVIDENCE BUT LARGER BUILDS = PLAN FOR

#### 7. REAL KNOWLEDGE GRAPH (Not Just File Imports)
**Evidence:** Graph Memory has the highest zero-co-occurrence with every other mechanism
**TormentNexus foundation:** RepoGraphService builds file-level graphs. KnowledgeService has graph interfaces. `@tormentnexus/memory` exports GraphNode/GraphEdge types.
**Gap:** All graph implementations are stubs or file-level only. No semantic entity graph.
**What to build:**
- Entity extraction: LLM identifies concepts (projects, tools, patterns, decisions) from memories
- Relationship edges: `uses_tool`, `depends_on`, `contradicts`, `evolved_from`, `fixed_by`
- Graph queries: "what tools do projects like this one typically use?"
- **Blast radius analysis**: before applying a fix, graph shows what else depends on the changed file/concept
- This enables the **Graph + HITL Gates** combination (signal: 1,984, zero implementations)
- **Why seventh:** Highest unique value but biggest build. Do memory tiering first so the graph has quality data.

#### 8. SKILL EVOLUTION WITH WIN-RATE TRACKING
**Evidence:** 37x enrichment (strongest signal in data)
**TormentNexus foundation:** SkillRegistry + SkillAssimilationService + DarwinService all exist in pieces.
**Gap:** Skills never improve. Darwin mutates prompts but doesn't touch skills. No /evolve command. No win/loss tracking.
**What to build:**
- `/evolve` command: takes a skill, runs it on 3 test cases, evaluates results, mutates prompt, re-tests
- Win-rate tracking: every skill execution records success/failure context
- Auto-retirement: skills with <30% win rate over last 20 uses get flagged
- Cross-agent skill sync: when one agent evolves a skill, broadcast via A2A so other agents see the update
- **Why eighth:** 37x signal but requires real memory tiering (#1) and progressive skill discovery (#3) first.

#### 9. GRAPH-MEMORY-INFORMED HITL GATES
**Evidence:** Graph + HITL Gates has signal 1,984 with ZERO implementations in 13,503 bookmarks
**TormentNexus foundation:** HumanVetoService exists. A2A broker works. KnowledgeService has graph types.
**Gap:** No system uses graph relationships to decide when to escalate to humans.
**What to build:**
- `BlastRadiusCalculator`: on every proposed action, query graph for dependent entities
- `AutoEscalationPolicy`: if blast radius > threshold (e.g., >5 dependent files, >2 active projects), require human approval
- `RiskScore`: combine blast radius + change type + historical failure rate → single risk number
- Low-risk changes auto-approve. High-risk changes gate with HumanVetoService.
- **Why ninth:** Uniquely valuable (no one has this) but needs real graph (#7) first.

---

## WHAT NOT TO BUILD

The data says these layers are OVERBUILT in the ecosystem. TormentNexus should **consume** them, not compete:

| Don't Build | Reason | Instead |
|-------------|--------|---------|
| Another agent framework | 52% saturated | TormentNexus's swarm/council is sufficient — focus on making it reliable |
| Another MCP server | 66% saturated | TormentNexus should ROUTE existing servers, not create new ones |
| Another dashboard framework | 52% saturated | 91 pages is enough — focus on data quality, not more pages |
| Another CLI harness | Crowded | TormentNexus should COORDINATE existing harnesses via parity aliases |
| P2P mesh / federation | Vision only | Not justified until single-node is rock solid |

---

## BUILD ORDER SUMMARY

```
Phase 1 — MEMORY FOUNDATION
  ├─ #1 Real Tiered Memory with Heat Promotion
  │    (foundation everything else needs)
  └─ #6 Memory → Tool Selection Feedback Loop
       (connects two existing systems, low effort)

Phase 2 — AUTONOMY LOOP  
  ├─ #2 Close the Self-Healing Loop
  │    (makes the control plane actually autonomous)
  └─ #5 Planner-Checker-Revise Loop
       (improves output quality without cost increase)

Phase 3 — SKILL INTELLIGENCE
  ├─ #3 Progressive Skill Discovery
  │    (apply proven MCP routing architecture to skills)
  └─ #4 Context Re-Injection After Compaction
       (solves the context window budget problem)

Phase 4 — THE UNIQUE VALUE
  ├─ #7 Real Knowledge Graph
  │    (semantic entity graph, not just file imports)
  ├─ #8 Skill Evolution with Win-Rate Tracking
  │    (37x signal, requires #1 + #3 first)
  └─ #9 Graph-Memory-Informed HITL Gates
       (signal 1,984, zero implementations, needs #7)
```
