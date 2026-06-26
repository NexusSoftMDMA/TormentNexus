# Borg Intelligence Atlas

AI/developer tools and resources database with automated ingestion, LLM-powered enrichment, and data hygiene.

## Current Stats (v8)

| Metric | Value |
|--------|-------|
| Total entries | **7,607** |
| Enriched | 7,607 (100%) |
| Standout | 2,037 |
| Layers | 14 |

## Layer Index

| # | Layer | Tools | Standout | Avg Signal | File |
|---|-------|-------|----------|------------|------|
| 1 | Agent Orchestration & Workflow | 2,248 | 572 | 79 | [AGENT_ORCHESTRATION_WORKFLOW.md](AGENT_ORCHESTRATION_WORKFLOW.md) |
| 2 | Context Engineering & Isolation | 568 | 164 | 80 | [CONTEXT_ENGINEERING_ISOLATION.md](CONTEXT_ENGINEERING_ISOLATION.md) |
| 3 | Memory & Persistence Architecture | 286 | 86 | 73 | [MEMORY_PERSISTENCE_ARCHITECTURE.md](MEMORY_PERSISTENCE_ARCHITECTURE.md) |
| 4 | Interface & Developer UX | 608 | 139 | 78 | [INTERFACE_DEVELOPER_UX.md](INTERFACE_DEVELOPER_UX.md) |
| 5 | Connectivity / MCP / A2A | 759 | 170 | 73 | [CONNECTIVITY_MCP_A2A.md](CONNECTIVITY_MCP_A2A.md) |
| 6 | Infrastructure & Proxy Layers | 373 | 126 | 84 | [INFRASTRUCTURE_PROXY_LAYERS.md](INFRASTRUCTURE_PROXY_LAYERS.md) |
| 7 | Guides & Industry Trends | 809 | 82 | 70 | [GUIDES_INDUSTRY_TRENDS.md](GUIDES_INDUSTRY_TRENDS.md) |
| 8 | Coding Harness Tools | 189 | 40 | 74 | [CODING_HARNESS_TOOLS.md](CODING_HARNESS_TOOLS.md) |
| 9 | AI Agents & Frameworks | 621 | 283 | 86 | [AI_AGENTS_FRAMEWORKS.md](AI_AGENTS_FRAMEWORKS.md) |
| 10 | Search & Discovery | 106 | 42 | 84 | [SEARCH_DISCOVERY.md](SEARCH_DISCOVERY.md) |
| 11 | Coding Tools & IDEs | 210 | 49 | 80 | [CODING_TOOLS_IDES.md](CODING_TOOLS_IDES.md) |
| 12 | Developer Workflow & Tools | 663 | 232 | 86 | [DEVELOPER_WORKFLOW_TOOLS.md](DEVELOPER_WORKFLOW_TOOLS.md) |
| 13 | Vector Databases & Embeddings | 42 | 12 | 85 | [VECTOR_DATABASES_EMBEDDINGS.md](VECTOR_DATABASES_EMBEDDINGS.md) |
| 14 | Security & Red Teaming | 125 | 40 | 85 | [SECURITY_RED_TEAMING.md](SECURITY_RED_TEAMING.md) |

Full index: [BORG_ATLAS_INDEX.md](BORG_ATLAS_INDEX.md)

## Pipeline

- **Ingest**: `_ingest3.py` — URLs from `incoming_resources.txt` → `atlas.db`
- **Enrich**: `_research_worker_pass2.py` — LLM classification, descriptions, scoring (via LM Studio)
- **Export**: `_gen_atlas_v8.py` — Regenerates all markdown layer files

## Key Files

- `atlas.db` — SQLite database (primary data store)
- `incoming_resources.txt` — URLs to be ingested
- `_harness_report.py` — Curated AI CLI/TUI coding harness tools report
