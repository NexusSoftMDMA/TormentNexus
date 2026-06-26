package mcpimpl

import (
	"context"
)

// HandleSelect returns a sample result for a SELECT query.
func HandleSelect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	table, _ :=getString(args, "table")
	return success("Selected from table: " + table)
}

// HandleExecute returns a sample result for an EXECUTE statement.
func HandleExecute_flamerobin_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("Executed query: " + query)
}