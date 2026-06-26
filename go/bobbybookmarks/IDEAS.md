# Ideas & Expansions

## 1. Decentralized Scraping Swarm
- **Pivot:** Instead of centralizing the research worker on a single instance, create a distributed scraping task queue that can be operated by lightweight client worker nodes. Nodes pull target URLs from an external Redis/Supabase instance and return results into `borg.db`.

## 2. Graph Database Transition
- **Refactor:** Migrate the complex relational logic of `tormentnexus.db` and the lost `catalog.db` relationships into Neo4j or another graph-based database to more easily construct the Knowledge Graph backbone highlighted in `VISION.md`.

## 3. WebAssembly (Wasm) Frontend Integration
- **Port:** Move the high-speed data enrichment tools natively into the browser using Wasm compiled from the existing Go codebase. This would lessen the server load drastically when scanning locally fetched markdown repositories.

## 4. MCP Proxy Node Extension
- **Feature:** Stand up an MCP proxy node that allows external agent instances (like Claude Desktop) to natively query the `atlas.db` dataset using standard JSON-RPC calls.

## 5. Aggressive Local LLM Offloading
- **Re-architecture:** Incorporate an explicit connection class for `ollama` or `llama.cpp` to aggressively offload non-critical categorization queries off paid-APIs onto local hardware.
