# MCP Assimilation Area

## Scope
The goal of this area is to fully assimilate high-value MCP servers from `~/.tormentnexus/mcp.json` into the native Go codebase.

## Workflow
For each server in the MCP config:
1. **Ingest**: Add the server's GitHub repo as a Git submodule.
2. **Analyze**: Read the source code, identify all features, API endpoints, and implementation patterns.
3. **Document**: Write detailed documentation of features and implementation notes.
4. **Re-implement**: Re-write the functionality as a native Go module in `internal/tools/`.
5. **Verify**: Ensure parity and improve where possible (performance, safety, native integration).
6. **Retire**: Remove the submodule once the implementation is completely redundant.

## Conventions
- New Go modules should be placed in `internal/tools/`.
- Documentation should be stored in `docs/mcp/` or within the tool's internal comments.
- Ensure all `go.mod` dependencies are correctly updated.
- Remove all submodules before marking a server as "Assimilated".

## Success Criteria
- Server is no longer required as a submodule.
- All identified features are available as native Go functions.
- Documentation is complete.
- Code is tested and integrated into the main executable.
