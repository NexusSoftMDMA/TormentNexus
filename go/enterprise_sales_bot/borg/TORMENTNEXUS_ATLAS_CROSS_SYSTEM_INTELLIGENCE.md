# TORMENTNEXUS x ATLAS CROSS-SYSTEM INTELLIGENCE REPORT
_Generated from live database cross-reference_

---

## 1. SYSTEM INVENTORY

| System | Database | Total Entries | GitHub Repos | High-Signal (>=85) |
|--------|----------|--------------|-------------|-------------------|
| **Atlas** | atlas.db | 7,944 | 4,919 | 3,632 |
| **TormentNexus Backlog** | tormentnexus.db links_backlog | 15,753 | 6,116 | N/A |
| **TormentNexus Tools** | tormentnexus.db tools | 651 | N/A | N/A |
| **TormentNexus MCP Servers** | tormentnexus.db mcp_servers | 68 | N/A | N/A |
| **TormentNexus Sessions** | tormentnexus.db imported_sessions | 9,774 | N/A | N/A |
| **TormentNexus Memories** | tormentnexus.db imported_session_memories | 22,749 | N/A | N/A |

### Cross-Reference Overlap

- **Shared repos** (in both systems): **4,664**
- **Atlas-only** repos: **255** -- candidates for TormentNexus assimilation
- **TormentNexus-only** repos: **1,452** -- candidates for Atlas ingestion

---

## 2. TORMENTNEXUS CODEBASE STATUS vs FEATURE ASSESSMENT

Based on audit of 231 Go files + 583 TS files + 91 dashboard pages:

| Feature | Status | Go | TS | Key Gap |
|---------|--------|:--:|:--:|---------|
| Progressive MCP Tool Routing | STABLE | Y | Y | None |
| LLM Waterfall | STABLE | Y | Y | None |
| Session Import/Export | STABLE | Y | Y | None |
| MCP Catalog Ingestion | STABLE | Y | Y | None |
| Tiered Memory L1/L2 | BETA | Y | Y | Heat schema exists but no L3 archive, no consolidation |
| Healer (Self-Healing) | BETA | Y | Y | HealAndVerify loop EXISTS. Missing StopHook, IdleHealer |
| Skill Decision System | BETA | Y | - | SearchAndLoad+LRU works. Missing win-rate SQLite persistence |
| Skill Evolution | BETA | Y | - | EvolveSkill+RecordOutcome exist. Missing auto-retirement |
| Context Harvester | BETA | Y | Y | No LLM-based semantic compaction |
| Knowledge Graph | STUB | Y | Y | GraphNode/GraphEdge interfaces only, undefined impls |
| PairOrchestrator | EXP | Y | - | State machine works, not wired to real sessions |
| Swarm Controller | EXP | Y | Y | Role rotation works, no real consensus |
| A2A Broker | EXP | Y | Y | Message routing works, no multi-process agents |
| Council/Debate | EXP | Y | Y | Debate manager works, no skill/prompt evolution |
| WASM Sandbox | STUB | Y | - | Falls back to exec.Command |
| Browser Extension | STUB | - | Y | MemoryCaptureService stubbed only |
| Graph+HITL Gates | NONE | - | - | Zero implementations |

---

## 3. TOP ASSIMILATION CANDIDATES (Atlas -> TormentNexus)

From 255 Atlas-only repos, ranked by signal x innovation x architecture gap alignment:

