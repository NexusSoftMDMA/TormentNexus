#!/usr/bin/env bash
# Thin wrapper — all real logic lives in bin/cli.js.
# Args are forwarded so `start.sh --port 4000 --api-port 9000` works.
set -euo pipefail
cd "$(dirname "$0")"
exec node bin/cli.js "$@"
