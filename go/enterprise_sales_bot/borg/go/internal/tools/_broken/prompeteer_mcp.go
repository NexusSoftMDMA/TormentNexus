package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
)

type prompt struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

var prompts = []prompt{
	{Name: "greeting", Text: "Hello, how can I help you today?"},
	{Name: "farewell", Text: "Goodbye! Have a great day."},
}

func HandleListPrompts_prompeteer_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data, e := json.Marshal(prompts)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal prompts: %v", e))
}

	return success(string(data))
}

func HandleGetPrompt_prompeteer_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name argument is required")
}

	for _, p := range prompts {
		if p.Name == name {
			data, e := json.Marshal(p)
			if e != nil {
				return err(fmt.Sprintf("failed to marshal prompt: %v", e))
}

			return success(string(data))

	}
	return err(fmt.Sprintf("prompt '%s' not found", name))
}
}