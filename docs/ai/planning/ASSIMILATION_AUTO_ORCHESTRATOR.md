# Planning: Assimilation of OpenCode Autopilot (tormentnexus Auto-Orchestrator)

## 🌌 Overview
This document outlines the plan to assimilate `github.com/robertpelloni/opencode-autopilot` into the tormentnexus monorepo as a first-class orchestration layer. The assimilated project will be renamed to **tormentnexus Auto-Orchestrator** and will provide multi-model council supervision for autonomous agent workflows.

## 🏗️ Target Architecture

### Package Information
- **Name**: Integrated directly into `@tormentnexus/core`
- **Location**: `packages/core/src/orchestrator/council`
- **Primary Runtime**: Node.js 22 (migrating from Bun)
- **Framework**: Express/tRPC integration (migrating from Hono)

### Core Components to Migrate
1. **Council Service**: Orchestrates debates between multiple LLM supervisors.
2. **Supervisor Adapters**: Pluggable adapters for OpenAI, Anthropic, Gemini, etc. (to be aligned with `@tormentnexus/core` providers).
3. **Consensus Engine**: Implements majority, weighted, and CEO-override voting models.
4. **Universal PTY Harness**: Native orchestration of terminal-based AI tools (Aider, Claude Code, Gemini CLI, etc.).
5. **Diagram & Swarm Service**: Mermaid-to-Plan parsing and visual architecture generation.
6. **Self-Evolution Engine**: Automated weight optimization and codebase self-modification.
7. **Quota & Analytics**: Provider-level rate limiting and supervisor performance tracking.

## 🔄 Integration Strategy

### 1. Low-Level Substrate Alignment
The Auto-Orchestrator will consume `@tormentnexus/core` for:
- Database access (Drizzle schema).
- Tool discovery (MCP Aggregator).
- Process supervision (Session Supervisor).
- Provider authentication (Provider Truth).

### 2. Dashboard Integration
A new "Council" section will be added to `@tormentnexus/web` (`apps/web`):
- `/dashboard/council`: Overview of active council debates.
- `/dashboard/council/history`: Audit trail of past decisions.
- `/dashboard/council/config`: Configuration for supervisor weights and consensus modes.

### 3. API & Communication
- Endpoints currently in `opencode-autopilot` (Hono) will be migrated to tRPC procedures in `@tormentnexus/core` or a new router in the orchestrator package.
- Real-time updates will continue using WebSockets, likely unified under the main tormentnexus socket.

## 🚀 Phases

### Phase 1: Preparation (Active)
- [x] Analyze source repository (`opencode-autopilot`).
- [x] Define package structure (decided to integrate into `@tormentnexus/core`).

### Phase 2: Foundation & Skeleton
- [x] Copy source files to `packages/core/src/orchestrator/council`.
- [x] Port shared types to local `types.ts` within the council directory.
- [x] Update internal imports to use relative paths instead of `@tormentnexus-orchestrator/shared`.

### Phase 3: Core Logic Migration (In Progress)
- [ ] Refactor `CouncilService` and `ConsensusEngine` to use tormentnexus primitives.
- [ ] Re-implement Supervisor adapters to use tormentnexus's common provider logic.
- [ ] Migrate Hono/Bun routes to Express/tRPC.
- [ ] Integrate with tormentnexus's SQLite database.

### Phase 4: Interface & Wiring
- [ ] Expose council triggers as MCP tools in `@tormentnexus/core`.
- [ ] Add Council Dashboard pages to `apps/web`.
- [ ] Implement Ink-based CLI if standalone usage is required (under `packages/cli`).

## 🛡️ Principles
- **Truthfulness**: Council debates must be transparent and cite evidence from the session context.
- **Isolation**: Debates should happen in isolated contexts to prevent state leakage.
- **Consistency**: Use the same coding standards, linting, and formatting as the rest of tormentnexus.
