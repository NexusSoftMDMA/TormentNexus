package mcpimpl

import "context"

func HandleGetMemory_memorylens_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
	}
	return success("Memory: " + query)
}