package tools

import (
	"context"
	"sync"
)

var store = make(map[string]string)
var mu sync.RWMutex

func HandleSet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" || value == "" {
		return err("key and value are required")
}

	mu.Lock()
	store[key] = value
	mu.Unlock()
	return success("set successfully")
}

func HandleGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	mu.RLock()
	value, found := store[key]
	mu.RUnlock()
	if !found {
		return err("key not found")
}

	return ok(value)
}