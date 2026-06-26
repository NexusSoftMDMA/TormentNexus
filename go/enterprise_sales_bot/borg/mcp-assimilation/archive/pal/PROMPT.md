# Task: Assimilate pal MCP Server

## Goal
Fully assimilate the `pal` MCP server into the native Go codebase.

## Details
- Type: STDIO
- Command/URL: uvx
- Args: ['--from', 'git+https://github.com/BeehiveInnovations/pal-mcp-server.git', 'pal-mcp-server']

## Steps
1. **Add as submodule**: If not already native, clone/submodule repo.
2. **Analyze source**: Read tool handlers and definitions.
3. **Re-implement**: Build native Go handler under `go/internal/tools/`.
4. **Register**: Add mapping in `registry.go`.
5. **Verify**: Run tests to confirm correctness.
6. **Retire**: Remove submodule and clean directories.
