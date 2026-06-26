package mcpimpl

import (
	"context"
	"net/http"
	"encoding/json"
)

func HandleGetFact_peacock_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	animal, _ :=getString(args, "animal")
	if animal == "" {
		animal = "peacock"
	}
	fact := animal + "s are known for their colorful feathers."
	return ok(fact)
}

func HandlePing_peacock_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("pong")
}