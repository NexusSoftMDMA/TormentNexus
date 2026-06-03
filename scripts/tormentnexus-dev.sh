#!/usr/bin/env bash
# tormentnexus-dev.sh — Quick development launcher
# Usage: ./scripts/tormentnexus-dev.sh [--skip-go] [--skip-install]
set -euo pipefail

cd "$(dirname "$0")/.."
VER=$(cat VERSION 2>/dev/null || echo "dev")

SKIP_GO=false
SKIP_INSTALL=false
for arg in "$@"; do
  case $arg in
    --skip-go) SKIP_GO=true ;;
    --skip-install) SKIP_INSTALL=true ;;
  esac
done

echo "⬡ TormentNexus TORMENTNEXUS v${VER} — Dev Launcher"
echo ""

# 1. Install if needed
if [ "$SKIP_INSTALL" = false ]; then
  echo "[1/4] Installing dependencies..."
  pnpm install --frozen-lockfile 2>/dev/null || pnpm install
  pnpm rebuild better-sqlite3 2>/dev/null || true
else
  echo "[1/4] Skipping install"
fi

# 2. Build Go sidecar
if [ "$SKIP_GO" = false ] && command -v go &>/dev/null; then
  echo "[2/4] Building Go sidecar..."
  (cd go && go build -ldflags "-X github.com/tormentnexushq/tormentnexus-go/internal/buildinfo.Version=${VER}" -buildvcs=false -o ../bin/tormentnexus ./cmd/tormentnexus)
  echo "      ✓ bin/tormentnexus built"
else
  echo "[2/4] Skipping Go build"
fi

# 3. Build TypeScript
echo "[3/4] Building TypeScript..."
pnpm -C packages/core exec tsc 2>/dev/null && pnpm -C packages/cli exec tsc 2>/dev/null
echo "      ✓ TypeScript compiled"

# 4. Launch
echo "[4/4] Starting services..."
echo ""

# Go sidecar (background)
if [ -x bin/tormentnexus ]; then
  bin/tormentnexus -port 4300 &>/dev/null &
  GO_PID=$!
  echo "  Go sidecar:  http://127.0.0.1:4300 (PID $GO_PID)"
fi

# TS control plane (foreground)
echo "  TS server:   http://0.0.0.0:4000/trpc"
echo "  Dashboard:   http://localhost:3000/dashboard"
echo ""
echo "  Press Ctrl+C to stop"
echo ""

cleanup() {
  echo ""
  [ -n "${GO_PID:-}" ] && kill "$GO_PID" 2>/dev/null
  exit 0
}
trap cleanup SIGINT SIGTERM

node packages/cli/dist/cli/src/index.js start --port 4000 "$@"
