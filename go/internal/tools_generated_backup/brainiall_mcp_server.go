package tools

import (
	"context"
	"fmt"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok(fmt.Sprintf("Hello, %s! Welcome to Brainiall MCP Server.", name))
}

func HandleAnalyze(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("Missing input parameter")
}

	result := fmt.Sprintf("Analysis of '%s': [Brainiall says it's interesting]", input)
	return ok(result)
}