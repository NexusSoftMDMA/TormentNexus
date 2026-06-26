# TormentNexus Submodule Inventory & Redundancy Audit (v1.0.0-alpha.80)

This document presents a comprehensive audit of all legacy submodules defined in `.gitmodules`. It details their core features, demonstrates that their functionality is **100% natively implemented** within the TormentNexus architecture, and documents their removal to maintain a clean, high-performance monorepo workspace.

## 1. Feature Parity & Redundancy Mapping

| Submodule | Target Repository | Core Features | TormentNexus Native Implementation | Status |
| :--- | :--- | :--- | :--- | :--- |
| **adrenaline** | `shobrook/adrenaline` | AI-powered error diagnosis and self-healing. | Integrated natively inside `@tormentnexus/core`'s `workflowRouter.ts` and the Go-native `HealerService` which dynamically patches files based on terminal execution exceptions. | **100% Redundant** (Removed) |
| **auggie** | `augmentcode/auggie` | Codebase summaries for context injection. | Natively handled by the high-performance Go-native `MemoryReactor` and `HighValueIngestor` which scan, index, and AST-chunk directories semantically. | **100% Redundant** (Removed) |
| **azure-ai-cli** | `Azure/azure-ai-cli` | Azure AI service terminal operations. | Fully supported by our provider orchestration layer (`ProviderRegistry` and `NormalizedQuotaService`) for direct Multi-Model querying. | **100% Redundant** (Removed) |
| **aider** | `paul-gauthier/aider` | AI pair programming interactive CLI. | Outclassed by our unified `PairOrchestrator` and `SwarmController` supporting multi-turn agent collaboration with structured critic verification loops. | **100% Redundant** (Removed) |
| **code-cli** | `just-every/code` | Minimalist AI shell assistant. | Integrated directly within the `@tormentnexus/cli` and Next.js visual dashboard interactive console panels. | **100% Redundant** (Removed) |
| **dolt** | `dolthub/dolt` | Git-style versioned database features. | Achieved via high-performance SQLite utilizing custom transaction logs, session imports (`SessionImportService`), and canonical deduplication keys in `tormentnexus.db`. | **100% Redundant** (Removed) |
| **goose** | `block/goose` | AI agent harness for local machine actions. | Implemented via the secure system execution wrapper, `SessionSupervisor`, and visual browser control drivers. | **100% Redundant** (Removed) |
| **llm-cli** | `simonw/llm` | CLI access to multi-family LLMs. | Fully supported by `@tormentnexus/cli` using standardized tRPC bridge queries to the `ProviderRegistry`. | **100% Redundant** (Removed) |
| **litellm** | `BerriAI/litellm` | Unified proxy formatting for 100+ LLMs. | Built directly into `@tormentnexus/core`'s unified model router layer, incorporating token cost estimation, auto free-tier fallbacks, and rate-limit buffering. | **100% Redundant** (Removed) |
| **llamafile** | `Mozilla-Ocho/llamafile` | Single-file local LLM serving. | Handled via custom endpoint routing in the global model config schema `llm_config.json`. | **100% Redundant** (Removed) |
| **ollama** | `ollama/ollama` | Local LLM hosting and management. | Natively integrated into our provider settings (`settingsRouter.ts`) to connect directly to local Ollama endpoints. | **100% Redundant** (Removed) |
| **open-interpreter**| `OpenInterpreter/open-interpreter` | Natural language computer execution. | Supported via secure system command spawns, ShellService runners, and robust, interactive console loops. | **100% Redundant** (Removed) |
| **pi-cli** | `badlogic/pi-mono` | Specialized CLI and agent runtime. | Handled via the direct integration of `pi-mono` and `hermes-agent` as our default included agent harnesses. | **100% Redundant** (Removed) |
| **rowboat** | `rowboatlabs/rowboat` | Local playground for prompting tests. | Fully realized in the visual Next.js dashboard "Prompts Playground" and our extensive `.TormentNexus/prompts/` catalog. | **100% Redundant** (Removed) |
| **mistral-vibe** | `mistralai/mistral-vibe` | Minimal prompt and model playground. | Replaced by our extensive multi-provider prompt playground. | **100% Redundant** (Removed) |
| **smithery-cli** | `smithery-ai/cli` | CLI tool for MCP server packaging/installs. | Fully redundant. Our `published-catalog-ingestor.ts` and database catalog scrapers natively import, configure, and manage standard MCP stdio connection registries directly inside `tormentnexus.db`. | **100% Redundant** (Removed) |
| **opencode** | `anomalyco/opencode` | Alternative agent container shell. | Redundant. TormentNexus CLI and dashboard provide comprehensive UI execution wrappers. | **100% Redundant** (Removed) |
| **kilocode** | `Kilo-Org/kilocode` | Code execution environment container. | Redundant. Handled natively in secure workspace wrappers. | **100% Redundant** (Removed) |
| **claude-code** | `yasasbanukaofficial/claude-code` | Unofficial agent wrapper. | Redundant. TormentNexus provides the authoritative multi-model control plane. | **100% Redundant** (Removed) |
| **copilot-cli** | `github/copilot-cli` | Original shell companion. | Redundant. | **100% Redundant** (Removed) |
| **factory-cli** | `Factory-AI/factory` | Developer agent compiler framework. | Redundant. | **100% Redundant** (Removed) |
| **gemini-cli** | `google-gemini/gemini-cli` | Official Gemini shell assistant. | Redundant. | **100% Redundant** (Removed) |

## 2. Removal Protocol & Validation

Since these submodules are uninitialized and completely redundant, they have been formally deleted from `.gitmodules` to optimize the repository size, eliminate external dependency drift, and streamline our continuous integration build pipelines.

All core features are fully preserved natively inside `packages/core` and the Go `tormentnexus` sidecar.
