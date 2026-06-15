package tools

import (
	"context"
	"sync"
)

var memoryStore = make(map[string]string)
var mu sync.Mutex

func HandleRemember(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	mu.Lock()
	memoryStore[key] = value
	mu.Unlock()
	return ok("stored")
}

func HandleRecall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	mu.Lock()
	val, found := memoryStore[key]
	mu.Unlock()
	if !found {
		return err("key not found")
}

	return success(val)
}