package tools

import (
	"context"
	"encoding/json"
	"strings"
)

func HandleParseDesignSystem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("input is required")
	}
	tokens := strings.Fields(input)
	data, e := json.Marshal(tokens)
	if e != nil {
		return err("failed to marshal tokens")
	}
	return ok(string(data))
}

func HandleGenerateComponent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
	}
	component := "// Generated for: " + prompt + "\nfunction MyComponent() { return <div>Hello</div>; }"
	return success(component)
}