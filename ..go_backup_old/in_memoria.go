package tools

import (
	"context"
	"sync"
)

var storage sync.Map

func HandleReadNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	val, found := storage.Load(key)
	if !found {
		return err("key not found")
}

	return success(val.(string))
}

func HandleWriteNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" || value == "" {
		return err("key and value are required")
}

	storage.Store(key, value)
	return ok("note stored")
}