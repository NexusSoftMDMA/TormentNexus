package tools

import "context"

var memoryStore = map[string]string{}

func HandleStore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	memoryStore[key] = value
	return ok("stored: " + key)
}

func HandleRetrieve(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, found := memoryStore[key]
	if !found {
		return err("key not found")
}

	return ok("value: " + value)
}