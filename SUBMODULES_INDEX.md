# Submodules Index & Project Structure

_Last updated: 2026-04-19_

This document tracks all external submodules merged into the HyperCode workspace to ensure feature parity and architectural integration.

## Current Submodules

| Submodule | Location | Upstream | Purpose / Status |
|-----------|----------|----------|------------------|
| **jules-autopilot** | `apps/cloud-orchestrator` | `robertpelloni/jules-autopilot` | Cloud orchestration and TS fallback routing. (Synced) |
| **Maestro** | `apps/maestro` | `robertpelloni/Maestro` | Electron-based GUI client. (Synced, slated for native replacement) |
| **OmniRoute** | `external/OmniRoute` / `archive/OmniRoute` | `diegosouzapw/OmniRoute` | Reference for unified LLM provider routing. |
| **litellm** | `submodules/litellm` | `BerriAI/litellm` | Proxy server for 100+ LLMs. Reference for Go-native provider abstraction. |
| **mcpproxy** | `submodules/mcpproxy` | `Dumbris/mcpproxy` | Reference for Go-based MCP routing and middleware. |
| **claude-mem** | `packages/claude-mem` | `robertpelloni/claude-mem` | Memory ingestion system. Fully ported to Go `MemoryManager`. |
| **hyperharness** | `submodules/hyperharness` | `robertpelloni/hyperharness` | Core CLI harness integration. |
| **multica** | `submodules/multica` | `multica-ai/multica` | Multi-agent conversation structures. |
| **unifyroute** | `submodules/unifyroute` | `unifyroute/unifyroute` | Reference for fallback chains. |
| **coding_agent_usage_tracker** | `submodules/coding_agent_usage_tracker` | `Dicklesworthstone/coding_agent_usage_tracker` | Billing, token tracing, and telemetry reference. |
| **CLIProxyAPIPlus** | `submodules/CLIProxyAPIPlus` | `robertpelloni/CLIProxyAPIPlus` | Shell proxy utilities. |
| **LinJun** | `submodules/LinJun` | `wangdabaoqq/LinJun` | Agent workflows reference. |
| **HyperHarness** | `submodules/HyperHarness` | `robertpelloni/HyperHarness` | Primary local CLI orchestration system. |
| **pi-mono** | `submodules/pi-mono` | `robertpelloni/pi-mono` | Reference implementations. |

## Project Structure

```text
/
├── apps/               # Standalone applications (cloud-orchestrator, Maestro)
├── go/                 # The Go-native control plane (New primary backend)
│   ├── cmd/            # Go binaries
│   └── internal/       # Core Go implementations (Memory, MCP, HTTP API, CodeExec)
├── packages/           # TypeScript libraries
│   ├── ai/             # TS Provider gateways
│   ├── browser/        # Chrome/Firefox extension
│   ├── cli/            # Node.js CLI entrypoint
│   ├── core/           # TS Control plane (Legacy fallback)
│   └── ui/             # React dashboard (Primary UI)
├── submodules/         # External repositories actively being integrated
├── archive/            # Retired submodules or legacy ports
└── docs/               # Architecture, API, and LLM instructions
```
