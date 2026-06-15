# Project Structure and Module Diagram

This document maps the current structure of the TormentNexus/TormentNexus repository, shows where the main modules live, and explains how the major pieces connect.
This document maps the current structure of the tormentnexus/tormentnexus repository, shows where the main modules live, and explains how the major pieces connect.

It is intentionally descriptive rather than aspirational: it focuses on what exists in the repo today.

## 1. High-level shape

TormentNexus is a **pnpm monorepo** with four major layers:
tormentnexus is a **pnpm monorepo** with four major layers:

1. `apps/` — operator-facing applications and shells
2. `packages/` — shared libraries and the main TypeScript control plane
3. `go/` — experimental Go sidecar and compatibility port
4. `submodules/` plus `data/` — imported external ecosystems and local knowledge assets

At a product level, the repo currently centers on:

- a TypeScript control plane in `packages/core`
- a CLI in `packages/cli`
- a web dashboard in `apps/web`
- a shared UI layer in `packages/ui`
- a memory and retrieval layer in `packages/memory`
- an experimental Go sidecar in `go/`

## 2. Top-level repository map

```text
tormentnexus/
tormentnexus/
├─ apps/                  Operator-facing applications and shells
├─ packages/              Shared libraries and the main TypeScript backend
├─ go/                    Experimental Go sidecar / cli-orchestrator port
├─ cli/                   Legacy/compatibility CLI wrapper area
├─ data/                  Local data assets, bookmarks, imported knowledge
├─ docs/                  Project and contributor documentation
├─ scripts/               Workspace build/dev/maintenance scripts
├─ submodules/            External upstream code assimilated into the repo
├─ archive/               Archived material and retired code/assets
├─ README.md              Product-level overview
├─ ROADMAP.md             Direction and sequencing
├─ TODO.md                Active work tracking
├─ AGENTS.md              Contributor/agent operating rules
├─ package.json           Root scripts and workspace tooling
└─ pnpm-workspace.yaml    Workspace boundaries
```

## 3. Workspace boundaries

The repository is not one flat package list. The root workspace includes:

- `apps/*`
- `packages/*`
- the nested cloud workspace under `apps/cloud-orchestrator/apps/*`, `apps/cloud-orchestrator/packages/*`, and `apps/cloud-orchestrator/server`

It explicitly excludes:

- `apps/tormentnexus-extension`
- `apps/tormentnexus-extension`
- `apps/vscode`
- everything under `archive/**`

That means some directories in `apps/` exist as product surfaces without being first-class root workspace packages.

## 4. Main architectural picture

```mermaid
flowchart TD
    CLI["packages/cli\nCLI orchestrator"]
    WEB["apps/web\nNext.js dashboard"]
    UI["packages/ui\nshared React UI"]
    CORE["packages/core\nTypeScript control plane"]
    AI["packages/ai\nprovider/model integrations"]
    MEMORY["packages/memory\nmemory + vector retrieval"]
    TOOLS["packages/tools\nshared tool definitions/helpers"]
    TYPES["packages/types\nshared types"]
    MCP["packages/mcp-registry + packages/mcp-client\nMCP metadata and client surfaces"]
    GO["go/\nexperimental Go sidecar"]
    SUB["submodules/tormentnexus\nexternal CLI harness upstream"]
    SUB["submodules/tormentnexus\nexternal CLI harness upstream"]
    DATA["data/\nbookmarks, imports, local knowledge"]

    CLI --> CORE
    WEB --> CORE
    WEB --> UI
    UI --> CORE
    CORE --> AI
    CORE --> MEMORY
    CORE --> TOOLS
    CORE --> TYPES
    CORE --> MCP
    MEMORY --> AI
    GO --> CORE
    CLI --> SUB
    CORE --> DATA
    GO --> DATA
```

## 5. Runtime lanes

The repo exposes three main orchestrator identities plus one experimental sidecar:

| Lane | Primary location | Role |
| --- | --- | --- |
| CLI orchestrator | `packages/cli` | Main local terminal entrypoint (`tormentnexus`) |
| CLI orchestrator | `packages/cli` | Main local terminal entrypoint (`tormentnexus`) |
| Web/dashboard lane | `apps/web` | Operator dashboard and API proxy surface |
| Desktop lane | `apps/maestro` | Electron-based desktop shell |
| Experimental Go lane | `go/` | Go sidecar and bridge-first port work |

There is also a nested cloud stack under `apps/cloud-orchestrator/`, which is its own sub-workspace rather than the primary root dashboard, plus `apps/maestro-go`, which is a Go-adjacent desktop lane rather than the main desktop implementation.

