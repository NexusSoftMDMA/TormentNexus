package mcpimpl

import (
	"context"
	"sync"
)

var memoryStore_hasna_mementos = make(map[string]string)
var mu_hasna_mementos sync.Mutex

func HandleRemember_hasna_mementos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	mu.Lock()
	memoryStore[key] = value
	mu.Unlock()
	return ok("stored")
}

func HandleRecall_hasna_mementos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	mu.Lock()
	val, found := memoryStore[key]
	mu.Unlock()
	if !found {
		return err("key not found")
}

	return success(val)
}