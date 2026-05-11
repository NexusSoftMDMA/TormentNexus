[PROJECT_MEMORY]

# Super AI Plugin (Borg) - Comprehensive Architectural Memory

Based on a detailed analysis of the repository's history, the massive 16,735-item Borg-researched bookmark database, and the recovery processes undertaken during this session, the project's architecture, patterns, and decisions are centered around **14 High-Conviction Features**. This document serves as the global memory for these architectural decisions and implementation strategies.

## 1. Dynamic Tool Discovery & Registry (ToolRAG)
**The Problem:** The MCP ecosystem has over 25,000 tools. Static loading exhausts the LLM context window (imposing a 32% token overhead penalty).
**The Converging Solution:** "RAG but for tools." Embed tool names only and fetch full JSON schemas strictly on-demand.
*   **Borg Implementation:** Build `borg-tool-registry` to index all discovered MCP servers, embed schemas using SQLite Vector Search, and inject only the 3-5 most relevant tools per query.

## 2. Tiered Memory Architecture (Core/Archival/Recall)
**The Problem:** Agent summarization decay, memory hoarding (noise), and context pollution.
**The Converging Solution:** A 3-tier hierarchy.
*   *Core:* Active working context (always in prompt).
*   *Archival:* Long-term compressed storage (Borg's 17K database).
*   *Recall:* Episodic memory with biological decay (Ebbinghaus curve) and heat-based promotion.
*   **Borg Implementation:** Utilize the existing database as the Archival tier. Build a Core tier that auto-curates 5-10 items and a Recall tier that tracks session access patterns.

## 3. Progressive Context Optimization
**The Problem:** Token bloat is the #1 pain point, degrading reasoning and increasing costs.
**The Converging Solution:** Schema elimination (Code Mode), Intelligent Chunking (cAST), Context Compaction (Union-Find), and Progressive Disclosure.
*   **Borg Implementation:** Store "compressed fingerprints" (category + 3 key tags + innovation score) instead of raw HTML text. Hydrate full context only when explicitly requested by the agent. Implemented "Fit Markdown" filtering to strip out boilerplate HTML tags (`<nav>`, `<header>`, `<footer>`, `<aside>`).

## 4. MCP Ecosystem Intelligence (Registry + Package Manager)
**The Problem:** Thousands of unverified MCP servers exist with massive injection risks and inconsistent auth.
**The Converging Solution:** Curated global registries (like `mcpm` or `Smithery`) providing 1-click deployments, profile management, and 5-dimension security scoring.
*   **Borg Implementation:** Build `borg-mcp-catalog` referencing ingested `servers.json` catalogs. Implement one-command install profiles (`borg install memory`) with automated poisoning detection scans.

## 5. Multi-Agent Orchestration with Verification
**The Problem:** "Verification Debt" - single-agent generation scales infinitely, but verifying accuracy remains the bottleneck, leading to a 28% failure rate on scraped extractions.
**The Converging Solution:** Planner-Executor-Verifier loops. Cross-model validation is critical.
*   **Borg Implementation:** When extracting URL data, spawn a lightweight Review Agent (using a different model) to cross-validate the category/tags against the taxonomy.

## 6. Sandboxed Code Execution Layer
**The Problem:** LLMs guess at repository capabilities based on READMEs instead of actual code analysis.
**The Converging Solution:** Execution isolation ranging from Docker to MicroVMs (BoxLite/Firecracker) and the "Code-as-Action" paradigm.
*   **Borg Implementation:** Spawn a sandboxed container to clone a repo, run `tree-sitter` for AST parsing, and execute deterministic extraction scripts to replace LLM guesswork.

## 7. Self-Improving Research Pipeline
**The Problem:** Static extraction scripts degrade over time with no feedback loop to learn from errors.
**The Converging Solution:** Telemetry tracking, self-editing memory blocks, and background "Dreaming" consolidation.
*   **Borg Implementation:** Tag extraction quality in the database. Auto-re-process low-quality items using stronger models. Implement a reasoning "flight recorder" to trace why specific routing/extraction decisions failed. (Phase 1 Garbage Filter & Flight Recorder already implemented).

## 8. Agent Communication Protocols (A2A / AG-UI / ACP)
**The Problem:** The agent ecosystem is fragmented into siloed communication models.
**The Converging Solution:** A unified stack.
*   *A2A:* Discovery and task delegation.
*   *AG-UI:* Dynamic generative UI rendering.
*   *ACP:* Persistent daemon sessions.
*   **Borg Implementation:** Expose Borg as an A2A agent to allow other tools to discover and delegate URL processing. Implement AG-UI for the Kinetic HUD, and ACP daemon mode to persist sessions beyond context resets.

## 9. Graph/RAG/Vector Intelligence
**The Problem:** Flat SQL queries prevent semantic and relationship-aware discovery.
**The Converging Solution:** Generation 3 (GraphRAG) and Generation 5 (Compression-Optimized Vectors).
*   **Borg Implementation:** Build a tag co-occurrence graph (weighted edge graph) for finding related tasks. Use DuckDB VSS and Arctic Embed (128-byte scalar quantization) for hyper-fast semantic search.

## 10. Security, Governance & Agent Guardrails
**The Problem:** Unrestricted code execution and injection risks in MCP servers.
**The Converging Solution:** Network isolation, command guardrails, and audit provenance.
*   **Borg Implementation:** Implement SafeExec-style interception to block destructive actions. Mask PII during extraction. Run tool poisoning detection during MCP server ingestion. Generate "AI Receipts" for every decision.

## 11. Observability, Telemetry & Agent Debugging
**The Problem:** Fire-and-forget scripts make debugging 28% failure rates impossible.
**The Converging Solution:** Reasoning Traces, OTEL Spans, and Anomaly Detection.
*   **Borg Implementation:** Implement Fiddler-style anomaly detection (Jensen-Shannon Divergence) to flag statistically abnormal reasoning outputs. Add Reticle-style session recording to capture full JSON-RPC payloads for replay debugging.

## 12. Model Intelligence & Cost Optimization
**The Problem:** Single-model routing is inefficient; 10x cost variations exist.
**The Converging Solution:** Complexity-based tiered routing.
*   **Borg Implementation:** The `WaterfallRouter` implements tiered routing: Simple URLs (Reddit, YCombinator) route to a local 1.2B SLM, and complex tasks (GitHub repos) route to frontier models. Track model performance ELO and auto-demote degrading models.

## 13. Browser Automation & Web Interaction
**The Problem:** Naive HTTP fetching fails on SPAs, captchas, and auth walls.
**The Converging Solution:** Accessibility Tree extraction (Generation 2) and Zero-Copy Vision (Generation 3).
*   **Borg Implementation:** Use Accessibility Trees for SPA/Docs to achieve 90% token reduction vs raw DOM. Fall back to Zero-Copy POSIX shared memory vision for complex diagrams.

## 14. CI/CD & Git-Integrated Workflows
**The Problem:** Moving from code generation to code *verification* as the primary bottleneck.
**The Converging Solution:** Agent-native git workflows, worktree isolation, and review-as-quality-gates.
*   **Borg Implementation:** Version the database schema using git-commit-like operations. Implement closed-loop self-healing (if extraction quality is poor, automatically checkout a new worktree and re-process with a different model).

---

## Architectural Meta-Patterns & Rules of Engagement

During this session, we established a strict recovery protocol. The codebase previously suffered from "Slop at Scale" where blind text-replacement scripts structurally corrupted the Go workspace (deleting closing braces, duplicating imports, and breaking interfaces). The definitive pattern for modifying this project is:

1.  **Do not use blind `sed` or line-number deletions across multiple files.**
2.  **Use AST-aware tooling or surgically precise Python scripts** that check surrounding context before mutating code.
3.  **Ensure a clean `go build ./...` state** before committing or merging feature branches.
4.  **Single Source of Truth:** `ROADMAP.md`, `CHANGELOG.md`, `VISION.md`, and `docs/PROJECT_MEMORY.md` heavily dictate immediate priorities.

*This memory is continuously synced with `ROADMAP.md` and `VISION.md` to ensure the project architecture converges on these 14 principles.*
