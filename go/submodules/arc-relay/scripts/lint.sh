#!/bin/bash
# Local lint check - mirrors CI lint job
# Usage: ./scripts/lint.sh
set -e

GO=/usr/local/go/bin/go
GOPATH=$($GO env GOPATH)
export PATH="/usr/local/go/bin:$PATH"

echo "=== go vet ==="
$GO vet ./...

echo "=== gofmt ==="
unformatted=$($GO fmt ./... 2>&1 || true)
bad=$(gofmt -l . | grep -v ".claude/worktrees" || true)
if [ -n "$bad" ]; then
  echo "FAIL: not formatted:"
  echo "$bad"
  gofmt -d $bad
  exit 1
fi
echo "PASS"

echo "=== staticcheck ==="
if [ ! -f "$GOPATH/bin/staticcheck" ]; then
  echo "Installing staticcheck..."
  $GO install honnef.co/go/tools/cmd/staticcheck@latest
fi
$GOPATH/bin/staticcheck ./...

echo "=== golangci-lint ==="
if [ ! -f "$GOPATH/bin/golangci-lint" ]; then
  echo "Installing golangci-lint v2..."
  $GO install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.0
fi
CGO_ENABLED=1 $GOPATH/bin/golangci-lint run ./...

echo "=== tests ==="
$GO test -count=1 ./internal/...

echo ""
echo "All checks passed."