## 6. `apps/` — operator-facing applications

### Main apps

| Path | Purpose | Notes |
| --- | --- | --- |
| `apps/web` | Main Next.js dashboard and operator UI | Primary browser-facing operator surface |
| `apps/maestro` | Electron desktop shell | Main `electron-orchestrator` lane |
| `apps/maestro-go` | Go-adjacent desktop/port lane | Supplementary experimental area |
| `apps/mobile` | Mobile-facing surface | Companion app lane |
| `apps/cloud-orchestrator` | Nested cloud orchestrator stack | Separate mini-workspace with its own apps/packages/server |
| `apps/tormentnexus-extension` | Browser extension application | Exists in repo but excluded from root `pnpm-workspace.yaml` |
| `apps/tormentnexus-extension` | Browser extension application | Exists in repo but excluded from root `pnpm-workspace.yaml` |
| `apps/vscode` | VS Code extension app | Exists in repo but excluded from root `pnpm-workspace.yaml` |

### How `apps/` connect

- `apps/web` depends directly on `@tormentnexus/core` and `@tormentnexus/ui`
- `apps/web` depends directly on `@tormentnexus/core` and `@tormentnexus/ui`
- `apps/web` is the main browser-facing dashboard for the TypeScript control plane
- `apps/maestro` is the broader desktop/operator shell
- `apps/maestro-go` is an experimental Go-adjacent lane, not the main desktop runtime
- `apps/cloud-orchestrator` is its own nested ecosystem with its own `apps/`, `packages/`, and `server/`
- extension apps connect editor/browser environments back to the main control-plane concepts

## 7. `packages/` — shared libraries and the main control plane

### Key packages

| Path | Purpose |
| --- | --- |
| `packages/core` | Main TypeScript control plane, routers, services, orchestration |
| `packages/cli` | Main CLI entrypoint and command surface |
| `packages/ui` | Shared React UI components and dashboard widgets |
| `packages/ai` | Model/provider SDK integration layer |
| `packages/memory` | Memory storage, retrieval, embeddings, vector DB integration |
| `packages/types` | Shared types |
| `packages/tools` | Tool definitions/helpers |
| `packages/mcp-registry` | MCP metadata and registry surfaces |
| `packages/mcp-client` | MCP client integration |
| `packages/mcp-router-cli` | MCP router CLI package |
| `packages/agents` | Agent-related logic and adapters |
| `packages/adk` | Agent development kit layer |
| `packages/search` | Search and indexing support |
| `packages/tormentnexus-supervisor` | Native Windows supervisor bridge for process/input/UI automation |
| `packages/supervisor-plugin` | Supervisor plugin integration |
| `packages/browser` | Legacy browser support package |
| `packages/browser-extension` | Shared browser-extension package |
| `packages/tormentnexus` | Claude/TormentNexus memory-related package surface |
| `packages/tormentnexus-supervisor` | Native Windows supervisor bridge for process/input/UI automation |
| `packages/supervisor-plugin` | Supervisor plugin integration |
| `packages/browser` | Legacy browser support package |
| `packages/browser-extension` | Shared browser-extension package |
| `packages/tormentnexus` | Claude/tormentnexus memory-related package surface |
| `packages/jetbrains` | JetBrains integration package surface |
| `packages/vscode` | VS Code shared extension package |
| `packages/zed-extension` | Zed extension package surface |
| `packages/tsconfig` | Shared TypeScript configuration |

### The most important dependency chain

```mermaid
flowchart LR
    WEB["apps/web"]
    CLI["packages/cli"]
    UI["packages/ui"]
    CORE["packages/core"]
    AI["packages/ai"]
    MEMORY["packages/memory"]
    TYPES["packages/types"]
    TOOLS["packages/tools"]
    MCPR["packages/mcp-registry"]

    WEB --> UI
    WEB --> CORE
    CLI --> CORE
    UI --> CORE
    CORE --> AI
    CORE --> MEMORY
    CORE --> TYPES
    CORE --> TOOLS
    CORE --> MCPR
    MEMORY --> AI
```

### What `packages/core` does

`packages/core` is the hub of the repository.

It contains:

- the TypeScript control plane
- tRPC routers
- MCP server management
- provider routing logic
- session and supervision logic
- memory integration
- orchestration and council-related services
- many operator-facing APIs that the dashboard and CLI call

If you want to understand “where the system comes together,” start in `packages/core`.

## 8. `apps/web` and `packages/ui`

`apps/web` is the main browser dashboard.

It depends on:

- `@tormentnexus/core` for control-plane access and route integration
- `@tormentnexus/ui` for shared components

