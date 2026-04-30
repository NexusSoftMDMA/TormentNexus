# Missing Features Analysis (April 2026)

Based on the recent massive Go-native integration wave, we've successfully brought the core control plane to feature parity with the TypeScript version, executing:
- MCP Decision System (ranked discovery, auto-load, LRU eviction)
- Universal Tool Parity Aliases
- Memory Manager (sectioned states, caching, context harvesting)
- Code Executor (multi-language sandbox with tool call intercept)
- UI Dashboards for all the above

However, looking at `TODO.md` and `ROADMAP.md` along with user instructions, the following critical features remain missing or incomplete:

## 1. Multi-Model Chatroom & Advanced A2A Protocols
- While the foundational `A2A Negotiation` and `Go A2A signal auditing` were completed, the actual **Multi-model chatroom — shared context between rotating models (PairOrchestrator)** still feels experimental.
- **Debate & Consensus Protocols:** Needs UI visualization and tighter integration. Models should rotate roles (Planner, Implementer, Tester) seamlessly in the Go backend.

## 2. Browser Extension
- The `packages/browser` Chrome/Firefox extension currently has manual sync but lacks:
  - Deep injection into native web chats (e.g., intercepting Claude.ai or ChatGPT to expose local MCP tools directly).
  - Browser History ingestion into local `MemoryManager`.
  - Background memory syncing.

## 3. Supervisor Tool Prediction
- The `mcp.ToolPredictor` was added to `server.go`, but **Supervisor tool prediction (preemptively injecting tool ads based on conversation context)** is not fully realized in the active chat prompt chain.

## 4. Free-Tier Fallback Expansion
- Fallback works for Gemini 2.5 Flash, but we need to add explicitly configured chains for:
  - OpenRouter Free
  - Google AI Studio (native)

## 5. UI Parity
- **Mobile responsiveness** is listed as incomplete.
- **Native UI** to replace Electron `Maestro` is listed under P2/Vision. This aligns with the user's prompt: *"let's make our own native UI instead of using electron, we want to be super lightweight and super fast."*

## 6. Progressive Skill Disclosure
- We built "Progressive Tool Disclosure", but the exact same LRU/ranking engine needs to be applied to **Skills**, so models don't get overwhelmed with thousands of skill instructions at once.
