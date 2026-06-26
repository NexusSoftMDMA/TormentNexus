package tools

import (
	"context"
	"fmt"
)

func HandleCompose(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	engine, _ :=getString(args, "engine")
	text, _ :=getString(args, "text")
	if engine == "" {
		return err("engine is required")
	}
	if engine != "monolithic" && engine != "zettel" {
		return err("engine must be 'monolithic' or 'zettel'")
	}
	return ok(fmt.Sprintf("Composed with '%s' engine: %s", engine, text))
}