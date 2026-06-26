package mcpimpl

import "context"

var memory_piia_engram = make(map[string]string)

func HandleStore_piia_engram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	content, _ :=getString(args, "content")
	if key == "" {
		return err("key is required")
}

	if content == "" {
		return err("content is required")
}

	memory[key] = content
	return ok("memory stored")
}

func HandleRecall_piia_engram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	content, found := memory[key]
	if !found {
		return err("memory not found")
}

	return ok(content)
}