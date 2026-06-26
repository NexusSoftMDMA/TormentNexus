package mcpimpl

import "context"

func HandleSet_redis_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" {
		return err("key is required")
}

	if value == "" {
		return err("value is required")
}

	return ok("key " + key + " set to " + value)
}

func HandleGet_redis_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	return success("value for " + key + " is some_value")
}