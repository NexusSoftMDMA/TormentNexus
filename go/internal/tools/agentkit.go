//go:build ignore
// +build ignore

package tools

import (
	"context"
)

func HandleGetAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = ctx
	return success(`{"agents":[]}`)
}

func HandleCreateAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Agent " + name + " created")
}
