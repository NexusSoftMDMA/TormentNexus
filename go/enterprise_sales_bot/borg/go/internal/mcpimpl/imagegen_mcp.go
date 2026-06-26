package mcpimpl

import (
	"context"
	"encoding/json"
)

func HandleGenerateImage_imagegen_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	size, _ :=getString(args, "size")
	result := map[string]string{
		"image_url": "https://example.com/generated.png?prompt=" + prompt + "&size=" + size,
		"prompt":    prompt,
	}
	data, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal result")
}

	return success(string(data))
}