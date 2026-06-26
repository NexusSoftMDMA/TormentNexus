# WASM and Edge Deployment

`mnem-core` - the retrieval and storage engine - has no tokio, no filesystem calls, and no network dependencies. The same code that runs on your laptop compiles unchanged to `wasm32`.

This means mnem runs anywhere:

- **Browser tab** - Chrome, Firefox, Safari via WASM
- **Cloudflare Workers** - edge retrieval with zero cold-start overhead
- **AWS Lambda** - serverless, no persistent daemon
- **Air-gapped environments** - no outbound network required at any point

## Why this is unusual

Most agent-memory systems are Python stacks that depend on external databases (PostgreSQL, Redis, Neo4j, Chroma). They require a running server, a network connection, and a specific OS. They cannot run at the edge.

mnem's Rust core was designed from the start with `no_std`-compatible targets in mind. The embedded redb store and the HNSW index both compile to WASM without modification.

## Status

WASM builds are tested in CI. The `wasm32-unknown-unknown` target is the canonical edge target. Browser and Workers bindings ship separately from the CLI binary.

## See also

- [Install](../src/install.md)
- [mnem-core crate](../../crates/mnem-core/)
