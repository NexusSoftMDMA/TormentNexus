package mcpimpl

import "context"

func HandleX_thinair_data(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return ok("Thinair Data MCP server ready")
}

	return ok("Received query: " + query)
}