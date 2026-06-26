# Ogham MCP — Architecture Overview

## What It Is

Ogham is a **persistent, searchable memory system** exposed as an MCP (Model Context Protocol) server. AI clients (Claude Code, Claude Desktop, Cursor, etc.) connect to it and get tools for storing, searching, and managing memories backed by PostgreSQL + pgvector.

**Version:** 0.8.0
**Author:** Kevin Burns
**License:** MIT
**Python:** ≥3.13
**Build:** hatchling

## High-Level Data Flow

```
AI Client (Claude Code, etc.)
    │
    │  MCP protocol (stdio or SSE)
    ▼
┌─────────────────────────────────┐
│  MCP Server (server.py)         │
│  - FastMCP framework            │
│  - Registers tools + prompts    │
│  - Health check on startup      │
└──────────┬──────────────────────┘
           │
           ▼
┌─────────────────────────────────┐
│  Service Layer (service.py)     │
│  - store_memory_enriched()      │
│  - search_memories_enriched()   │
│  - Enrichment pipeline:         │
│    date extraction, entity      │
│    extraction, importance       │
│    scoring, surprise scoring,   │
│    auto-linking                 │
└──────────┬──────────────────────┘
           │
     ┌─────┴─────┐
     ▼           ▼
┌──────────┐ ┌──────────────────┐
│ Embedding│ │ Database Facade  │
│ Layer    │ │ (database.py)    │
│          │ │                  │
│ embeddings│ │ Delegates to:   │
│ .py      │ │ - SupabaseBackend│
│          │ │ - PostgresBackend│
│ Providers:│ │ - GatewayBackend│
│ - Ollama │ └────────┬─────────┘
│ - OpenAI │          │
│ - Mistral│          ▼
│ - Voyage │   PostgreSQL + pgvector
└──────────┘
```

## Module Map

### Entry Points
| File | Role |
|------|------|
| `server.py` | MCP server startup — validates health, runs FastMCP |
| `cli.py` | Typer CLI — `ogham serve`, `ogham store`, `ogham search`, `ogham init`, etc. |
| `app.py` | FastMCP application instance (`mcp = FastMCP(...)`) |

### Core Pipeline
| File | Role |
|------|------|
| `service.py` | **Orchestrator** — enrichment pipeline for store and search |
| `database.py` | **Facade** — thin delegates to the active backend |
| `embeddings.py` | Embedding generation with multi-provider support + caching |
| `extraction.py` | NLP utilities — date extraction, entity extraction, importance scoring, temporal intent detection |

### Storage Backends (`backends/`)
| File | Role |
|------|------|
| `protocol.py` | Abstract protocol defining the backend interface |
| `supabase.py` | PostgREST/Supabase backend (original, uses `postgrest` library) |
| `postgres.py` | Direct PostgreSQL backend (uses `psycopg` connection pool) |
| `gateway.py` | HTTP gateway backend (delegates to a remote Ogham gateway API) |

### MCP Interface (`tools/`)
| File | Role |
|------|------|
| `memory.py` | All MCP tool definitions — store, search, delete, update, profiles, TTL, linking, graph exploration, compression, re-embed |
| `stats.py` | Cache statistics tool |

### Supporting
| File | Role |
|------|------|
| `config.py` | Pydantic Settings — loads from `.env` or `~/.ogham/config.env` |
| `embedding_cache.py` | SQLite-backed LRU cache for embedding vectors |
| `compression.py` | Tiered memory compression (full → key sentences → one-line summary) |
| `export_import.py` | Profile export (JSON/markdown) and import with dedup |
| `health.py` | Health checks for database + embedding provider |
| `http_health.py` | Optional HTTP health endpoint (for Docker/k8s) |
| `hooks.py` | Claude Code hook integration — auto-capture from conversations |
| `hooks_cli.py` | `ogham hooks install/uninstall/status` subcommands |
| `hooks_install.py` | Hook installation logic |
| `init_wizard.py` | Interactive `ogham init` setup wizard |
| `prompts.py` | MCP prompt templates |
| `retry.py` | Generic retry decorator with exponential backoff |
| `openapi.py` | Generate OpenAPI spec from MCP tool definitions |

