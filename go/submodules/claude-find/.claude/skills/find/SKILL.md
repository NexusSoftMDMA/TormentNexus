---
name: find
description: Search past Claude Code sessions for context. Use when user wants to recall or pull in context from a previous session or conversation.
---

Use the `search_sessions` MCP tool (from the claude-find server) to search past Claude Code sessions.

Pass the user's query to the tool. If the user said `/find auth discussion`, use "auth discussion" as the query.

**Scope rules:**
- Default: search only the current project (`scope: "current"`)
- If user says "across all projects" or "in all projects": use `scope: "all"`
- If user mentions a specific project name: use `project_filter` with that name

After receiving results, synthesize the key context for the user — including reasoning, decisions, failed approaches, and constraints from the past session. Present it as useful context for the current conversation.

If no results are found, suggest trying `scope: "all"` to search across all projects, or different search terms.
