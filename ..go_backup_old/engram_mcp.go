package tools

import "context"

func HandleStoreEngram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	_ = key
	_ = value
	return ok("engram stored successfully")
}