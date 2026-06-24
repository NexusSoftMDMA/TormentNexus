package tools

import "context"

func HandleDiscover(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "artifact_id")
	return ok("discovered artifact: " + id)
}

func HandleRoute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "artifact_id")
	cap, _ :=getString(args, "capability")
	return ok("routing " + id + " to capability " + cap)
}