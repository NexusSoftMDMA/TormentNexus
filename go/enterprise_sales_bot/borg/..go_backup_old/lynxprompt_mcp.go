package tools

import (
	"context"
	"encoding/json"
)

func HandleListPrompts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompts := []map[string]string{
		{"name": "greeting", "description": "A friendly greeting prompt"},
		{"name": "summarize", "description": "Summarize text"},
	}
	data, e := json.Marshal(prompts)
	if e != nil {
		return err("failed to marshal prompts")
}

	return ok(string(data))
}

func HandleGetPrompt(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name argument is required")
}

	// Simulate fetching a prompt by name
	prompt := map[string]string{
		"name":     name,
		"content": "You are a helpful assistant. Answer concisely.",
	}
	data, e := json.Marshal(prompt)
	if e != nil {
		return err("failed to marshal prompt")
}

	return ok(string(data))
}