## Configuration (`config.py`)

Uses `pydantic-settings` with env file cascade:
1. Project `.env` (highest priority)
2. `~/.ogham/config.env` (global fallback)

Key settings:

| Setting | Default | Purpose |
|---------|---------|---------|
| `DATABASE_BACKEND` | `supabase` | `supabase`, `postgres`, or `gateway` |
| `DATABASE_URL` | — | PostgreSQL connection string (for `postgres` backend) |
| `SUPABASE_URL` / `SUPABASE_KEY` | — | Supabase credentials (for `supabase` backend) |
| `EMBEDDING_PROVIDER` | `ollama` | `ollama`, `openai`, `mistral`, or `voyage` |
| `EMBEDDING_DIM` | provider-dependent | Vector dimensions (ollama=512, others=1024) |
| `OLLAMA_URL` | `http://localhost:11434` | Ollama endpoint |
| `OLLAMA_EMBED_MODEL` | `embeddinggemma` | Ollama model name |
| `DEFAULT_PROFILE` | `default` | Default memory profile |
| `OGHAM_TRANSPORT` | `stdio` | MCP transport (`stdio` or `sse`) |
| `EMBEDDING_CACHE_MAX_SIZE` | 10000 | Max cached embeddings |

## Server Startup Sequence

1. `cli.py:main()` → Typer dispatches to `serve` command
2. `server.py:main()` → calls `validate_startup()`
3. `validate_startup()` checks database connectivity and embedding provider
4. Exits with error if either fails
5. Imports `ogham.prompts`, `ogham.tools.memory`, `ogham.tools.stats` (registers MCP tools)
6. `mcp.run()` — starts FastMCP server on configured transport

## Store Pipeline (`service.py:store_memory_enriched`)

When a memory is stored, it goes through this enrichment pipeline:

1. **Validation** — content must be non-empty
2. **Secret masking** — `hooks._mask_secrets()` redacts sensitive patterns
3. **Date extraction** — `extraction.extract_dates()` finds dates, stores in metadata
4. **Entity extraction** — `extraction.extract_entities()` finds entities, adds as tags
5. **Recurrence extraction** — `extraction.extract_recurrence()` detects recurring patterns (16 languages)
6. **Importance scoring** — `extraction.compute_importance()` scores content significance
7. **Embedding generation** — `embeddings.generate_embedding()` (with cache)
8. **Surprise scoring** — searches for similar existing memories, surprise = 1 - max_similarity
9. **TTL calculation** — checks profile TTL setting, computes `expires_at`
10. **Storage** — delegates to backend
11. **Auto-linking** — finds similar memories and creates relationship edges

## Search Pipeline (`service.py:search_memories_enriched`)

Search is surprisingly sophisticated, with multiple retrieval strategies:

### Standard Path
- Generate query embedding
- Hybrid search (vector similarity + full-text keyword matching via RRF)
- Optional graph traversal (follow relationship edges)
- Optional temporal re-ranking

### Ordering Queries (detected by `is_ordering_query`)
- Elastic K: fetch 10x the limit
- **Strided retrieval** — divides timeline into N equal buckets, round-robin top result from each
- **Entity threading** — two-pass: Pass 1 identifies entities, Pass 2 searches for all occurrences
- Chronological sort

### Multi-hop Temporal Queries (detected by `is_multi_hop_temporal`)
- **Bridge retrieval** — extracts entity anchors, runs separate searches per entity
- Interleaves results round-robin across anchors
- Merges with standard search, logarithmic boost for bridge results
- Entity threading on merged results

