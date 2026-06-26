package mcpimpl

import "context"

func HandleGetQuarterbackStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return success("Quarterback stats for " + name)
}