package mcpimpl

import "context"

func HandleListGems_bundler_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("List of gems")
}

func HandleGemInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("Missing argument: name")
}

	return success("Info for gem: " + name)
}