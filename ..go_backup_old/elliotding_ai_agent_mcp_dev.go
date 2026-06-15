package tools

import (
	"context"
	"encoding/json"
)

func HandleGetTools(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tools := []map[string]interface{}{
		{"name": "weather", "description": "Get current weather data"},
		{"name": "calculator", "description": "Perform arithmetic calculations"},
		{"name": "translate", "description": "Translate text between languages"},
	}
	data, e := json.Marshal(tools)
	if e != nil {
		return err("failed to marshal tools")
}

	return ok(string(data))
}// touch 1781132125
