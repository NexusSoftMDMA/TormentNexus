# A2A and Agent SDK Evaluation

## Overview
As tormentnexus moves towards a decentralized, native, multi-agent architecture, we must evaluate the landscape of Agent SDKs and Agent-to-Agent (A2A) protocols. The goal is to create a lightweight, high-performance runtime (the Go port) while supporting interoperability across various agentic frameworks.

## 1. Agent SDK Evaluation

### 1.1 LlamaIndex / LangChain / LangGraph
- **Pros:** Massive ecosystem, huge number of integrations, mature concepts for memory and routing.
- **Cons:** Extremely heavy, python/node-centric, highly opinionated abstractions that often conflict with our `Council` / `SessionSupervisor` model.
- **Verdict:** Do not use as the core engine. Support them as downstream harnesses (running in sandboxes) but keep tormentnexus core agnostic.

### 1.2 Microsoft AutoGen
- **Pros:** Strong support for multi-agent conversations, group chats, and state machines.
- **Cons:** Very Python-heavy, steep learning curve, abstraction overhead.
- **Verdict:** Good for specific complex reasoning tasks. We can wrap AutoGen scripts as "Skills" or "Toolchains" rather than adopting it as the core.

### 1.3 CrewAI
- **Pros:** Excellent role-based agent design, very intuitive for end-users.
- **Cons:** Built on top of LangChain (heavy), Python only.
- **Verdict:** Integrate via external process invocation.

### 1.4 Google Agent SDK (A2A)
- **Pros:** Official protocol from Google, supports discovery via Agent Cards, asynchronous task polling, multi-language support (Go, Python, Java).
- **Cons:** Still evolving (v0.3.0), specific focus on Agent-to-Agent communication rather than a full orchestrator.
- **Verdict:** **Mandatory Support.** This is the primary standard for external agent interoperability. We must implement `AgentCard` exposure and the A2A client in our Go port.

### 1.5 Native tormentnexus Agents (Our custom approach)
- **Pros:** Zero-dependency, lightweight, strictly adheres to our Memory, MCP, and CLI Harness paradigms.
- **Cons:** Requires us to build state management and A2A routing from scratch.
- **Verdict:** Continue building our own `SessionSupervisor` and `Council` layers in Go/TS, utilizing the `auto_call_tool` pattern and `PreemptiveToolAdvertiser` for dynamic routing.

## 2. Agent-to-Agent (A2A) Protocols

To achieve true swarm capability, our agents must talk to each other and to external agents.

### 2.1 Google A2A Protocol
- **Concept:** Standardized communication via `A2AClient` and `AgentCard`.
- **Verdict:** **Adopt as the primary external A2A standard.** Implement the full specification in the Go control plane.

### 2.2 Model Context Protocol (MCP) as A2A
- **Concept:** Treat other agents as MCP servers. Agent A calls a tool `ask_agent_b(query)` which routes over Stdio/SSE to Agent B.
- **Pros:** We already have a world-class MCP Aggregator. It unifies tools and agents into a single interface.
- **Verdict:** **Highly Recommended.** Use this for local/low-latency agent-to-tool and agent-to-subagent communication.

### 2.3 Agent Client Protocol (ACP)
- **Concept:** Emerging standard for IDE-to-Agent communication.
- **Verdict:** We should implement an ACP adapter to allow external IDEs (like Cursor/Windsurf) to talk to tormentnexus's background agent daemon.

### 2.4 Custom WebSockets / gRPC
- **Concept:** High-throughput streaming between Go nodes.
- **Verdict:** Will be implemented in the Go port (`tormentnexus-server`) for internal control-plane communication, but MCP/A2A remains the standard for the actual LLM payload routing.

## Conclusion & Next Steps
1. **Adopt Google A2A:** Implement the Google Agent SDK protocols for external interoperability.
2. **Double down on MCP:** Use MCP not just for tools, but for wrapping external Agent SDKs. If a user wants a CrewAI swarm, they run it as an MCP server, and tormentnexus talks to it via tool calls.
3. **Go Port:** Implement the high-performance A2A router in Go, capable of multiplexing thousands of SSE/Stdio connections.
4. **Progressive Disclosure:** Expose these sub-agents to the master LLM via the same progressive semantic search we use for standard tools.
