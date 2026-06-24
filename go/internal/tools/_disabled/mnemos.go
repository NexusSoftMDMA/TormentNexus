package tools

import (
	"context"
	"sync"
)

var memories sync.Map

func HandleRemember(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" || value == "" {
		return err("missing key or value")
}

	memories.Store(key, value)
	return success("remembered: " + key)
}

func HandleRecall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("missing key")
}

	val, found := memories.Load(key)
	if !found {
		return err("not found: " + key)
}

	return ok(val.(string))
}