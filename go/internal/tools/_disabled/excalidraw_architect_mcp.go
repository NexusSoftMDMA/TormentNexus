package tools

import (
	"context"
	"fmt"
)

func HandleListTemplates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(fmt.Sprintf(`{"templates": ["blank","flowchart","wireframe"]}`))
}

func HandleCreateDiagram(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	template, _ :=getString(args, "template")
	if name == "" {
		return err("name is required")
}

	if template == "" {
		template = "blank"
	}
	return ok(fmt.Sprintf(`{"status":"created","name":"%s","template":"%s"}`, name, template))
}