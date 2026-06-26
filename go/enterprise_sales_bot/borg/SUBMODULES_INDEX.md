# Submodules Index & Project Structure

_Last updated: 2026-06-25, version 1.0.0-alpha.159_

> **All legacy submodules have been removed as fully redundant.** The remaining submodules are minimal and maintained at their pinned commits via the repo sync protocol.

## Current Active Submodules

| Submodule | Location | Upstream | Commit | Purpose |
|-----------|----------|----------|--------|---------|
| **bobbybookmarks** | `bobbybookmarks/` | `robertpelloni/bobbybookmarks` | `c50f155` (main) | Bookmark sync and scraped data ingestion |
| **enterprise_sales_bot** | `enterprise_sales_bot/` | `robertpelloni/enterprise_sales_bot` | `fdafa92` (main) | Enterprise sales bot with nested borg submodule |
| **borg** (nested) | `enterprise_sales_bot/borg/` | `robertpelloni/borg` | `f3314909` (main) | Core OS module, tracks TormentNexus main |

## Legacy Submodules (Removed)

The following legacy submodules have been removed from `.gitmodules` as fully redundant, their functionality ported to native Go implementations inside `go/internal/`:

- **jules-autopilot** — Cloud orchestration (Go-native replacement in controlplane)
- **Maestro** — GUI client (slated for Wails native replacement)
- **OmniRoute** — LLM provider routing (Go-native provider abstraction)
- **litellm** — LLM proxy (replaced by FreeLLM proxy on port 4000)
- **mcpproxy** — MCP routing (Go-native MCP server)
- **tormentnexus** — Memory ingestion (fully ported to Go MemoryManager)
- **hyperharness** — CLI harness (integrated into Go sidecar)
- **multica** — Multi-agent structures (Go-native A2A protocol)
- **unifyroute** — Fallback chains (Go-native router)
- **coding_agent_usage_tracker** — Billing/tracing (Go-native telemetry)

## Repository Layout

```
tormentnexus/
├── go/                      # Go sidecar (kernel, control plane, tools)
│   ├── cmd/tormentnexus/    # Main binary
│   ├── internal/
│   │   ├── controlplane/   # HTTP API server (Fiber)
│   │   ├── memorystore/    # Vector DB, caching, reinforcement
│   │   ├── mcpimpl/        # MCP tool implementations (~3,900+ handlers)
│   │   └── tools/          # Tool dispatch & registry
│   └── go.mod
├── apps/
│   └── web/                 # Next.js dashboard (port 7779)
├── bobbybookmarks/          # Submodule: bookmark sync
├── enterprise_sales_bot/    # Submodule: sales bot + nested borg
├── data/                    # Database files (.db, assimilated states)
├── scripts/                 # Build and sync utilities
├── .memory/                 # Brain agent memory (roadmap, logs, branches)
├── VERSION                  # Canonical version: 1.0.0-alpha.159
└── CHANGELOG.md             # Full release history
```

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
