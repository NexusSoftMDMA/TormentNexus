# Changelog

## [1.0.0-alpha.78] - 2026-05-30

### Added
- **Definitive Fallback Cascade**:
  - Expanded `DEFAULT_PROVIDER_CATALOG` in `ProviderRegistry.ts` with a massive, definitive list of free OpenRouter fallback models (DeepSeek V4, Llama 3 8B, Gemma 2 9B, Phi 3 Medium, Mistral 7B, GPT OSS 120B, Nemotron 3).
  - Implemented exhaustive recursive fallback checks in `CoreModelSelector.ts` that scan and attempt every single configured and healthy model in the catalog before falling back to LMStudio or Ollama.
  - Guarantees the system never runs out of models to try.
  - Verified with 100% clean tsc compilation check.

## [1.0.0-alpha.77] - 2026-05-30

### Added
- **LiteLLM Free Models Synchronization**:
  - Audited authoritative LiteLLM settings at `c:\Users\hyper\.hermes\litellm-config.yaml`.
  - Added new free OpenRouter model definitions (`openrouter/nousresearch/hermes-3-llama-3.1-405b:free`, `openrouter/liquid/lfm-2.5-1.2b-instruct:free`, `openrouter/qwen/qwen3-coder:free`) directly into `ProviderRegistry.ts`.
  - Seamlessly aligns the router fallback mechanism with active LiteLLM and Hermes configurations.
  - Verified with 100% clean tsc compilation check.

## [1.0.0-alpha.76] - 2026-05-30

### Added
- **LLM-Based Predictive Tool Ads (Tool Disclosure)**:
  - Upgraded `getPredictedToolAds` inside `MCPServer.ts` to leverage dynamic LLM-based tool predictions.
  - Automatically targets cheap/free OpenRouter models with cascading fallbacks to alternative models and a final local fallback to LMStudio.
  - Keeps a secondary fallback to the Go sidecar to maintain high execution resilience.
  - Seamlessly injects optimal MCP tool suggestions directly into the context window system prompts.
  - Verified with a 100% clean tsc compilation check.

## [1.0.0-alpha.75] - 2026-05-30

### Added
- **Dashboard Persistent Auto-Startup (Lazy Boot)**:
  - Added an automated, lazy background boot handler (`ensureDashboardRunning`) in `backgroundCoreBootstrap.ts`.
  - Seamlessly starts up the Next.js WebUI server (running either `.next/standalone/server.js` or `next dev` as fallback) on the first stdio/client request.
  - Spawned as a detached, unreferenced background server that remains running persistently.
  - Verified with 100% clean TypeScript compilation check.

## [1.0.0-alpha.74] - 2026-05-29

### Added
- **Deep MCP.directory Scraping (28,534 unique servers)**:
  - Extracted 1,797 server profiles directly from `mcp.directory` using high-performance concurrent crawler.
  - Successfully harvested descriptions, categorizations, website links, exact package identifiers (NPM, PyPI, Docker), and precise JSON launch configurations.
  - Added **502 new unique MCP servers** and updated **1,295 existing profiles** with high-fidelity metadata.
  - Finalized registry: **28,534 unique MCP servers** in `borg.db`.

## [1.0.0-alpha.73] - 2026-05-29

### Added
- **Massive MCP Registry Metadata Enrichment (28,032 unique servers)**:
  - Fetched all 307 cursor-paginated pages from the official MCP registry (`registry.modelcontextprotocol.io`), extracting deep runtime dependencies and package metadata.
  - Extracted **environment variables**, **auth models**, and **required secrets** for all official packages.
  - Deep-scraped Smithery (`registry.smithery.ai`) for detailed `configSchema` structures.
  - Enriched GitHub metadata (stars, topics, languages) for discovered servers.
  - Final catalog: **28,032 unique MCP servers** in `borg.db`.
  - Quality metrics: **9,726** high-confidence config recipes and **9,688** servers categorized with an explicit authentication model.

## [1.0.0-alpha.72] - 2026-05-29

