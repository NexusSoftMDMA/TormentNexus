package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGetSummoner_opgg_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok(fmt.Sprintf("Summoner '%s' found", name))
}

func HandleGetStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	region, _ :=getString(args, "region")
	if name == "" || region == "" {
		return err("name and region are required")
}

	return ok(fmt.Sprintf("Stats for %s on %s retrieved", name, region))
}