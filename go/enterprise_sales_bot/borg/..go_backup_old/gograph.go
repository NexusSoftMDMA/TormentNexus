package tools

import (
	"context"
	"fmt"
)

func HandleAddNode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	label, _ :=getString(args, "label")
	msg := fmt.Sprintf("Added node '%s' with label '%s'", name, label)
	return success(msg)
}

func HandleQueryGraph(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	msg := fmt.Sprintf("Executed query: %s", query)
	return success(msg)
}