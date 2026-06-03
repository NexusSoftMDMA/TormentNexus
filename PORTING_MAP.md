# 🧪 Phase P: FULL ASSIMILATION Porting Map

This document outlines the strategic migration of logic from reference submodules into the native **@tormentnexus** monorepo packages.
This document outlines the strategic migration of logic from reference submodules into the native **@tormentnexus** monorepo packages.

## 🧬 Core Strategy
Move from "integration" (wrapping external tools) to "assimilation" (native implementation). This enables:
1. **Total Autonomy**: No reliance on external repository states or drift.
2. **Infinite Context**: Direct integration with TormentNexus's internal session and memory managers.
2. **Infinite Context**: Direct integration with TormentNexus's internal session and memory managers.
3. **Unified Performance**: Single-process orchestration without IPC bottlenecks.

---

## 🛰️ Jules-Autopilot -> @tormentnexus/agents & @tormentnexus/core
## 🛰️ Jules-Autopilot -> @tormentnexus/agents & @tormentnexus/core

| Logic Component | Source File | Target Location | Rationale |
| :--- | :--- | :--- | :--- |
| **Risk Scoring** | `packages/shared/src/orchestration/supervisor.ts` | `packages/agents/src/orchestration/RiskEvaluator.ts` | Native safety gating for autonomous changes. |
| **Multi-Agent Debate** | `packages/shared/src/orchestration/debate.ts` | `packages/agents/src/orchestration/DebateEngine.ts` | High-fidelity consensus mechanism for the TormentNexus Council. |
| **Multi-Agent Debate** | `packages/shared/src/orchestration/debate.ts` | `packages/agents/src/orchestration/DebateEngine.ts` | High-fidelity consensus mechanism for the TormentNexus Council. |
| **Conference Logic** | `packages/shared/src/orchestration/debate.ts` | `packages/agents/src/orchestration/ConferenceManager.ts` | Team-wide sync points for complex plan validation. |
| **Provider Wrappers** | `packages/shared/src/orchestration/providers/*` | `packages/ai/src/providers/*` | Unify AI provider logic (Gemini, Anthropic, OpenAI). |

---

## 🔖 BobbyBookmarks -> @tormentnexus/core (Memory)
## 🔖 BobbyBookmarks -> @tormentnexus/core (Memory)

| Logic Component | Source File | Target Location | Rationale |
| :--- | :--- | :--- | :--- |
| **Research Worker** | `research.py` (Port to TS) | `packages/core/src/services/Memory/ResearchWorker.ts` | Background enrichment of session memories and links. |
| **Metadata Extraction** | `research.py` (BeautifulSoup -> linkedom) | `packages/core/src/utils/MetadataExtractor.ts` | Automated title/desc/favicon harvesting for RAG. |
| **LLM Tagger** | `tagger.py` (Port to TS) | `packages/core/src/services/Memory/AutoTagger.ts` | Semantic classification of all ingested knowledge. |
| **DB Sync Logic** | `sync_dbs.py` | `packages/core/src/services/Memory/PeerSync.ts` | Distributed memory synchronization across TormentNexus nodes. |

---

## 🏛️ Maestro -> @tormentnexus/core & @tormentnexus/ui
| **DB Sync Logic** | `sync_dbs.py` | `packages/core/src/services/Memory/PeerSync.ts` | Distributed memory synchronization across TormentNexus nodes. |

---

## 🏛️ Maestro -> @tormentnexus/core & @tormentnexus/ui

| Logic Component | Source File | Target Location | Rationale |
| :--- | :--- | :--- | :--- |
| **Agent Detector** | `src/main/ipc/handlers/agents.ts` | `packages/core/src/services/AgentDiscovery.ts` | Native detection of local AI tools (Claude Code, etc.). |
| **Process Manager** | `src/main/process-manager/ProcessManager.ts` | `packages/core/src/services/ProcessManager.ts` | Advanced PTY/CLI process orchestration with native input. |
| **Context Groomer** | `src/main/utils/context-groomer.ts` | `packages/core/src/services/Context/Groomer.ts` | Automated context compression and summarization. |
| **Visual Orchestrator** | `src/renderer/components/VisualOrchestrator/*` | `packages/ui/src/components/Orchestrator/*` | Native React Flow visualization of agent execution. |
| **Director Notes** | `src/main/ipc/handlers/director-notes.ts` | `packages/core/src/services/Director/Notes.ts` | AI-generated high-level summaries of complex workflows. |

---

## 🚀 Execution Order
1. **Foundation**: Port AI Provider unification from Jules-Autopilot to `@tormentnexus/ai`.
2. **Intelligence**: Port Debate and Risk logic to `@tormentnexus/agents`.
3. **Memory**: Port Research and Tagging logic to `@tormentnexus/core/Memory`.
4. **Interface**: Port Process Management and Agent Discovery to `@tormentnexus/core`.
5. **Visualization**: Port React Flow components to `@tormentnexus/ui`.
1. **Foundation**: Port AI Provider unification from Jules-Autopilot to `@tormentnexus/ai`.
2. **Intelligence**: Port Debate and Risk logic to `@tormentnexus/agents`.
3. **Memory**: Port Research and Tagging logic to `@tormentnexus/core/Memory`.
4. **Interface**: Port Process Management and Agent Discovery to `@tormentnexus/core`.
5. **Visualization**: Port React Flow components to `@tormentnexus/ui`.

**DON'T STOP THE PARTY. THE COLLECTIVE GROWS.**