`packages/ui` is the shared component layer used to avoid duplicating operator UI across apps. In this repo, the documented convention is that `apps/web` should import shared UI from `@tormentnexus/ui`.

## 9. `packages/cli`

`packages/cli` is the main CLI lane and publishes the `tormentnexus` command.

It depends on `@tormentnexus/core`, so the CLI is not a separate backend; it is another surface over the same TypeScript control-plane concepts.
- `@tormentnexus/core` for control-plane access and route integration
- `@tormentnexus/ui` for shared components

`packages/ui` is the shared component layer used to avoid duplicating operator UI across apps. In this repo, the documented convention is that `apps/web` should import shared UI from `@tormentnexus/ui`.

## 9. `packages/cli`

`packages/cli` is the main CLI lane and publishes the `tormentnexus` command.

It depends on `@tormentnexus/core`, so the CLI is not a separate backend; it is another surface over the same TypeScript control-plane concepts.

Conceptually:

```mermaid
flowchart TD
    User["Operator"]
    CLI["tormentnexus CLI\npackages/cli"]
    CLI["tormentnexus CLI\npackages/cli"]
    CORE["TypeScript control plane\npackages/core"]
    Services["routers + services + orchestration"]

    User --> CLI
    CLI --> CORE
    CORE --> Services
```

## 10. `go/` — experimental sidecar and bridge-first port

The `go/` workspace is a separate Go module:

- module: `github.com/NexusSoftMDMA/TormentNexus-go`
- module: `github.com/tormentnexushq/tormentnexus-go`
- top-level layout:
  - `go/cmd/` — Go entrypoints
  - `go/internal/` — Go internal packages and HTTP API

Its role today is **bridge-first** and **coexistence-first**, not full replacement.

In practice, the Go sidecar:

- exposes native Go health/runtime/status surfaces
- mirrors selected operator APIs
- bridges many requests back into the TypeScript control plane
- helps test how much of the operator-facing stack can move into Go safely

### Go sidecar relationship

```mermaid
flowchart LR
    GO["go sidecar"]
    GOAPI["Go HTTP API"]
    TS["packages/core\nTypeScript control plane"]
    LOCK["lock/config/runtime discovery"]
    UI["apps/web or CLI consumers"]

    UI --> GOAPI
    GO --> GOAPI
    GOAPI --> TS
    GOAPI --> LOCK
```

Important truth: the Go workspace is still an **Experimental** sidecar/port lane, not the sole default runtime for the whole project.

## 11. `apps/cloud-orchestrator/` — nested sub-workspace

`apps/cloud-orchestrator` is effectively its own mini-monorepo inside the main repository.

Observed package surfaces include:

| Path | Package name |
| --- | --- |
| `apps/cloud-orchestrator/package.json` | `jules-ui` |
| `apps/cloud-orchestrator/apps/cli/package.json` | `@jules/cli` |
| `apps/cloud-orchestrator/packages/shared/package.json` | `@jules/shared` |
| `apps/cloud-orchestrator/server/package.json` | `@jules/server` |

This means the repo has:

- the primary TormentNexus/TormentNexus workspace at the root
- the primary tormentnexus/tormentnexus workspace at the root
- a nested cloud-orchestrator workspace with its own app/server/shared package structure

That nested stack should be understood as adjacent infrastructure, not the same thing as `apps/web`.

## 12. `submodules/` and external assimilation

The current submodules directory includes at least two important external lanes:

| Path | Purpose |
| --- | --- |
| `submodules/tormentnexus` | Experimental external CLI harness/upstream assimilation lane |
| `submodules/tormentnexus` | Experimental external CLI harness/upstream assimilation lane |
| `submodules/tormentnexus-mcp` | External TormentNexus MCP reference lane |

This submodule matters because the CLI/session/harness story is no longer only local handwritten code; it also depends on tracked external harness contracts and source-backed tool inventories.

## 13. `data/` and knowledge assets

`data/` holds local knowledge assets, ingestion material, and bookmark ecosystems.

This is part of the project’s “control plane plus memory plus operator knowledge” shape rather than a random asset folder.

It connects mainly to:

- import and indexing logic in `packages/core`
- memory/retrieval logic in `packages/memory`
- bookmark and registry workflows
- persistent local artifacts such as `data/bobbybookmarks`, `tormentnexus.db`, and `cipher-sessions.db`
- persistent local artifacts such as `data/bobbybookmarks`, `tormentnexus.db`, and `cipher-sessions.db`

## 14. Connection map by responsibility

