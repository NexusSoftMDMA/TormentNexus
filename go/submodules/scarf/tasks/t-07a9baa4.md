---
id: t-07a9baa4
title: MCP per-server keepalive_interval editor field (v0.17)
status: todo
added: 2026-06-21
priority: low
---

## Description

Deferred from the Hermes v0.17 Tier 2 catch-up (low value / most plumbing). v0.17 adds a per-MCP-server `keepalive_interval` config key (seconds; default 180, floor 5) that keeps short-TTL HTTP/SSE MCP sessions alive (tools/mcp_tool.py:1696, :291-292). Hand-editable today; Scarf's YAML patcher preserves it.

To implement (mirror the existing `sseReadTimeout` plumbing):
- HermesMCPServer.swift: add `keepaliveInterval: Int?` (sibling of sseReadTimeout:42) + init param.
- HermesFileService: read `keepalive_interval` in parseMCPServersBlock; add a `setMCPServerKeepalive` patcher mirroring setMCPServerSSETimeout (~:752).
- MCPServerEditorView: optional Int field gated on a re-added `hasMCPKeepalive { atLeastSemver(0,17,0) }` capability flag (+ test refs in HermesCapabilitiesTests v017 cluster).
Risk: LOW (additive, gated).

## Plan



## Artifacts



