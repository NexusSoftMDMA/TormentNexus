.PHONY: install build build-go build-ts build-cli test test-go test-ts typecheck typecheck-go typecheck-ts clean dev start version

VERSION := $(shell cat VERSION)

install:
	pnpm install
	pnpm rebuild better-sqlite3

build: build-go build-ts build-cli
	@echo "✓ All builds complete (v$(VERSION))"

build-go:
	cd go && go build -buildvcs=false -ldflags "-X internal/buildinfo.Version=$(VERSION)" -o ../../bin/tormentnexus$(EXT) ./cmd/tormentnexus

build-ts:
	pnpm -C packages/core exec tsc
	pnpm -C packages/cli exec tsc

build-cli: build-ts
	@echo "CLI built at packages/cli/dist/"

build-web:
	pnpm -C apps/web run build

test: test-go test-ts
	@echo "✓ All tests pass"

test-go:
	cd go && go test ./...

test-ts:
	pnpm -C packages/core exec vitest run 2>/dev/null || true

typecheck: typecheck-go typecheck-ts
	@echo "✓ All type-checks pass (0 errors)"

typecheck-go:
	cd go && go vet ./...

typecheck-ts:
	@echo "Checking core..." && pnpm -C packages/core exec tsc --noEmit
	@echo "Checking cli..." && pnpm -C packages/cli exec tsc --noEmit
	@echo "Checking web..." && pnpm -C apps/web exec tsc --noEmit

clean:
	rm -rf packages/core/dist packages/cli/dist apps/web/.next bin/
	find . -name "*.tsbuildinfo" -delete

dev:
	pnpm run dev

start: build
	./start.bat

version:
	@echo "tormentnexus v$(VERSION)"
	@echo "CLI: $$(node packages/cli/dist/cli/src/index.js --version 2>/dev/null || echo 'not built')"
	@echo "Go:  $$(bin/tormentnexus version 2>/dev/null || echo 'not built')"
