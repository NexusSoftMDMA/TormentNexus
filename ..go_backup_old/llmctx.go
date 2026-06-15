package tools

import "context"

func GetContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	return success("Context for " + key + ": sample value")
}

func ListContexts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Available contexts: default, user, system")
}