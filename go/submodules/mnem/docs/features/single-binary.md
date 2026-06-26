# Single Binary

The `mnem` binary is ~40 MB and self-contained. There is no daemon to start, no cloud account to create, no database server to manage.

Install it, run `mnem init`, and you're writing and retrieving knowledge. Everything runs in-process.

## What's included

- Embedded redb key-value store (the graph backing store)
- HNSW vector index
- Prolly tree chunker
- Ingest pipeline (parse, chunk, entity extraction)
- MCP server (stdio)
- HTTP server
- CLI

All of these are the same binary, selected by subcommand.

## Offline operation

mnem has no required network calls at runtime. The bundled ONNX embedder runs entirely in-process. Ingest, retrieval, branching, merging - everything works with the laptop offline.

Optional integrations (Ollama, OpenAI, Cohere) use the network only when explicitly configured.

## Binary size

The ~40 MB size comes primarily from the bundled ONNX runtime and the MiniLM-L6-v2 model weights. Build without `--features bundled-embedder` to get a smaller binary that uses an external embedder.

## See also

- [Install](../src/install.md)
- [Configuration](../src/configuration.md)
