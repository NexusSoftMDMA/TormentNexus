package tools

import "context"

var memory = make(map[string]string)

func HandleStore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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

func HandleRecall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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