| Area | Primary module(s) | Connected to |
| --- | --- | --- |
| Operator UI | `apps/web`, `apps/maestro`, `packages/ui` | `packages/core` |
| CLI orchestration | `packages/cli` | `packages/core`, `submodules/tormentnexus` |
| CLI orchestration | `packages/cli` | `packages/core`, `submodules/tormentnexus` |
| Core control plane | `packages/core` | nearly everything |
| Provider routing | `packages/ai`, `packages/core` | model/provider SDKs, memory, dashboard |
| Memory and retrieval | `packages/memory`, `packages/core` | sessions, imports, dashboard, providers |
| MCP registry and client flows | `packages/mcp-registry`, `packages/mcp-client`, `packages/core` | dashboard, CLI, imported catalogs |
| Tool execution/helpers | `packages/tools`, `packages/core` | CLI, session runtime, MCP/tool UX |
| Editor and extension bridges | `packages/vscode`, `packages/browser-extension`, `packages/jetbrains`, `packages/zed-extension`, `apps/vscode`, `apps/tormentnexus-extension` | `packages/core` |
| Editor and extension bridges | `packages/vscode`, `packages/browser-extension`, `packages/jetbrains`, `packages/zed-extension`, `apps/vscode`, `apps/tormentnexus-extension` | `packages/core` |
| Go coexistence lane | `go/internal/httpapi`, `go/internal/*` | `packages/core`, workspace state, lock/config files |
| Local knowledge assets | `data/bobbybookmarks` and other `data/` files | `packages/core`, `packages/memory` |

## 15. How the main pieces connect in practice

### Request flow: dashboard

```mermaid
sequenceDiagram
    participant Operator
    participant Web as apps/web
    participant UI as packages/ui
    participant Core as packages/core
    participant Memory as packages/memory
    participant AI as packages/ai

    Operator->>Web: Open dashboard / perform action
    Web->>UI: Render shared UI components
    Web->>Core: Call control-plane API / router
    Core->>Memory: Query or update memory when needed
    Core->>AI: Resolve provider/model operations when needed
    Core-->>Web: Return operator-visible state
    Web-->>Operator: Updated UI
```

### Request flow: Go sidecar bridge

```mermaid
sequenceDiagram
    participant Operator
    participant Client as CLI or UI client
    participant Go as go/internal/httpapi
    participant Core as packages/core

    Operator->>Client: Invoke Go-facing endpoint
    Client->>Go: HTTP request
    Go->>Core: Bridge selected procedure to TS control plane
    Core-->>Go: Return result
    Go-->>Client: Structured Go-side response
```

## 16. Repo relationship diagram

```mermaid
flowchart TD
    ROOT["Repository root"]

    APPS["apps/\nweb, maestro, maestro-go, mobile,\ncloud-orchestrator, browser/editor apps"]
    PKGS["packages/\ncore, cli, ui, ai, memory,\nmcp, tools, agents, extensions"]
    GO["go/\nexperimental sidecar"]
    SUBS["submodules/\ntormentnexus, tormentnexus-mcp"]
    SUBS["submodules/\ntormentnexus, tormentnexus-mcp"]
    DATA["data/\nbobbybookmarks, local DBs"]
    DOCS["docs/\nREADME, roadmap, contributor guidance"]

    ROOT --> APPS
    ROOT --> PKGS
    ROOT --> GO
    ROOT --> SUBS
    ROOT --> DATA
    ROOT --> DOCS

    APPS --> PKGS
    GO --> PKGS
    PKGS --> DATA
    PKGS --> SUBS
```

## 17. Where to start reading

If you are new to the repo, the best reading order is:

1. `README.md`
2. `docs/PROJECT_STRUCTURE.md` (this file)
3. `packages/core`
4. `packages/cli`
5. `apps/web`
6. `packages/ui`
7. `packages/memory`
8. `go/`

## 18. Practical mental model

The shortest accurate mental model is:

- `packages/core` is the brain of the current system
- `packages/cli` and `apps/web` are the main operator surfaces
- `packages/ui` is the shared presentation layer
- `packages/ai` and `packages/memory` are major supporting subsystems
- `go/` is an experimental sidecar/port that increasingly mirrors and bridges the TypeScript control plane
- `apps/cloud-orchestrator` is a separate nested stack, not the same thing as the main dashboard
- `submodules/` and `data/` extend the system with external harnesses and local knowledge assets

## 19. Status framing

Truthfully:

- the TypeScript control-plane path is the primary implementation
- the dashboard, CLI, MCP, provider, session, and memory layers are the most mature parts
- council/autonomy/automation-heavy areas are broader and less uniform in maturity
- the Go workspace is real and expanding, but still an **Experimental** coexistence/port lane rather than a complete default replacement

