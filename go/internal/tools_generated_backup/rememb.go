package tools

import (
	"context"
	"sync"
)

var memStore sync.Map

func HandleRemember(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	val, _ :=getString(args, "value")
	memStore.Store(key, val)
	return ok("stored")
}

func HandleRecall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	val, found := memStore.Load(key)
	if !found {
		return err("not found")
}

	return success(val.(string))
}