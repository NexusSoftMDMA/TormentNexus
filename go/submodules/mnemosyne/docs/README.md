# Mnemosyne Documentation

Local-first, zero-cloud memory for AI agents. SQLite-backed. Sub-millisecond. Fully private.

## Guides

| Document | Description |
|---|---|
| [Getting Started](getting-started.md) | Installation, quickstart, storing your first memory |
| [Architecture](architecture.md) | BEAM tiers, SQLite backend, hybrid search, knowledge graph |
| [API Reference](api-reference.md) | Python API: `remember`, `recall`, `sleep`, triples, stats |
| [Integrations](integrations/README.md) | Platform guides: Cursor, Claude Code, Codex, OpenWebUI, Windsurf, OpenClaw, Hermes |
| [OpenWebUI Deep Integration](integrations/openwebui-deep.md) | Auto-save every chat, memory browser dashboard, cross-session recall |
| [Integration Template](integrations/integration-template.md) | ~100-line pattern for adding any new platform |
| [Hermes Integration](hermes-integration.md) | Using Mnemosyne as a Hermes memory backend |
| [LLM Installation Guide](llm-installation-guide.md) | Installation instructions for AI agents/LLMs |
| [Configuration](configuration.md) | Environment variables, data directory, vector compression |
| [Mnemosyne Sync](sync/index.md) | Bidirectional sync, encryption, deployment, tutorial |
| [Security & Privacy](security.md) | Threat model, encryption internals, BYOK comparison |
| [Benchmarking](benchmarking.md) | Maintainer guide: per-tool A/B benchmark env vars, diagnostics, pure-recall mode, test sequence template |
| [Benchmark Results Analysis](benchmark-results-analysis.md) | Output-file schemas + analysis recipes (per-ability scores, paired bootstrap CIs, voice attribution). AI-assistant-friendly reference |
| [Changelog](changelog.md) | Version history and release notes |

## Quick Links

- **Repository:** [github.com/AxDSan/mnemosyne](https://github.com/AxDSan/mnemosyne)
- **PyPI:** [mnemosyne-memory](https://pypi.org/project/mnemosyne-memory/)
- **Contributing:** See [CONTRIBUTING.md](../CONTRIBUTING.md) in the repo root
- **Updating:** See [UPDATING.md](../UPDATING.md) for upgrade and rollback instructions
- **License:** MIT -- see [LICENSE](../LICENSE)