### Memory & Tiering (49 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [keli-wen/agentic-harness-patterns-skill](https://github.com/keli-wen/agentic-harness-patterns-skill) | 90 | 10 | 308 | Agent skill for harness engineering — memory, permissions, contex... |
| 2 | [justnau1020/claude-os](https://github.com/justnau1020/claude-os) | 92 | 10 | 293 | A framework for Claude Code that enhances agent orchestration, co... |
| 3 | [Kalki-M/BlackSwanX](https://github.com/Kalki-M/BlackSwanX) | 98 | 10 | 288 | A decentralized AI intelligence platform that leverages a diverse... |
| 4 | [agenteractai/lodmem](https://github.com/agenteractai/lodmem) | 90 | 9 | 268 | LODM is a memory tool that organizes session data into structured... |
| 5 | [vstorm-co/pydantic-deepagents?referrer=grok.com](https://github.com/vstorm-co/pydantic-deepagents?referrer=grok.com) | 95 | 9 | 266 | A modular framework for building autonomous agents that implement... |
| 6 | [yantrikos/yantrikdb](https://github.com/yantrikos/yantrikdb) | 95 | 9 | 266 | Cognitive memory engine for AI agents — temporal decay, contradic... |
| 7 | [RealZST/HarnessKit](https://github.com/RealZST/HarnessKit) | 95 | 10 | 265 | More than a skill manager — manage skills, MCP servers, plugins, ... |
| 8 | [nambok/mentedb](https://github.com/nambok/mentedb) | 91 | 9 | 264 | MenteDB is a purpose-built Rust-based storage engine designed for... |
| | _...and 41 more_ | | | | |

### Self-Healing (4 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [browser-use/browser-harness](https://github.com/browser-use/browser-harness) | 89 | 10 | 237 | A self-healing browser harness enabling LLMs to complete tasks ac... |
| 2 | [Garrus800-stack/genesis-agent](https://github.com/Garrus800-stack/genesis-agent) | 84 | 10 | 236 | Self-aware cognitive AI agent that reads, modifies &amp; verifies... |
| 3 | [metedata/pdf-proof](https://github.com/metedata/pdf-proof) | 87 | 9 | 230 | A Claude skill that visualizes AI-generated proof by highlighting... |
| 4 | [aurite-ai/agent-verifier](https://github.com/aurite-ai/agent-verifier) | 81 | 9 | 194 | This repository implements a framework for verifying and managing... |

### Skill Evolution (21 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [mcpware/cross-code-organizer](https://github.com/mcpware/cross-code-organizer) | 98 | 10 | 278 | Cross-Code Organizer (formerly Claude Code Organizer): cross-harn... |
| 2 | [RealZST/HarnessKit](https://github.com/RealZST/HarnessKit) | 95 | 10 | 265 | More than a skill manager — manage skills, MCP servers, plugins, ... |
| 3 | [Bitterbot-AI/bitterbot-desktop](https://github.com/Bitterbot-AI/bitterbot-desktop) | 93 | 10 | 261 | Bitterbot is a local-first AI agent designed for persistent memor... |
| 4 | [yitianlian/harnessbridge](https://github.com/yitianlian/harnessbridge) | 91 | 10 | 261 | Portable agent harness configuration. Convert rules, skills, hook... |
| 5 | [multica-ai/multica](https://github.com/multica-ai/multica) | 95 | 10 | 235 | Multica turns coding agents into real teammates by assigning task... |
| 6 | [metedata/pdf-proof](https://github.com/metedata/pdf-proof) | 87 | 9 | 230 | A Claude skill that visualizes AI-generated proof by highlighting... |
| 7 | [Leonxlnx/taste-skill](https://github.com/Leonxlnx/taste-skill) | 88 | 9 | 228 | A frontend framework designed to enhance AI agent interfaces with... |
| 8 | [gotalab/cc-sdd](https://github.com/gotalab/cc-sdd) | 89 | 10 | 227 | Turn approved specs into long-running autonomous implementation. ... |
| | _...and 13 more_ | | | | |

### Context Engineering (31 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [keli-wen/agentic-harness-patterns-skill](https://github.com/keli-wen/agentic-harness-patterns-skill) | 90 | 10 | 308 | Agent skill for harness engineering — memory, permissions, contex... |
| 2 | [justnau1020/claude-os](https://github.com/justnau1020/claude-os) | 92 | 10 | 293 | A framework for Claude Code that enhances agent orchestration, co... |
| 3 | [Kalki-M/BlackSwanX](https://github.com/Kalki-M/BlackSwanX) | 98 | 10 | 288 | A decentralized AI intelligence platform that leverages a diverse... |
| 4 | [mcpware/cross-code-organizer](https://github.com/mcpware/cross-code-organizer) | 98 | 10 | 278 | Cross-Code Organizer (formerly Claude Code Organizer): cross-harn... |
| 5 | [agenteractai/lodmem](https://github.com/agenteractai/lodmem) | 90 | 9 | 268 | LODM is a memory tool that organizes session data into structured... |
| 6 | [vstorm-co/pydantic-deepagents?referrer=grok.com](https://github.com/vstorm-co/pydantic-deepagents?referrer=grok.com) | 95 | 9 | 266 | A modular framework for building autonomous agents that implement... |
| 7 | [yantrikos/yantrikdb](https://github.com/yantrikos/yantrikdb) | 95 | 9 | 266 | Cognitive memory engine for AI agents — temporal decay, contradic... |
| 8 | [nambok/mentedb](https://github.com/nambok/mentedb) | 91 | 9 | 264 | MenteDB is a purpose-built Rust-based storage engine designed for... |
| | _...and 23 more_ | | | | |

### Knowledge Graph (10 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [sachitrafa/YourMemory](https://github.com/sachitrafa/YourMemory) | 87 | 9 | 250 | YourMemory introduces persistent memory for AI agents, enabling t... |
| 2 | [aouicher/graphmind](https://github.com/aouicher/graphmind) | 93 | 9 | 243 | A tool that transforms codebases into interactive knowledge graph... |
| 3 | [safishamsi/graphify](https://github.com/safishamsi/graphify) | 92 | 10 | 237 | A tool that integrates with Claude Code and Gemini CLI to leverag... |
| 4 | [recallium/recallium](https://github.com/recallium/recallium) | 78 | 10 | 228 | Recallium: Universal Memory |
| 5 | [ousatov-ua/memgraph-ingester](https://github.com/ousatov-ua/memgraph-ingester/blob/main/README.md) | 87 | 9 | 225 | Ingester of Java structure in Memgraph. Speed up your AI agent! -... |
| 6 | [neo4j/mcp-neo4j](https://github.com/neo4j/mcp-neo4j) | 78 | 10 | 208 | Neo4j MCP: GraphRAG |
| 7 | [goharbor/harbor](https://github.com/goharbor/harbor) | 91 | 10 | 207 | Harbor is a cloud-native registry that extends Docker distributio... |
| 8 | [rafaskb/awesome-libgdx#readme](https://github.com/rafaskb/awesome-libgdx) | 89 | 9 | 199 | A curated list of libGDX resources to help developers make awesom... |
| | _...and 2 more_ | | | | |

### Agent Orchestration (58 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [VibePod/vibepod-cli](https://github.com/VibePod/vibepod-cli) | 95 | 9 | 256 | A unified CLI for running AI coding agents in isolated Docker con... |
| 2 | [1jehuang/jcode#detailed-installation](https://github.com/1jehuang/jcode) | 93 | 10 | 253 | A next-generation coding agent harness designed to enhance develo... |
| 3 | [jazzenchen/VibeAround](https://github.com/jazzenchen/VibeAround) | 98 | 10 | 253 | VibeAround integrates multiple AI coding agents into a unified wo... |
| 4 | [hyspacex/harness-cli](https://github.com/hyspacex/harness-cli) | 93 | 10 | 249 | Harness CLI orchestrates AI agents in a structured workflow to bu... |
| 5 | [chauncygu/collection-claude-code-source-code](https://github.com/chauncygu/collection-claude-code-source-code) | 90 | 10 | 246 | This repository contains a comprehensive collection of Claude Cod... |
| 6 | [RMANOV/sqlite-memory-mcp](https://github.com/RMANOV/sqlite-memory-mcp) | 91 | 9 | 239 | A production-grade SQLite-based memory management stack for Claud... |
| 7 | [multica-ai/multica](https://github.com/multica-ai/multica) | 95 | 10 | 235 | Multica turns coding agents into real teammates by assigning task... |
| 8 | [wesm/agentsview](https://github.com/wesm/agentsview) | 95 | 10 | 235 | Local-first session intelligence and analytics for coding agents,... |
| | _...and 50 more_ | | | | |

### Harness Integration (39 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [cloveric/cc-telegram-bridge](https://github.com/cloveric/cc-telegram-bridge) | 98 | 10 | 238 | Real Claude Code &amp; Codex CLI on Telegram — native CLI harness... |
| 2 | [safishamsi/graphify](https://github.com/safishamsi/graphify) | 92 | 10 | 237 | A tool that integrates with Claude Code and Gemini CLI to leverag... |
| 3 | [browser-use/browser-harness](https://github.com/browser-use/browser-harness) | 89 | 10 | 237 | A self-healing browser harness enabling LLMs to complete tasks ac... |
| 4 | [wesm/agentsview](https://github.com/wesm/agentsview) | 95 | 10 | 235 | Local-first session intelligence and analytics for coding agents,... |
| 5 | [Leonxlnx/taste-skill](https://github.com/Leonxlnx/taste-skill) | 88 | 9 | 228 | A frontend framework designed to enhance AI agent interfaces with... |
| 6 | [gotalab/cc-sdd](https://github.com/gotalab/cc-sdd) | 89 | 10 | 227 | Turn approved specs into long-running autonomous implementation. ... |
| 7 | [aayoawoyemi/Aries-cli](https://github.com/aayoawoyemi/Aries-cli) | 84 | 10 | 226 | Agentic coding harness with persistent memory and a REPL body. Bu... |
| 8 | [edwarddgao/agent-traces](https://github.com/edwarddgao/agent-traces) | 91 | 9 | 226 | Agent-friendly semantic search over your local Claude Code and Co... |
| | _...and 31 more_ | | | | |

### Code Execution / Sandbox (5 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [Infisical/agent-vault](https://github.com/Infisical/agent-vault) | 95 | 10 | 240 | A secure HTTP credential proxy and vault for AI agents, enabling ... |
| 2 | [cloveric/cc-telegram-bridge](https://github.com/cloveric/cc-telegram-bridge) | 98 | 10 | 238 | Real Claude Code &amp; Codex CLI on Telegram — native CLI harness... |
| 3 | [patrickdappollonio/dux](https://github.com/patrickdappollonio/dux) | 88 | 9 | 196 | A terminal UI for managing multiple AI coding agents in parallel ... |
| 4 | [agentspan-ai/agentspan](https://github.com/agentspan-ai/agentspan) | 88 | 9 | 196 | A distributed runtime for AI agents that ensures durability, cras... |
| 5 | [phr00t/FocusEngine?tab=readme-ov-file](https://github.com/phr00t/FocusEngine?tab=readme-ov-file) | 68 | 9 | 190 | Context Engineering & Isolation |

### MCP Infrastructure (3 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [janbjorge/rekal](https://github.com/janbjorge/rekal) | 89 | 9 | 232 | Rekal is a local SQLite-based MCP server that enables persistent ... |
| 2 | [open-pgx/openpgx](https://github.com/open-pgx/openpgx) | 93 | 10 | 203 | OpenPGx enables AI-driven pharmacogenomic analysis by providing a... |
| 3 | [navbuildz/gmail-mcp-server](https://github.com/navbuildz/gmail-mcp-server) | 92 | 10 | 198 | Enables AI agents and assistants to manage multiple Gmail account... |

### Security / HITL (7 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [Kretski/MicroSafe-RL](https://github.com/Kretski/MicroSafe-RL) | 95 | 10 | 220 | A lightweight, model-free safety engine for Reinforcement Learnin... |
| 2 | [modelcontextprotocol/specification](https://github.com/modelcontextprotocol/specification) | 91 | 9 | 216 | This repository provides the specification and documentation for ... |
| 3 | [microsoft/lib0xc](https://github.com/microsoft/lib0xc) | 78 | 8 | 205 | A project focused on enhancing C language safety through safer sy... |
| 4 | [Exocija/ZetaLib](https://github.com/Exocija/ZetaLib/blob/main/The%20Gay%20Jailbreak/The%20Gay%20Jailbreak.md) | 87 | 9 | 202 | Explores a novel jailbreak technique leveraging AI-generated pers... |
| 5 | [govctl-org/govctl](https://github.com/govctl-org/govctl) | 66 | 9 | 189 | A governance harness for AI coding. Contribute to govctl-org/govc... |
| 6 | [kotarimorm/-Report-AI-coding-agent-programmatically-bypassing-OS-security-policies-Trace-ID-f4b806d4...-](https://github.com/kotarimorm/-Report-AI-coding-agent-programmatically-bypassing-OS-security-policies-Trace-ID-f4b806d4...-/blob/main/README.md) | 80 | 8 | 187 | Describes a technical report on an AI agent that bypasses OS secu... |
| 7 | [usewombat/gateway](https://github.com/usewombat/gateway) | 78 | 8 | 170 | Resource-level permissions for MCP agents: rwxd on any resource, ... |

### Search / RAG (24 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [kitfunso/hippo-memory](https://github.com/kitfunso/hippo-memory) | 98 | 10 | 238 | A biologically-inspired memory system for AI agents that enables ... |
| 2 | [WhitehatD/crag](https://github.com/WhitehatD/crag) | 95 | 10 | 231 | crag enables automated governance and workflow orchestration for ... |
| 3 | [sqliteai/sqlite-memory](https://github.com/sqliteai/sqlite-memory) | 95 | 10 | 230 | SQLite-Memory provides a persistent, searchable memory solution f... |
| 4 | [edwarddgao/agent-traces](https://github.com/edwarddgao/agent-traces) | 91 | 9 | 226 | Agent-friendly semantic search over your local Claude Code and Co... |
| 5 | [Irina1920/WMB-100K](https://github.com/Irina1920/WMB-100K) | 91 | 9 | 224 | A benchmark evaluating AI memory systems' ability to retrieve acc... |
| 6 | [vishalveerareddy123/Lynkr?utm_source=chatgpt.com](https://github.com/vishalveerareddy123/Lynkr) | 95 | 9 | 221 | A self-hosted universal LLM proxy that enables proprietary AI cod... |
| 7 | [Kretski/MicroSafe-RL](https://github.com/Kretski/MicroSafe-RL) | 95 | 10 | 220 | A lightweight, model-free safety engine for Reinforcement Learnin... |
| 8 | [mem0ai/mcp-mem0](https://github.com/mem0ai/mcp-mem0) | 78 | 10 | 218 | Mem0 + Qdrant: Semantic Memory |
| | _...and 16 more_ | | | | |

### Browser Use (5 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [Tanq16/local-content-share?tab=readme-ov-file](https://github.com/Tanq16/local-content-share?tab=readme-ov-file) | 95 | 9 | 206 | Self-hosted app with browser frontend that enables sharing and st... |
| 2 | [DustinBrett/daedalOS?tab=readme-ov-file](https://github.com/DustinBrett/daedalOS?tab=readme-ov-file) | 84 | 9 | 202 | Desktop environment in the browser |
| 3 | [algonius/algonius-browser?tab=readme-ov-file](https://github.com/algonius/algonius-browser?tab=readme-ov-file) | 95 | 9 | 196 | An open-source MCP server that bridges external AI systems to Chr... |
| 4 | [elebumm/RedditVideoMakerBot](https://github.com/elebumm/RedditVideoMakerBot) | 90 | 9 | 190 | Automates video creation on Reddit with a single command. |
| 5 | [robinovitch61/jeeves](https://github.com/robinovitch61/jeeves) | 88 | 9 | 186 | A powerful AI agent conversation history browser for developers. |

### Session / Transcript (1 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [BasedHardware/omi](https://github.com/BasedHardware/omi) | 90 | 10 | 210 | AI-powered software that integrates screen monitoring, conversati... |

### Observability (16 candidates)

| # | Repo | Sig | Inn | Score | Description |
|---|------|-----|-----|-------|-------------|
| 1 | [Krixx1337/burner-net](https://github.com/Krixx1337/burner-net) | 98 | 10 | 238 | BurnerNet is a C++20 anti-forensic networking engine that securel... |
| 2 | [chernistry/bernstein](https://github.com/chernistry/bernstein) | 92 | 10 | 213 | Bernstein is a deterministic Python scheduler managing multiple C... |
| 3 | [samfoy/pi-total-recall](https://github.com/samfoy/pi-total-recall) | 85 | 9 | 213 | This repository showcases a comprehensive tool for analyzing and ... |
| 4 | [BasedHardware/omi](https://github.com/BasedHardware/omi) | 90 | 10 | 210 | AI-powered software that integrates screen monitoring, conversati... |
| 5 | [IronManus/ironmanus](https://github.com/IronManus/ironmanus) | 81 | 10 | 206 | Iron Manus Orchestrator |
| 6 | [iii-hq/iii](https://github.com/iii-hq/iii) | 84 | 9 | 199 | iii is the easiest way to compose, extend, and observe every serv... |
| 7 | [TheCraigHewitt/seomachine](https://github.com/TheCraigHewitt/seomachine) | 90 | 10 | 196 | SEO Machine is a specialized Claude Code workspace designed to au... |
| 8 | [alecthomas/proctor](https://github.com/alecthomas/proctor) | 90 | 10 | 196 | Proctor is a local development process manager that integrates wi... |
| | _...and 8 more_ | | | | |

---

## 4. PRIORITY VERIFICATION: High-Value Overlap

These repos exist in BOTH systems. Verify data freshness and sync:

| # | Repo | Sig | Inn | Description |
|---|------|-----|-----|-------------|
| 1 | [letta-ai/letta-code](https://github.com/letta-ai/letta-code) | 98 | 10 | Letta Code is a memory-first coding agent that replaces session-based ... |
| 2 | [puppeteer/puppeteer](https://github.com/puppeteer/puppeteer) | 98 | 10 | A high-level JavaScript API for controlling headless or full-featured ... |
| 3 | [xorrkaz/cml-mcp](https://github.com/xorrkaz/cml-mcp) | 98 | 10 | A model context protocol server enabling natural language interaction ... |
| 4 | [apify/actors-mcp-server](https://github.com/apify/actors-mcp-server) | 98 | 10 | Apify MCP Server enables AI agents to integrate with external data sou... |
| 5 | [taewoong1378/notion-readonly-mcp-server](https://github.com/taewoong1378/notion-readonly-mcp-server) | 98 | 10 | An optimized read-only MCP server for Notion API, designed to enhance ... |
| 6 | [yugabyte/yugabytedb-mcp-server](https://github.com/yugabyte/yugabytedb-mcp-server) | 98 | 10 | A MCP server enabling LLMs to interact with YugabyteDB for data access... |
| 7 | [elusznik/mcp-server-code-execution-mode](https://github.com/elusznik/mcp-server-code-execution-mode) | 98 | 10 | This project implements a discovery-first MCP bridge that executes Pyt... |
| 8 | [doggybee/mcp-server-ccxt](https://github.com/doggybee/mcp-server-ccxt) | 98 | 10 | High-performance integration of cryptocurrency exchanges using the Mod... |
| 9 | [TalaoDAO/connectors](https://github.com/TalaoDAO/connectors) | 98 | 10 | Wallet4Agent enables AI agents to securely interact with external serv... |
| 10 | [mKeRix/toolscript](https://github.com/mKeRix/toolscript) | 98 | 10 | Toolscript is a tool execution layer that minimizes context bloat by d... |
| 11 | [moazbuilds/CodeMachine-CLI](https://github.com/moazbuilds/CodeMachine-CLI) | 98 | 10 | CodeMachine-CLI is an open-source tool designed to orchestrate AI codi... |
| 12 | [vectorize-io/hindsight](https://github.com/vectorize-io/hindsight) | 98 | 10 | Hindsight is a biomimetic agent memory system that moves beyond simple... |
| 13 | [robotocore/robotocore](https://github.com/robotocore/robotocore) | 98 | 10 | A digital twin of AWS enabling local testing and simulation of AWS ser... |
| 14 | [badlogic/pi-mono](https://github.com/badlogic/pi-mono/blob/main/packages/coding-agent/docs/custom-provider.md) | 98 | 10 | A custom coding agent extension for integrating advanced AI models int... |
| 15 | [cocoindex-io/cocoindex-code](https://github.com/cocoindex-io/cocoindex-code) | 98 | 10 | A lightweight, AST-based semantic code search tool for developers, int... |
| 16 | [rafaljanicki/x-twitter-mcp-server](https://github.com/rafaljanicki/x-twitter-mcp-server) | 98 | 10 | A powerful AI-driven Twitter Management Platform enabling natural lang... |
| 17 | [pearl-com/pearl_mcp_server](https://github.com/pearl-com/pearl_mcp_server) | 98 | 10 | A standardized interface for integrating Pearl's AI and Expert service... |
| 18 | [cameronking4/programmatic-tool-calling-ai-sdk](https://github.com/cameronking4/programmatic-tool-calling-ai-sdk) | 98 | 10 | This SDK introduces Programmatic Tool Calling (PTC) to drastically red... |
| 19 | [jakops88-hub/Long-Term-Memory-API](https://github.com/jakops88-hub/Long-Term-Memory-API) | 98 | 10 | MemVault is a production-grade API platform that provides AI agents wi... |
| 20 | [OpenInterpreter/open-interpreter](https://github.com/OpenInterpreter/open-interpreter) | 98 | 10 | An open-source, local implementation of OpenAI's Code Interpreter that... |
| 21 | [bcharleson/instantly-mcp](https://github.com/bcharleson/instantly-mcp) | 98 | 10 | An MCP server for the Instantly.ai V2 API, enabling scalable and secur... |
| 22 | [allvoicelab/allvoicelab-mcp](https://github.com/allvoicelab/allvoicelab-mcp) | 98 | 10 | AllVoiceLab MCP server enabling AI-driven text-to-speech, video transl... |
| 23 | [kunihiros/kv-extractor-mcp-server](https://github.com/kunihiros/kv-extractor-mcp-server) | 98 | 10 | A powerful key-value extraction tool that automatically parses unstruc... |
| 24 | [verygoodplugins/automem](https://github.com/verygoodplugins/automem) | 98 | 10 | AutoMem is a production-grade long-term memory service for AI assistan... |
| 25 | [ChromeDevTools/chrome-devtools-mcp](https://github.com/ChromeDevTools/chrome-devtools-mcp) | 98 | 10 | An official Model Context Protocol (MCP) server that enables AI agents... |

---

## 5. BUILD PRIORITIES WITH ECOSYSTEM EVIDENCE

| Rank | Feature | Status | Evidence | Action | Ref |
|------|---------|--------|----------|--------|-----|
| 1 | Real Tiered Memory w/ Heat | Schema exists, no promotion | 49 candidates | Wire heat_score decay+promotion. L3 archive. LLM consolidation. | hindsight, Mimir, cognee |
| 2 | Close Self-Healing Loop | HealAndVerify EXISTS | 4 candidates | Add StopHook, IdleHealer. Wire healer to L2 vault. | context-foundry, agentic-qe |
| 3 | Progressive Skill Discovery | SkillDecisionSystem works | 21 candidates | Persist SkillEvolutionRecord. Add /evolve. Auto-retire low win-rate. | anthropics/skills, mcp-skills |
| 4 | Context Re-Injection | Harvester works | 31 candidates | CompactionHook. PreToolUse/PostToolUse. Token budget per tool. | zep, probe |
| 5 | Planner-Checker-Revise | PairOrchestrator exists | 58 candidates | Wire to real sessions. PlanMode: premium plan, budget execute. | agentmux, oh-my-opencode |
| 6 | Memory-Tool Feedback Loop | Both systems mature | 1.7% ecosystem deficit | MemoryInformedRanking. Store tool outcomes in L2. | roampal-core |
| 7 | Real Knowledge Graph | Interfaces only | 10 candidates | Entity extraction via LLM. Relationship edges. Blast radius queries. | cognee, Mimir, infranodus |
| 8 | Skill Win-Rate Tracking | EvolveSkill exists | 37x enrichment signal | Persist to SQLite. A/B test mutations. Auto-retire. | anthropics/skills |
| 9 | Graph+HITL Gates | Zero implementations | Signal 1,984 | BlastRadiusCalculator. AutoEscalationPolicy. HumanVetoService. | NOVEL - build from scratch |

---

## 6. RECOMMENDED SYNC PIPELINE

```
atlas.db (7,944 entries)
  |
  +-> high-signal MCP servers --> tormentnexus.db mcp_servers (68 -> ~120)
  |                            --> tormentnexus.db tools (651 -> ~900)
  |
  +-> architecture gap repos --> tormentnexus.db links_backlog (15,753 -> 15,900)
  |                            --> tormentnexus.db skill_candidate_queue
  |
  +-> innovation top-100 --> .tormentnexus/skills/ (0 -> curated set)

tormentnexus.db (15,753 backlog entries)
  |
  +-> missing from atlas --> atlas.db entries (7,944 -> ~9,400)
  |                       --> incoming_resources.txt for research worker
```