### Added
- **Massive MCP Registry Expansion (18,881 unique servers)**:
  - Executed 8 parallel scraper waves targeting all known MCP directories, registries, and community lists.
  - **Sources scraped**: Official MCP Registry (`registry.modelcontextprotocol.io`), Smithery.ai (paginated, 294 servers), Glama.ai (99), PulseMCP, mcp.so, MCPHubX, NPM deep search (4,542 packages), GitHub Topics (11 topics × 5 pages), GitHub Search API (30+ queries), HackerNews MCP posts, PyPI, Docker Hub, Reddit MCP subreddits, `ever-works/awesome-mcp-servers`, `korchasa/awesome-mcp`, `tolkonepiu/best-of-mcp-servers`, `punkpeye/awesome-mcp-servers` (2,400+ repos), `wong2/awesome-mcp-servers`, `appcypher/awesome-mcp-servers`, `mcpso/servers`, `punkpeye/awesome-mcp-clients`, Cline marketplace, vibehackers.io, MCPNest, MCPPedia, and more.
  - **Bobbybookmarks deep file mining**: Extracted GitHub repos from 13 category `.md` files (1.7MB AGENT_ORCHESTRATION_WORKFLOW, 514KB AI_AGENTS_FRAMEWORKS, 496KB CONNECTIVITY_MCP_A2A, etc.), 8 `.txt` files (1.2MB incoming_resources.txt), and all 14 atlas.json layers (13,412 entries).
  - **atlas.db deep scan**: Re-extracted all GitHub URLs from all atlas.db entries (not just GitHub-hosted URLs).
  - **Deduplication**: All 58 source types are deduplicated against canonical IDs (github/owner/repo, npm/package, docker/image, etc.).
  - **Config recipes**: Auto-generated `npx`/`pip`/`docker` stdio recipes for all 18,877 servers with no existing recipe.
  - Final catalog: **18,881 unique MCP servers** with **58 source types** and **18,877 config recipes** in `borg.db`.

## [1.0.0-alpha.71] - 2026-05-29

### Added
- **Aggressive Text Ingestion & Integration (Borg Repos, Reprocess Queue, Incoming Resources)**:
  - Scraped additional online repository listings and raw resources inside `bobbybookmarks` directory.
  - Successfully extracted **2,078 unique MCP repositories** across multiple raw list files.
  - Ingested **1,392 new, unique MCP servers** and prunes/consolidated **686 existing servers**.
  - Generated fallback stdio configurations for all 1,392 newly registered servers.
  - Consolidated internal catalog database to hold **10,370 total unique MCP servers** under `published_mcp_servers` in `borg.db`.

## [1.0.0-alpha.70] - 2026-05-29

### Added
- **Global Multi-API Ingestion & Integration (Smithery, Glama, NPM, GitHub)**:
  - Triggered the core ingestion engine across all active remote directory APIs.
  - Ingested **767 servers** successfully across Glama, Smithery, NPM, and GitHub registries.
  - Advanced 188 newly discovered servers to `normalized` and created 32 active recipes.
  - Total internal catalog database now indexes **8,978 unique MCP servers** under `published_mcp_servers` inside `borg.db`.

## [1.0.0-alpha.69] - 2026-05-29

### Added
- **Online Registry Scraper (Awesome MCP Servers)**:
  - Fetched and scraped active repositories listed in three popular awesome-mcp-servers lists on GitHub.
  - Ingested **1,996 net-new MCP servers** and prunes/consolidated **629 existing servers** in `published_mcp_servers` inside `borg.db`.
  - Generated dynamic configuration recipes for all newly scraped servers.

## [1.0.0-alpha.68] - 2026-05-29

### Added
- **MCP Catalog Registry Ingestion & Integration**:
  - Integrated `bobbybookmarks` bookmarks database (`bookmarks.db` and `atlas.db`) into the central `borg.db` catalog.
  - Extracted **6,124 new, unique public MCP servers** and consolidated **72 existing servers** with rich descriptions, tags, and category taxonomies.
  - Automatically generated standard CLI stdio recipes for all newly ingested servers.
- **Aggressive Session and Memory Deduplication**:
  - Pruned **2,641 duplicate conversation/import sessions** in `imported_sessions`.
  - Pruned **15,104 duplicate memory blocks** in `imported_session_memories` to drastically optimize LLM context window payloads and database index speed.

## [1.0.0-alpha.65] - 2026-05-27

### Added
- **Dynamic Database Migration**:
  - Implemented automatic, safe, dynamic SQLite migrations for existing databases to append the missing `source_size` and `source_mtime` columns to the `imported_sessions` table on server initialization.
