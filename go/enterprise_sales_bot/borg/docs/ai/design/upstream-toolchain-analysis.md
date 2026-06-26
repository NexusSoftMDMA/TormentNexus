# Upstream Toolchain Analysis

This document is the current clean-room reference map for the imported toolchains. It is intended to guide feature assimilation and architectural borrowing without creating hidden dependency on upstream repos.

## TormentNexus
- **Role**: local-first control plane
- **Best traits**: MCP aggregation, provider routing, session import, memory substrate, runtime observability
- **Assimilation guidance**: treat as the substrate for routing, memory, MCP, and operator truth; do not duplicate these concerns inside the new harness unless reliability requires it.

## Pi
- **Role**: minimal coding harness foundation
- **Best traits**: small core, excellent extension seams, precise event vocabulary, multiple modes (interactive/json/rpc/sdk)
- **Assimilation guidance**: this is the best conceptual base for the Go harness. Port the contracts and philosophy first.

## Aider
- **Role**: code-editing specialist
- **Best traits**: repo map, multiple edit strategies, git-native workflow
- **Assimilation guidance**: port the context and edit engines, not the Python runtime shape.
- **Current native status**: first ranked repo-map baseline now exists in `foundation/repomap`, with lightweight cross-file graph influence, but it is still much simpler than Aider’s full graph-aware pagerank pipeline.

## Adrenaline
- **Role**: early repository-grounded coding assistant
- **Best traits**: codebase indexing and repo-grounded Q&A
- **Assimilation guidance**: use as a historical/contextual reference for repo-grounded interactions.

## Auggie
- **Role**: augmented coding workflow
- **Best traits**: editor-oriented augmentation and repo-aware assistance
- **Assimilation guidance**: mine ergonomic ideas, not core runtime architecture.

## Azure AI CLI
- **Role**: cloud/enterprise provider integration
- **Best traits**: Azure auth and deployment flows
- **Assimilation guidance**: expose through TormentNexus/provider adapters rather than as primary harness behavior.

## Bito CLI
- **Role**: engineering context layer
- **Best traits**: review assistance and workflow context
- **Assimilation guidance**: useful for review workflows and memory framing.

## ByteRover CLI
- **Role**: portable memory system
- **Best traits**: persistent memory, curation, retrieval
- **Assimilation guidance**: pair its memory ideas with TormentNexus’s local-first memory substrate.

## Claude Code mirror / templates
- **Role**: behavior reference and workflow templates
- **Best traits**: popular prompt/workflow patterns
- **Assimilation guidance**: only public behavior patterns should inform clean-room design.

## Code CLI
- **Role**: Codex-style harness compatibility
- **Best traits**: useful for tool-surface compatibility expectations
- **Assimilation guidance**: cross-check public tool contracts and UX expectations.

## Copilot CLI
- **Role**: official terminal coding surface
- **Best traits**: OAuth/subscription integration and shell ergonomics
- **Assimilation guidance**: leverage for auth/session import expectations, not for core architecture.

## Crush
- **Role**: native TUI excellence
- **Best traits**: Bubble Tea/Charm UX polish
- **Assimilation guidance**: primary TUI quality benchmark for the Go-native interface.

## Dolt
- **Role**: versioned SQL database
- **Best traits**: branching, diffing, and merging data itself
- **Assimilation guidance**: potential long-term basis for auditable memory or session stores.

## Factory Droid
- **Role**: persistent autonomous engineer/orchestrator
- **Best traits**: long-running sessions, verification, cross-surface persistence
- **Assimilation guidance**: strongest reference for detached/background workflows and verify loops.

## Gemini CLI
- **Role**: Google-integrated coding CLI
- **Best traits**: provider integration and subscription flow
- **Assimilation guidance**: bridge auth and model surfaces via TormentNexus/provider adapters.

## Goose
- **Role**: high-quality agent core
- **Best traits**: clean architecture, protocol orientation, systems rigor
- **Assimilation guidance**: architectural benchmark for how to structure a serious agent runtime.

## Grok CLI
- **Role**: autonomous agent with delegation and remote control
- **Best traits**: sub-agents, verify, batch mode, Telegram remote control
- **Assimilation guidance**: import the detached-work and delegation ideas behind a cleaner Go service boundary.

## Jules extension
- **Role**: Gemini-oriented workflow extension
- **Best traits**: workflow integration patterns
- **Assimilation guidance**: convert to skill/extension packs rather than core runtime logic.

## Kilo Code
- **Role**: autonomous engineering platform
- **Best traits**: throughput, orchestration, popularity as coding agent
- **Assimilation guidance**: benchmark for autonomous task decomposition and model presets.

## Kimi CLI
- **Role**: Kimi provider-specific coding harness
- **Best traits**: long-context model targeting
- **Assimilation guidance**: represent as a provider profile rather than a separate architecture.

## LLM CLI
- **Role**: shell-native LLM utility
- **Best traits**: Unix composability, plugins, restraint
- **Assimilation guidance**: important reference for scripting mode and shell ergonomics.

## LiteLLM
- **Role**: provider normalization/router
- **Best traits**: broad compatibility and normalization
- **Assimilation guidance**: overlap heavily with TormentNexus; assimilate only gaps, avoid duplicate routing stacks.

## Llamafile
- **Role**: local model packaging/runtime distribution
- **Best traits**: single executable model delivery
- **Assimilation guidance**: integrate as runtime target, not harness architecture.

## Mistral Vibe
- **Role**: minimal coding agent
- **Best traits**: small surface area and quick startup
- **Assimilation guidance**: use to keep the default harness from becoming bloated.

## Ollama
- **Role**: local model server
- **Best traits**: local inference, simple deployment, Go-native ecosystem fit
- **Assimilation guidance**: first-class local provider target for the Go harness.

## Open Interpreter
- **Role**: computer-use / local automation agent
- **Best traits**: desktop automation and host integration
- **Assimilation guidance**: import computer-use features only behind clear trust/sandbox boundaries.

## OpenCode
- **Role**: rich open-source coding platform
- **Best traits**: ambitious TUI, client/server shape, provider agnosticism, strong UX
- **Assimilation guidance**: direct competitive benchmark. Must beat it on speed, context accuracy, and trust.

## Qwen Code CLI
- **Role**: Qwen-specific harness
- **Best traits**: model-specific tuning
- **Assimilation guidance**: encode as provider presets and prompt profiles.

## Rowboat
- **Role**: repository analysis tool
- **Best traits**: browse/explain repository workflows
- **Assimilation guidance**: useful for read-only analysis mode and explainability features.

## Smithery CLI
- **Role**: MCP discovery/install ecosystem tool
- **Best traits**: discovery and install UX for MCP servers
- **Assimilation guidance**: complement TormentNexus runtime truth with better operator discovery UX.

## Synthesis
The overall direction is clear:
- **Pi** supplies the best minimal harness foundation.
- **Aider** supplies context and edit quality.
- **OpenCode** defines the modern benchmark to beat.
- **Goose** and **Crush** demonstrate what good native architecture and UX look like.
- **Factory** and **Grok CLI** define the detached/autonomous frontier.
- **TormentNexus** should stay the control-plane substrate.
