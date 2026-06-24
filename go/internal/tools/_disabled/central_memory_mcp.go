package tools

import (
	"context"
	"sync"
)

var memory sync.Map

func HandleRemember(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	memory.Store(key, value)
	return ok("stored")
}

func HandleRecall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, found := memory.Load(key)
	if !found {
		return err("not found")
}

	return success(value.(string))
}