package mcpimpl

import (
	"context"
	"encoding/json"
)

func HandleFocus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	maxTokens, _ :=getInt(args, "max_tokens")
	if text == "" {
		return err("text is required")
}

	if maxTokens > 0 && len(text) > maxTokens {
		text = text[:maxTokens]
	}
	return success("Focused: " + text)
}

func HandleListBricks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bricks := []string{"code", "docs", "chat", "search", "graph", "memory", "terminal"}
	data, e := json.Marshal(bricks)
	if e != nil {
		return err("failed to marshal")
}

	return ok(string(data))
}