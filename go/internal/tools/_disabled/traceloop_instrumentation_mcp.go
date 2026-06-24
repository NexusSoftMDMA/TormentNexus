package tools

import (
	"context"
)

func HandleGetInstrumentation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "default"
	}
	return ok("Instrumentation for: " + name)
}