### Broad Summary Queries (detected by `is_broad_summary_query`)
- Strided retrieval for timeline coverage
- **MMR re-ranking** (Maximal Marginal Relevance) — balances relevance with diversity

### Temporal Re-ranking
- Gaussian decay centered on anchor date (σ=3 days)
- Directional hard penalty (0.1x for wrong side of "after"/"before")
- Same-day grace period (1.5x boost)
- Timestamp tiebreaker for equal scores

## Database Layer

### Facade Pattern (`database.py`)
All database operations go through `database.py`, which lazily creates a singleton backend based on `settings.database_backend`. The facade is a collection of thin delegate functions — each one just calls `get_backend().method(...)`.

### Backend Selection
- `"supabase"` → `SupabaseBackend` — uses PostgREST client library, calls Supabase RPC functions
- `"postgres"` → `PostgresBackend` — direct SQL via `psycopg` connection pool
- `"gateway"` → `GatewayBackend` — HTTP proxy to a remote Ogham gateway

### Key Operations
- `store_memory()` — insert with embedding vector
- `hybrid_search_memories()` — combined vector similarity + full-text search (RRF)
- `auto_link_memory()` — find similar memories and create edges
- `get_related_memories()` — graph traversal with depth control
- `explore_memory_graph()` — seed search + graph expansion
- `record_access()` — bump access count/timestamp for scoring
- `update_confidence()` — reinforce or contradict a memory

## Embedding System (`embeddings.py`)

### Multi-Provider Architecture
Four providers supported via adapter pattern:
- **Ollama** (default) — local, free, uses `embeddinggemma` model
- **OpenAI** — `text-embedding-3-small`, supports dimension reduction
- **Mistral** — `mistral-embed`
- **Voyage** — `voyage-4-lite`, supports dimension control

### Caching
- SQLite-backed persistent cache (`embedding_cache.py`)
- Cache key = SHA256(provider + dim + text)
- Switching providers/dimensions automatically invalidates cache
- Supports batch embedding with cache-first strategy

### Retry
- `@with_retry(max_attempts=3, base_delay=0.5)` on all provider calls
- Catches `ConnectionError` and `OSError`

## Memory Graph

Memories can be linked via relationship edges:
- **Auto-linking** — on store, finds top-N similar memories (cosine sim > threshold) and creates edges
- **Manual linking** — `store_decision()` accepts `related_memories` UUIDs
- **Graph traversal** — `get_related_memories(depth, min_strength, relationship_types)`
- **Graph-augmented search** — standard hybrid search + follow edges from top results
- **Exploration** — `explore_memory_graph()` combines search seeds with N-hop expansion

Edge types: `supports`, `contradicts`, `similar`, `relates_to`, etc.

## Compression System (`compression.py`)

Tiered compression to manage memory growth:
- **Level 0** (recent) — full text preserved
- **Level 1** (7+ days, low activity) — compressed to key sentences (~30%)
- **Level 2** (30+ days, low activity) — compressed to one-line summary + tags

High-importance, frequently-accessed, or high-confidence memories resist compression. Original content is always preserved for restoration.

## Profile System

Memories are partitioned into **profiles** (like `"default"`, `"work"`, `"personal"`). Each profile is an independent memory namespace:
- Searches only return results from the active profile
- Profiles can have TTL settings (memories auto-expire)
- Export/import operates per-profile

## Hook System (`hooks.py`)

Integration with Claude Code's hook mechanism:
- Auto-captures relevant context from conversations
- Secret masking before storage
- `ogham hooks install` — sets up Claude Code hooks config
- `ogham hooks uninstall` — removes hooks

## SQL Schema

Located in `sql/`:
- `schema.sql` — base schema (Supabase)
- `schema_postgres.sql` — self-hosted PostgreSQL schema
- `sql/migrations/` — incremental migrations (cognitive scoring, confidence, hybrid search, relationships, graph, temporal columns, halfvec compression, etc.)
