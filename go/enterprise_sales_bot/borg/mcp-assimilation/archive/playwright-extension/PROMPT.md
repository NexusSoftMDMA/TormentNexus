# Task: Assimilate playwright-extension MCP Server

## Goal
Fully assimilate the `playwright-extension` MCP server into the native Go codebase.

## Details
- Type: STDIO
- Command/URL: npx
- Args: ['-y', '@playwright/mcp@latest', '--extension']

## Steps
1. **Add as submodule**: If not already native, clone/submodule repo.
2. **Analyze source**: Read tool handlers and definitions.
3. **Re-implement**: Build native Go handler under `go/internal/tools/`.
4. **Register**: Add mapping in `registry.go`.
5. **Verify**: Run tests to confirm correctness.
6. **Retire**: Remove submodule and clean directories.
