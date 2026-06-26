package tools

import (
	"context"
)

func HandleStoreContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	return ok("stored " + key + " = " + value)
}

func HandleRetrieveContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	return ok("retrieved " + key + " = placeholder")
}