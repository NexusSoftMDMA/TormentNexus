# TormentNexus Deployment Instructions

_This document contains the latest deployment instructions for the TormentNexus Universal AI Dashboard and Cognitive Control Plane (TormentNexus)._

## Prerequisites

1.  **Node.js**: >= 24.x (For maximum runtime compatibility)
2.  **pnpm**: Recommended package manager (`npm install -g pnpm@10.28`)
3.  **Go**: >= 1.23 (For the Go control plane sidecar)
4.  **Git**: For submodule fetching and version control.

## Initial Setup

1.  **Clone the Repository**:
    ```bash
    git clone https://github.com/NexusSoftMDMA/TormentNexus.git
    cd tormentnexus
    ```

2.  **Initialize Submodules**:
    ```bash
    git submodule update --init --recursive
    ```

3.  **Install Dependencies**:
    ```bash
    pnpm install
    ```
    *Note: If you are running Node 24, you must rebuild `better-sqlite3` bindings post-install:*
    ```bash
    pnpm rebuild better-sqlite3
    ```

4.  **Environment Variables**:
    Copy `.env.example` to `.env` and fill in the required API keys (OpenAI, Anthropic, Gemini, etc.).
    ```bash
    cp .env.example .env
    ```

## Running the Platform

TormentNexus is designed as a long-running service that manages PC memory, CPU, disk, and bandwidth usage.

### Standard Build & Run (Production/Development)

To run the TypeScript monorepo and its dashboard in one shot:
```bash
pnpm run build
pnpm run start
```
This will compile all TypeScript packages, build the web assets, and launch the primary Node.js CLI orchestrator alongside the web dashboard.

### Windows Startup Script

Use the provided startup batch file:
```powershell
.\start.bat
```
`start.bat` defaults to `pnpm run build:workspace` (skipping extension-only build stages for a much faster boot). It also triggers a native-runtime preflight to verify SQLite and Electron bindings.

**Startup Overrides**:
- **Bypass Install**: `set TORMENTNEXUS_SKIP_INSTALL=1`
- **Bypass Build**: `set TORMENTNEXUS_SKIP_BUILD=1`
- **Force Full Monorepo Build**: `set TORMENTNEXUS_FULL_BUILD=1`
- **Bypass Native Preflight**: `set TORMENTNEXUS_SKIP_NATIVE_PREFLIGHT=1`

### Linux/macOS Startup Script

```bash
./start.sh
```

### Start Electron Maestro Separately

Maestro is launched independently from the main control plane:
```bash
pnpm -C apps/maestro start
```

---

## Go Sidecar Kernel

### High-Performance Native Tools
TormentNexus includes native Go implementations for core development tools. Ensure these are in your PATH for maximum performance:
- **Ripgrep (rg)**: Required for high-speed regex search.
- **Anyquery**: SQL interface to various data sources.
- **Supervisor Monitor**: Background watchdog for session inactivity.

To build and run the Go control plane sidecar alongside the main TS engine:
```bash
cd go
go run ./cmd/tormentnexus serve
```
Alternatively, build the binary:
```bash
cd go
go build -buildvcs=false ./cmd/tormentnexus
```

---

## Extensions

To compile VS Code and Chrome/Firefox browser agents:
```bash
pnpm run build:extensions
```

## Production Docker

Build the production bundle inside a container:
```bash
docker build -f Dockerfile.prod -t tormentnexus:latest .
docker run -p 3000:3000 -p 4000:4000 -v tormentnexus-data:/root/.tormentnexus tormentnexus:latest
```

---

## Package Manager Requirement

**pnpm v10 is required.** The root `package.json` locks `packageManager: pnpm@10.28.0`. Using pnpm v9 or below will produce `ERR_PNPM_BAD_PM_VERSION` and fail the build.
```bash
npm install -g pnpm@10
```

## Release Gate Validation

Before checking in code, run the CI release verification pipeline:
```bash
pnpm run check:release-gate:ci
```
This will execute standard placeholder checks, perform type-checking across `packages/core`, and lint the entire workspace.

To perform strict visual and screenshot sync checks:
```bash
pnpm run check:release-gate:ci:strict-visuals
```

---

## Ports & Endpoints

| Service | Default Port | Override | Health Check Endpoint |
|---------|-------------|----------|-----------------------|
| Core API (TS Control Plane) | 4000 | `tormentnexus start --port <n>` | `http://localhost:4000/api/config/status` |
| Web Dashboard (Next.js) | 3000 | `PORT` | `http://localhost:3000` |
| Go Kernel sidecar | 4300 | — | `http://localhost:4300` |
| Socket.io Swarm Server | 3001 | — | — |
| tRPC Bridge Control | 4100 | — | — |
| Orchestrator tRPC | 3847 | `TORMENTNEXUS_ORCHESTRATOR_PORT` | `http://localhost:3847` |

## Health Checks
- `http://localhost:4000/api/config/status` - Main control plane health
- `http://localhost:3000` - Dashboard
