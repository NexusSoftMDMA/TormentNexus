# YADMS - Yet Another Desktop MCP Service

YADMS is a hardened, standalone desktop service designed for Game Developers and Engineers. It acts as an unbreakable bridge between your local testing environments, workflows, and Model Context Protocol (MCP) clients.

## What It Does
YADMS exposes a unified suite of developer tools directly to your local network via an encrypted, resilient Server-Sent Events (SSE) architecture. It allows MCP-compatible AI systems or testing rigs to securely read files, execute builds, capture screens, and orchestrate underlying system processes.

## Features
- **Total State Resiliency**: YADMS survives daemon reboots and implicitly restores all background jobs, cron tasks, and file watchers without leaving orphaned zombie processes.
- **Unified Build Orchestration**: Run lightweight debugging payloads or embed entire FFmpeg toolchains dynamically via single MSBuild flags.
- **Deep Desktop Access**:
  - Secure Terminal & Process execution
  - Real-time Video, Audio, and Screen capture
  - Advanced File I/O and Window Management
  - Local SSH and WebSocket Bridging
- **Cryptographic Security**: Built-in master key encryption utilizing DPAPI ensures sensitive states (like SSH credentials) remain exclusively bound to your local machine.

## Getting Started
To compile the lightweight service:
```bash
dotnet build controller_mcp/controller_mcp.csproj -p:BumpPatch=false
```

To compile the fully bundled service (includes embedded FFmpeg for A/V toolsets):
```bash
dotnet build controller_mcp/controller_mcp.csproj -p:IncludeFfmpeg=true -p:BumpPatch=false
```

## Setup & Configuration
Launch the executable and configure your `Log Directory` and `FFmpeg` paths from the UI. Your configuration is automatically broadcasted over the IPC interface and securely backed up.

## License
Project is Managed by the Blazium Games Contributors, and Created by Bioblaze Payne.
Licensed under the [MIT License](LICENSE).
