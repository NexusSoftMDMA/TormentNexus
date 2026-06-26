.PHONY: build build-cli build-all run test test-cli lint docker clean

BINARY := arc-relay
CLI_BINARY := arc-sync

build:
	CGO_ENABLED=1 go build -o $(BINARY) ./cmd/arc-relay

build-cli:
	CGO_ENABLED=0 go build -o $(CLI_BINARY) ./cmd/arc-sync

build-all: build build-cli

run: build
	./$(BINARY) --config config.example.toml

test:
	CGO_ENABLED=1 go test ./...

test-cli:
	CGO_ENABLED=0 go test ./internal/cli/... ./cmd/arc-sync/...

lint:
	go vet ./...
	CGO_ENABLED=1 golangci-lint run ./...

docker:
	docker compose build

clean:
	rm -f $(BINARY) $(CLI_BINARY)
	rm -rf dist/
