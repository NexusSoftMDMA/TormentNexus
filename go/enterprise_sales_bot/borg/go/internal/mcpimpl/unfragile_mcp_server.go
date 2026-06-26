package mcpimpl

import "context"

func HandleDiscover_unfragile_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "artifact_id")
	return ok("discovered artifact: " + id)
}

func HandleRoute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "artifact_id")
	cap, _ :=getString(args, "capability")
	return ok("routing " + id + " to capability " + cap)
}