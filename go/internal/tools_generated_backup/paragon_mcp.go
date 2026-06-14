package tools

import (
	"context"
	"encoding/json"
)

func HandleParagonInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data := map[string]interface{}{
		"name":    "Paragon MCP Server",
		"version": "1.0.0",
	}
	bytes, e := json.Marshal(data)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	return ok(string(bytes))
}

func HandleParagonGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello, " + name + "!")
}