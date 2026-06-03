# TormentNexus TORMENTNEXUS Architecture

## Overview
TormentNexus TORMENTNEXUS is a "Go-Powered Modular Monolith" designed for high-performance AI orchestration. It separates high-level user interaction and visualization (TypeScript/React) from low-level control, tool execution, and state management (Go).

## Core Components

### 1. Control Plane (Go) - `go/`
The authoritative engine of the system.
- **Port**: 4300
- **SkillStore**: Manages JIT skill and tool disclosure using BM25 ranking and LRU eviction.
- **EventBus**: High-frequency, resilient message broker for Swarm events.
- **PairOrchestrator**: Enforces the "Planner -> Implementer -> Tester -> Critic" collaboration cycle.
- **Vault**: Secure persistence for sessions, memories (L1/L2), and secrets.

### 2. Node.js Bridge - `packages/core`
Middleware layer connecting the UI to the Go sidecar.
- **Port**: 4100
- **NativeSidecarDaemon**: Manages the Go binary lifecycle.
- **tRPC Routers**: Provides type-safe APIs for the dashboard.
- **ResilientStream**: Buffers backend events to handle frontend socket drops.

### 3. Frontend Dashboard - `apps/web` & `packages/ui`
Reactive management interface.
- **Framework**: Next.js 16 / React 19 / Tailwind CSS 4.
- **Universal Responsiveness**: Uses `useResizeObserver` for dynamic canvas-based visualizations like the `KnowledgeGraph`.
- **Swarm Visualizer**: Real-time neural transcript viewer.

### 4. TormentNexus Supervisor - `packages/tormentnexus-supervisor`
Watchdog and automation agent.
- **Automation**: Uses PowerShell and Windows UI Automation to interact with external AI chat surfaces (Antigravity, Gemini, Claude).
- **Autopilot**: Implements an intelligent "bump" cycle to maintain development momentum autonomously.

## Communication Patterns
- **tRPC**: Primary API for command and control.
- **SSE (Server-Sent Events)**: Real-time event streaming from the Go sidecar.
- **JSON-RPC**: Standard communication for MCP servers.

## Memory Hierarchy (Hippocampus)
- **L1**: Short-term, in-memory session context.
- **L2**: Long-term, semantically indexed via SQLite-vec in the Go Vault.
- **TrafficObserver**: Passive fact extraction from system traffic.
