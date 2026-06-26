package mcpimpl

import "context"

func HandleGet_cachly_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	return ok("cached value for '" + key + "' is 'sample-data'")
}

func HandleSet_cachly_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" || value == "" {
		return err("key and value are required")
}

	return ok("stored key '" + key + "' with value '" + value + "'")
}