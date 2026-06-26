package mcpimpl

import "context"

func HandleSearchCode_mcp_code_graph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	return success("Search completed for: " + q)
}

func HandleGetDependencies_mcp_code_graph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	return success("Dependencies for: " + symbol)
}