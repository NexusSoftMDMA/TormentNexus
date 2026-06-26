# Go Foundation Assimilation Requirements

## Problem
The current repository has an aspirational Go CLI and orchestration shell, but the implementation is still mostly stub-level and does not yet provide truthful parity with Pi, OpenCode, Aider, Goose, Factory Droid, Grok CLI, TormentNexus, or the other imported submodules.

The user goal is not a narrow feature addition. The goal is to create a new **Go-native foundation** that can eventually become a best-in-class coding harness and operator-facing agent runtime with:

1. Pi-quality extension seams and minimal-harness philosophy.
2. OpenCode-level UX and provider-agnostic reach.
3. Aider-level repository context and edit strategy quality.
4. Goose/Crush-grade native performance and terminal feel.
5. Factory/Grok-grade long-running orchestration, verification, and delegation.
6. TormentNexus/TormentNexus-native integration for MCP routing, memory, provider routing, and context continuity.
7. Exact model-facing tool names and parameter contracts wherever compatibility matters.

## Non-Negotiable Requirements

### 1. Exact tool contracts where models depend on them
The new foundation must preserve exact model-facing tool names, parameter shapes, and observable behavior for the compatibility-critical tools.

Initial exact-name focus:
- `read`
- `write`
- `edit`
- `bash`

Longer-term exact-contract focus:
- Pi-compatible default tool surface
- Codex/OpenCode-compatible high-value tool surfaces where stable public contracts exist
- TormentNexus/TormentNexus-backed MCP tool surfaces exposed through stable adapters

### 2. Clean-room implementation
Assimilation should preserve behavior, not copy licensed internals from closed or questionable upstream sources.

Allowed:
- public documentation
- public CLI behavior
- publicly visible prompts, flags, contracts, and tool surfaces
- clean-room reimplementation

Not allowed:
- unlawful source reuse
- claiming parity that is not implemented
- silent compatibility drift

### 3. TormentNexus as substrate, not competitor
The new harness should not reimplement TormentNexus’s core control-plane responsibilities unless there is a compelling reliability reason.

TormentNexus should remain the preferred substrate for:
- MCP aggregation
- provider routing/fallback
- memory and session import/export
- runtime observability
- tool inventory and inspection

### 4. Native Go runtime
The foundation should be implemented in Go, with optional plugin or scripting layers on top rather than a JS runtime at the center.

### 5. Truthful maturity labeling
Every feature and route must have an explicit maturity state:
- planned
- bridged
- speced
- native
- verified

## Functional Scope

### Foundation parity with Pi
Required eventually:
- interactive mode
- print/json mode
- rpc/daemon mode
- session persistence
- branching/forking
- compaction hooks
- prompt templates
- skills
- themes
- extensions
- configurable keybindings
- model/provider abstraction
- exact event vocabulary for the agent loop

### Assimilation from other tools
Required eventually:
- repo map / code graph context
- multiple edit strategies
- autonomous planning and delegation
- verification mode
- shell-first workflows
- local model runtime support
- computer-use / automation integration
- MCP discovery/install UX
- provider-specific model presets
- remote-control and long-running session support

## Constraints
- Do not kill running processes.
- Do not overstate parity.
- Prefer additive work that establishes a stable foundation.
- Documentation must be updated as architecture evolves.
- The current repository baseline has failing tests/build issues unrelated to this foundation work; those should be documented rather than hidden.

## Initial Deliverable for This Phase
This phase should establish:
1. a truthful requirements and design baseline,
2. a Go-native compatibility/foundation type system,
3. an assimilation inventory covering the imported upstream tools,
4. a CLI inspection surface for the new foundation spec,
5. comprehensive documentation for follow-on implementation.
