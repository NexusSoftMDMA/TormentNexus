# Contributing to Forkit

Forkit is a hosted service — the server source is not open. This repository contains the public documentation.

## How to Contribute

### Report a bug or request a feature
Open an issue: https://github.com/forkit-mcp/docs/issues

Please include:
- What you were trying to do
- What happened vs. what you expected
- Your MCP client (Claude Code, Cursor, etc.) and version
- Relevant `execute_code` output or error message

### Improve the docs

All docs live in `docs/`. PRs are welcome for:
- Fixing errors or outdated information
- Adding examples for common patterns
- Improving clarity for a specific MCP client

### Share a pattern

If you've found a useful multi-agent pattern using Forkit, open a PR adding it to `docs/multi-agent.md`.

## Docs Structure

```
docs/
  tools.md          ← MCP tool reference (all 7 tools)
  codemode-api.md   ← Full codemode.* function reference
  concepts.md       ← Core concepts: tasks, trajectories, x402
  multi-agent.md    ← Coordination patterns and examples
```

## Questions

Open a GitHub Discussion or reach out on X: [@zientesit](https://x.com/zientesit)
