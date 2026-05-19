# Deployment Instructions

## HyperCode `1.0.0-alpha.62`

### 1. Build Requirements
- Node.js >= 24
- pnpm >= 10.28
- Go >= 1.26 (For the authoritative Go Kernel)

### 2. Standard Build & Run
```bash
pnpm install
pnpm run build
pnpm run start
```
This will compile the TypeScript monorepo, build the web assets, and launch the primary Node.js `cli-orchestrator` along with the web dashboard on port `3000`.

### 3. HyperCode Go Sidecar Kernel
To run the Go bridge alongside the main control plane:
```bash
cd go
go run ./cmd/borg serve
```

### 4. Extensions
To build extensions (VS Code, Chrome):
```bash
pnpm run build:extensions
```

### 5. Production Docker
```bash
docker build -f Dockerfile.prod -t hypercode:latest .
docker run -p 3000:3000 -p 4000:4000 -v hypercode-data:/root/.borg hypercode:latest
```

## Health Checks
- `http://localhost:4000/api/config/status` - Main control plane health
- `http://localhost:3000` - Dashboard
