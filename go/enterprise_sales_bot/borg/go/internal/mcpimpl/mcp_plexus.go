package mcpimpl

import "context"

func HandlePlexus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return ok("Hello from Plexus, " + name)
}