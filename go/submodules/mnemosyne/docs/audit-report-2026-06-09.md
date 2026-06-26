# Mnemosyne Documentation Audit Report — 2026-06-09

**Version:** 3.4.0
**Pages audited:** 67 stale + 8 unaudited
**Issues found:** 71 total (30 BROKEN, 33 STALE, 7 WRONG, 1 duplicate)
**Issues fixed:** 71

## Summary

Full end-to-end audit of all 67 stale_version pages plus 8 unaudited pages. Cross-referenced every page against v3.4.0 codebase. Three subagent batches covered API, architecture, config/deployment, comparison, migration, operations, retrieval, memory-systems, security, and use-cases pages. All findings fixed with patch tool. Generator updated. Pipeline hardened.

## Key Findings

### Generator (source of truth)
- **Ghost tool `mnemosyne_end`** removed — had zero handler implementation, would raise ValueError
- **CONFIG_ENTRIES: 7 of 9 env vars were fictional** (`MNEMOSYNE_DB_PATH`, `MNEMOSYNE_EMBEDDING_DEVICE`, `MNEMOSYNE_BEAM_ENABLED`, `MNEMOSYNE_BEAM_CPU_THRESHOLD`, `MNEMOSYNE_BEAM_COLD_THRESHOLD`, `MNEMOSYNE_MCP_HOST`, `MNEMOSYNE_MCP_PORT`)
- **Wrong default:** embedding model shown as `all-MiniLM-L6-v2`, actual is `BAAI/bge-small-en-v1.5`
- **Replaced with 19 real env vars** derived from actual `os.environ.get` / `_env_truthy` / `_env_disabled` calls
- **Config injection bug fixed** — was appending instead of replacing, creating triplicate tables
- **Tool count corrected:** 24 → 23 (23 tools with real handlers in `_TOOL_HANDLERS`)

### API Pages (4 pages, 21 issues)
- `mnemosyne_triple_end` completely fabricated — removed from mcp-server.mdx
- `recall` schema used `top_k` instead of `limit` in hermes-plugin.mdx
- `sleep` schema missing `all_sessions` parameter
- `TripleStore` import path wrong in python-sdk.mdx
- 16 stale param required/optional mismatches in tool-schema.mdx

### Architecture Pages (4 pages, 9 issues)
- Entity/fact extraction still referenced TripleStore (post-E6 migration to AnnotationStore)
- `EventType.MEMORY_STORED` renamed to `MEMORY_ADDED` — 3 streaming code examples broken
- `DeltaSync.apply_delta()` missing required `peer_id` param
- `DeltaSync.sync_from()` used invalid `table="memory"` (must be `working_memory` or `episodic_memory`)
- system-design.mdx claimed "no built-in synchronization" (DeltaSync exists since Phase 8)
- CompressionPlugin omitted from built-in plugins tables

### Config + Deployment (4 pages, 15 issues)
- 8 fictional env vars in configuration.mdx static table
- 1 wrong default (embedding model)
- Triplicate injection bug producing 3 copies of stale table
- Dockerfile diverged from real repo Dockerfile (Python 3.12 vs 3.11, CMD vs ENTRYPOINT, unnecessary sqlite3)

### Comparison (7 pages, 12 issues)
- All 7 comparison pages said "6 tools" — actual count 23. Fixed everywhere.

### Migration (2 pages, 2 issues)
- overview.mdx: "v3.1" → "v3.4.0"
- from-hindsight.mdx: "v3 includes" → "v3.4.0 includes"

### Operations (4 pages, 17 issues)
- monitoring.mdx: entirely fictional /health and /metrics endpoints, fictional semantic_memory table, wrong version 3.1.2. Rebuilt with real CLI commands.
- troubleshooting.mdx: 4 fictional exception classes, 2 fictional methods, fictional MNEMOSYNE_LOG_LEVEL
- migration.mdx: fictional migration function names (m_1_0_to_2_0 etc.)
- performance.mdx: semantic_memory references

### Retrieval (5 pages, 4 issues)
- hybrid-search.mdx: temporal formula SIGN wrong (using subtraction, code uses addition)
- temporal_halflife default: 168h → 24h
- fts5-search.mdx: wrong FTS5 syntax

### Remaining Architecture (14 pages)
- sleep-consolidation.mdx: TinyLlama → MiniCPM5-1B (2 references)
- semantic.mdx: mnemosyne_triple_end → mnemosyne_invalidate

### Other Pages (6 issues)
- api/rest.mdx: MNEMOSYNE_DB_PATH → MNEMOSYNE_DATA_DIR
- api/overview.mdx: 6 → 23 tools
- setting.mdx: triple_end → invalidate

## Pipeline Fixes

- **verify-docs.py:** Fixed cwd bug (git diff ran from wrong repo). Added `cwd=docs` to subprocess calls.
- **generate-docs.py:** Rewritten — 23 real tools, 19 real env vars, idempotent config injection.
- **Pre-commit hook:** Installed at `.git/hooks/pre-commit` — blocks commits with doc drift.

## Checkpoint

- **stale_version:** 67 → 0
- **drifted:** 12 → 0
- **clean:** 59 → 137
- **unaudited:** 8 → 9
- **last_full_audit:** 2026-06-09

## Remaining

- **9 unaudited pages:** Root config files (AGENTS.md, CLAUDE.md, README.md) + audit reports + updating.mdx + hermes integration page. Intentional — not documentation content.
- **L2 Defense not deployed:** Watchdog cron + Playwright smoke test remain for a follow-up session.

## Commits

- mnemosyne: `2ad88cb` — generator rewrite
- mnemosyne-docs: `9678f20b` — checkpoint final
- 14 commits total across both repos
