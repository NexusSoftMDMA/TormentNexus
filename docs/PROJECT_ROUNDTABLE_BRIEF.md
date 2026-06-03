# tormentnexus 0.9.1: Project Roundtable Brief

## Introduction
This document serves as the master briefing for all AI models (Claude 3.7 Sonnet, GPT-4o, Gemini 2.5 Pro, and others) evaluating, debating, or developing the tormentnexus Control Plane. It outlines the absolute state of the repository, what works, what is planned, and our philosophical boundaries.

## 1. The Core Philosophy
tormentnexus is a **Cognitive Control Plane** for AI agents. It sits between the agent (the brain) and the infrastructure (the tools, the files, the commands). 

Our primary mandate is: **Resistance to inefficiency is futile.**
- **Token Economy**: We do not overwhelm the LLM with 500+ tool schemas. We advertise latent Meta-Tools.
- **Persistent Memory**: Agents do not start from scratch. They have access to a continuous graph of knowledge.
- **Autonomy**: Background systems heal crashes, harvest knowledge, and predict user intent without being explicitly asked.

## 2. What is Done & Working (v0.9.1)

### 🧩 The Meta-Tool Architecture
- **`search_tools` & `load_tool`**: Working perfectly. Agents can discover tools semantically and hydrate them on-demand.
- **`auto_call_tool`**: The "One-Shot" discovery and execution mechanism is fully implemented, eliminating multi-turn latency.
- **Tool Limits**: Hard-capped at 450 advertised tools to respect Gemini's 512 function declaration limit. Tools are hidden unless marked `alwaysOn: true` or explicitly loaded.

### 🧠 The Cognitive Engine
- **Agent Memory System (LanceDB)**: Persists session context, working memory, and long-term knowledge.
- **Session Handoff/Pickup**: Operators can freeze a session state to a JSON artifact and restore it later. Auto-pickup restores the latest session automatically upon restart.
- **Context Compacting (`ContextPruner`)**: Manually and autonomously compresses chat history to prevent context rot.
- **Sensory Harvesting (`MemoryHarvestReactor`)**: Listens to file changes and semantically extracts architectural knowledge in the background.

### 🌐 The Web Dashboard (Mission Control)
- **Next.js UI**: Live at `http://localhost:3001/dashboard`.
- **Function Toggles**: Real-time enabling/disabling of "Always On" tools.
- **Neural Pulse**: A live WebSocket-driven feed of background cognitive events (harvesting, predicting, healing).
- **Context Health**: Live token counting and status probes for local LLMs (Ollama/LM Studio).
- **Cognitive Intake**: Manual UI for pasting docs/notes into the long-term vector store.
- **Project Constitution**: Editable UI for the master `project_context.md` laws.

### 🔌 Integrations & Client Bridge
- **Browser Extension**: `apps/tormentnexus-extension` is a 1:1 replacement for the legacy MCP-SuperAssistant. It captures web pages automatically and exposes the browser DOM to tormentnexus via WebSocket.
- **Local-First Watcher**: The `SuggestionService` proactively predicts tool needs based on chat history, prioritizing local models to save API costs.
- **Self-Healing (`HealerReactor`)**: Autonomously analyzes and attempts to fix terminal crashes.

## 3. What is Planned (The Road to v1.0)

1. **Full Swarm Orchestration**: Currently, `Director` and `Council` exist but are experimental. We need robust horizontal scaling of specialized agents (Coder, Researcher, Reviewer) working in parallel.
2. **Native Sandbox Isolation**: Moving beyond simple `child_process` execution to a fully containerized (Docker/WASM) execution environment for `run_code` and `bash` tools.
3. **Advanced Graph Visualization**: The `GraphWidget` is functional but needs interactive, deeply explorable D3.js/React-Flow nodes for the entire codebase memory.
4. **Universal Agent Attach**: The ability to seamlessly attach tormentnexus to a running `claude` CLI or `cursor` editor session and inject memory directly into their prompts.

## 4. What We Are NOT Doing (Explicitly Out of Scope)
- **Replacing Editors**: We are not building an IDE. We are building the control plane that *connects* to IDEs.
- **Vendor Lock-in**: We will never require OpenAI or Anthropic exclusively. The `ModelSelector` fallback chain is sacred.
- **Bloated Tool Contexts**: We will never revert to injecting 500 tool schemas into the system prompt. The Meta-Tool pattern is non-negotiable.

## 5. The 1:1 Compatibility Mandate
Models are fine-tuned on specific tools (e.g., Claude Code's `bash`, `glob`, `str_replace_editor`). tormentnexus intercepts these exact schemas and routes them to our secure implementations. **We do not alter these signatures.** A model running on tormentnexus should feel like it is running in its native training environment, but with superpowers.

---
**Debate Prompt for AI Models:**
*Analyze the architecture above. Are there failure points in the Meta-Tool pattern? How can the Semantic Harvesting be optimized without burning background tokens? What is the ideal consensus protocol for the planned Swarm Orchestration?*
