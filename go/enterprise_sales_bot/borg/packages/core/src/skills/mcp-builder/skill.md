---
name: mcp-builder
description: Guide for creating high-quality MCP (Model Context Protocol) servers.
---

# MCP Server Development Guide

## Overview
Create MCP (Model Context Protocol) servers that enable LLMs to interact with external services through well-designed tools.

## High-Level Workflow

### Phase 1: Deep Research and Planning
1.  **Understand Modern MCP Design**:
    - Balance API coverage vs workflow tools.
    - Use clear, descriptive tool names (e.g., `github_create_issue`).
    - Provide concise tool descriptions and actionable error messages.
2.  **Study Documentation**:
    - Protocol Specs: `https://modelcontextprotocol.io/specification/draft.md`
    - Best Practices: `[📋 View Best Practices](./reference/mcp_best_practices.md)`
3.  **Choose Stack**:
    - **TypeScript** (Recommended): High-quality SDK, strong typing.
    - **Python**: Good for data/ML heavy integration.
4.  **Plan Implementation**:
    - Review service API.
    - Select tools to implement (common operations first).

### Phase 2: Implementation
1.  **Project Structure**:
    - TypeScript: `package.json`, `tsconfig.json`, `src/index.ts`.
    - Python: `pyproject.toml`, `server.py`.
2.  **Core Infrastructure**:
    - Auth, Error handling, Pagination.
3.  **Implement Tools**:
    - Input Schema: Zod (TS) or Pydantic (Python).
    - Output Schema: Structured data + text.
    - Description: Summary, parameters, return type.
    - Annotations: `readOnlyHint`, `destructiveHint`.

### Phase 3: Review and Test
- **Code Quality**: DRY, Error handling, Type coverage.
- **Build**: `npm run build` or `py_compile`.
- **Test**: Use `npx @modelcontextprotocol/inspector`.

### Phase 4: Evaluations
- Create 10 independent, read-only, complex questions to verify the server.
- Format as XML evaluations.

## Reference Files
(Note: Source references are available in `external/skills_repos/anthropic-skills/skills/mcp-builder/reference`)

- [Protocol Spec](https://modelcontextprotocol.io/specification/draft.md)
- [TypeScript SDK](https://github.com/modelcontextprotocol/typescript-sdk)
- [Python SDK](https://github.com/modelcontextprotocol/python-sdk)