- **Robust MCP Tool Calling and Validation**:
  - Fixed client callTool parameter signature mismatch by explicitly passing `undefined` as the second argument when calling tools with custom options (resolving `safeParse` collision crash).
  - Validated full bidirectional MCP client-server communication over standard I/O (StdioClientTransport).
  - Tested native tools (`router_status`, `system_status`), filesystem fallbacks, and verified graceful connection recovery on downstream aggregated server crash (Healer Immune System reactor activation).

## [1.0.0-alpha.64] - 2026-05-25

### Added
- **TypeScript Compile Security & Alignment**:
  - Fully resolved all TypeScript compilation errors across `packages/core` by introducing the missing `ProviderAuthTruth` definitions and aligning `ProviderAuthState` and `ProviderQuotaSnapshot` with the new environment-telemetry models.
  - Eliminated unused `@ts-expect-error` directives, achieving a 100% clean type check.
- **Verification of Merged Feature Branches**:
  - Conducted deep graph audits and verified that all local and remote branches (`jules-...`, `nexus-...`) have been successfully merged into `main` with absolutely zero progress or feature regressions.

## [1.0.0-alpha.63] - 2026-05-25

### Added
- **Native Healer & L2 Vault Bridging**:
  - Implemented Go-native endpoints for `heal` and `vault/count` in the sidecar server.
  - Re-wired the TypeScript `healerRouter` to delegate all health and history queries to the Go kernel.
  - Unified the "Immune System" dashboard metrics with the Go `HealerService` state.
- **Ground Truth Mapping**:
  - Established field mapping (snake_case to PascalCase) for native records to ensure seamless UI integration without modifying the Go kernel's idiomatic output.

### Changed
- Updated all monorepo packages to version `1.0.0-alpha.63`.
- Improved accuracy of the Healer Vault counters by implementing total count queries in the SQLite backend.

## [1.0.0-alpha.62] - 2026-05-19

### Added
- **Deep Link Protocol Scheme (`hypercode://`) in Go**:
  - Built robust URI handling for `hypercode://attach?session=ID` and `hypercode://create?cliType=aider` commands.
  - Implemented single-instance CLI dispatcher. Clicking deep links routes actions through the active `hypercoded` daemon via HTTP REST.
- **SQLite L2 Vector Vault Visualizer**:
  - Implemented persistent database queries (`GetAllVaultRecords`) in Go fetching chronic vault memories ordered by importance and heat.
  - Wired the new tRPC `vaultRecords` query to the Next.js control plane to hook persistent SQLite vector records into the UI.
  - Re-designed the Healer dashboard in glassmorphic dark-mode, showing streaming active pathogens side-by-side with real persistent L2 Vault records.
- **Next.js Dashboard Routes**:
  - Added premium, highly interactive dashboard console cards for Blocks, Claude Chrome, Claude Cloud, Copilot, and OpenAI Codex.
- **LLM Instruction Unification**:
  - Resolved merge conflict markers and aligned role guidelines across `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `GPT.md`, and `copilot-instructions.md` under `docs/UNIVERSAL_LLM_INSTRUCTIONS.md`.

### Changed
- Standardized documentation identity to Hypercode Kernel & HyperCode.
- Replaced git merge conflict markers across multiple internal Kotlin and Markdown files with unified content logic.

## [1.0.0-alpha.61] - 2026-05-17

### Added
- **Autonomous Healer Loop (The Immune System)**:
  - New `HealerService` in the Go kernel with a multi-turn `diagnose -> fix -> verify -> retry` loop.
  - Integration with `CodeExecutor` for native, sandboxed verification (tsc, vitest, go test).
  - L2 Vault persistence: All healing events and extracted facts are saved as long-term memory for fleet-wide intelligence sharing.

### Changed
- Standardized documentation identity to Hypercode Kernel & HyperCode.
- Updated `VERSION.md`, `ROADMAP.md`, and `TODO.md` to reflect Phase 5 active sprint goals.
- Unified `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` as the single source of truth for all AI agents.
- Resolved merge conflict markers and aligned role guidelines across `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `GPT.md`, and `copilot-instructions.md`.

## [1.0.0-alpha.60] - 2026-05-16

### Added
- Fully integrated Go-native `MemoryManager` into the core TS control plane.
- Wires up `sqlite-vec` storage backend, replacing the deprecated `@hypercode/hypercode` implementation.
- Dual-tier cache invalidation for the L1/L2 memory boundaries.

### Changed
- Shifted authority of MCP configuration sync entirely to the Go sidecar.
- Removed legacy TS synchronization scripts for VSCode and Cursor.
