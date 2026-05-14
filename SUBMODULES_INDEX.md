# Monorepo Submodules Index

This document tracks the submodules, their roles, and integration status within the Borg monorepo.

## 🏗️ Project Structure
Borg is structured as a TypeScript monorepo using pnpm workspaces with a native Go sidecar for performance.

- **`packages/core`**: Main service orchestrator (Node.js/Go Sidecar).
- **`apps/web`**: Operator Dashboard (Next.js/React).
- **`packages/cli`**: Command-line Interface (`borg`).
- **`packages/adk`**: Agent Development Kit.
- **`packages/ui`**: Shared UI component library.
- **`go/`**: Native systems services and high-performance logic.
- **`submodules/`**: External reference implementations and forked dependencies.

## 📦 Submodule Inventory

| Name | Path | Purpose | Remote | Status |
|---|---|---|---|---|
| **Borg (Upstream)** | `submodules/borg` | Core logic reference | [robertpelloni/borg](https://github.com/robertpelloni/borg) | Integrated |
| **Maestro** | `apps/maestro` | Visual orchestrator | [robertpelloni/Maestro](https://github.com/robertpelloni/Maestro) | Integrated |
| **Claude Mem** | `packages/claude-mem` | Claude memory bridge | [robertpelloni/claude-mem](https://github.com/robertpelloni/claude-mem) | Integrated |
| **Aider** | `aider` | AI coding assistant | [paul-gauthier/aider](https://github.com/paul-gauthier/aider) | Reference |
| **Claude Code** | `claude-code` | Anthropic CLI reference | [yasasbanukaofficial/claude-code](https://github.com/yasasbanukaofficial/claude-code) | Parity Target |
| **OpenCode** | `opencode` | Open source AI harness | [anomalyco/opencode](https://github.com/anomalyco/opencode) | Parity Target |
| **LiteLLM** | `litellm` | Model abstraction | [BerriAI/litellm](https://github.com/BerriAI/litellm) | Reference |
| **Prism MCP** | `submodules/prism-mcp` | MCP server reference | [dcostenco/prism-mcp](https://github.com/dcostenco/prism-mcp) | Reference |
| **Goose** | `goose` | Block/GitHub AI agent | [block/goose](https://github.com/block/goose) | Reference |

## 📍 Hierarchy Summary
The project maintains a massive collection of 40+ submodules to ensure total feature parity with the state-of-the-art AI coding harnesses.

Key clusters:
- **CLIs**: `aider`, `claude-code`, `gemini-cli`, `opencode`, `kilocode`, `code-cli`.
- **Infrastructure**: `litellm`, `ollama`, `llamafile`, `azure-ai-cli`.
- **Reference**: `external/MetaMCP`, `submodules/prism-mcp`.
