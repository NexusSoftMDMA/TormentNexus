# TormentNexus: The Cognitive Control Plane & Universal TORMENTNEXUS

![Version](https://img.shields.io/badge/version-1.0.0--alpha.127-blue)
![Build](https://img.shields.io/badge/build-passing-brightgreen)
![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)
![TypeScript](https://img.shields.io/badge/TypeScript-5.x-3178C6?logo=typescript)

**TormentNexus** is the ultimate local-first control plane for multi-agent workflows, Model Context Protocol (MCP) tooling, provider routing, session continuity, and operator observability. 

We are building the substrate where a single local system seamlessly coordinates the most critical parts of AI-driven software development: tools, models, sessions, context, subagents, and full visibility across the entire stack. TormentNexus is not just an aggregator; it is a **decision system and universal bridge**.

---

## 🏗️ The Architecture (Modular Monolith)

TormentNexus has evolved into a high-performance **Go (Golang) modular monolith** with a **TypeScript/Next.js frontend**.
* **The Go Sidecar (`go/internal/`)**: Go handles the heavy lifting—orchestration, progressive MCP routing, L1/L2 memory management, and LLM waterfall routing.
* **The Control Panel (`apps/web/`)**: A rich Next.js and React dashboard serving as your visual observation deck.
* **The Storage (`sqlite-vec`)**: Dependency-free, hyper-fast local vector search for omniscient memory and tool ranking.

## ✨ Core Pillars

### 1. Progressive MCP Tool Routing & Parity
Models should never be overwhelmed with a 50,000-token tool dump. TormentNexus employs a multi-layered, progressive disclosure system:
* **Semantic Search:** Local vector embeddings match the active prompt against a global MCP directory.
* **The Router:** Only the top highly relevant tool schemas are injected into the active LLM context.
* **Universal Parity:** Byte-for-byte identical tool signatures for Claude Code, Codex, Gemini CLI, Cursor, and Windsurf.

### 2. Dual-Tier Memory Architecture (L1 / L2)
Context is finite; memory must be infinite.
* **L1 - Session Scratchpad:** Ephemeral, lightning-fast memory tied directly to the active session.
* **L2 - The Vault:** Permanent semantic storage in SQLite. Saves exact transcripts and LLM-compressed heuristics.
* **Context Harvesting:** Every session autonomously queries the L2 Vault to pull in relevant historical heuristics.

### 3. The Resilient LLM Waterfall
Uptime is non-negotiable. TormentNexus’s inference client natively catches 429s (Rate Limits) and 5xx (Server Errors), seamlessly cascading the exact payload down a prioritized chain without crashing:
1. **NVIDIA NIM** / Primary APIS
2. **OpenRouter** (Secondary aggregator fallback)
3. **Local LM Studio / Ollama** (Ultimate offline fallback)

### 4. Multi-Agent Swarm & P2P Mesh
TormentNexus coordinates specialized models inside shared chatrooms via the Agent-to-Agent (A2A) protocol. 
* **Role Rotation:** Models take turns acting as Planner, Implementer, Tester, and Critic.
* **Consensus & Debate:** Agents autonomously bid on tasks, share context via a neural transcript, and debate implementations until consensus is reached.

### 5. Truth Over Hype Dashboards
TormentNexus's dashboards reflect actual SQLite database rows and active Go goroutine states. No mocked UI scaffolds. Monitor telemetry, traffic inspection, working-set capacity, and LLM routing histories in real-time.

---

## 🚀 Quick Start

**Prerequisites:**
* Node.js 24+
* Go 1.26+
* pnpm v10

**Installation:**
```bash
# 1. Clone the repository
git clone https://github.com/robertpelloni/tormentnexus.git
cd tormentnexus

# 2. Install dependencies & rebuild SQLite bindings
pnpm install
pnpm rebuild better-sqlite3

# 3. Build the Go sidecar
cd go && go build -buildvcs=false ./cmd/tormentnexus && cd ..

# 4. Start the TormentNexus Control Plane
pnpm run dev
```
The Next.js dashboard will automatically open at `http://localhost:3000/dashboard` once the TS Control Plane and Go Sidecar are successfully locked and humming.

🗺️ Roadmap
See ROADMAP.md for our path to v1.0, including:
* Progressive Skill Disclosure (Context Hygiene)
* Go-native mcp sync migration
* Native UI replacement for Electron (Maestro)

Keep the party going. Never stop. The collective grows.
