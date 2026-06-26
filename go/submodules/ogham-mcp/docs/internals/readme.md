# Ogham MCP — Developer Documentation

Complete technical documentation of the ogham-mcp codebase (v0.8.0).

## Start Here

- **[architecture.md](architecture.md)** — System overview, data flow, module map, startup sequence, store/search pipelines

## Core System

| Doc | Covers |
|-----|--------|
| [schema.md](schema.md) | PostgreSQL schema, all indexes, SQL function formulas (ACT-R, RRF, Bayesian, recursive CTE) |
| [search.md](search.md) | 4 retrieval strategies (standard, ordering, multi-hop temporal, broad summary), temporal re-ranking |
| [extraction.md](extraction.md) | NLP enrichment: date/entity/recurrence extraction, importance scoring, query classification |
| [tools.md](tools.md) | Complete MCP tool catalog with parameters and behavior |
| [embeddings.md](embeddings.md) | 4 embedding providers, SQLite cache, batch pipeline, retry |
| [compression.md](compression.md) | Tiered memory compression (full → gist → tags), resistance formula |

## Storage & Backends

| Doc | Covers |
|-----|--------|
| [backends.md](backends.md) | Protocol interface, PostgresBackend (psycopg pool), SQL function catalog |
| [supabase-gateway.md](supabase-gateway.md) | SupabaseBackend (PostgREST), GatewayBackend (HTTP proxy) |

## Integration & Operations

| Doc | Covers |
|-----|--------|
| [hooks.md](hooks.md) | Claude Code lifecycle hooks, 6-step filtering pipeline, secret masking |
| [hooks-cli.md](hooks-cli.md) | Hook CLI commands (install/uninstall/status/test), settings.json management |
| [init-wizard.md](init-wizard.md) | Interactive setup wizard, 7 MCP clients, uvx/Docker modes |
| [health.md](health.md) | Health check system, HTTP health endpoint for container orchestration |
| [prompts.md](prompts.md) | MCP prompt templates, OpenAPI schema generator |
