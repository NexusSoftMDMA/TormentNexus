package tools

import (
	"context"
	"strings"
)

func HandleGetTemplate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter required")
}

	if name == "basic" {
		return ok("Basic MCP server template:\npackage main\n\nfunc main() {\n  // ...\n}")
}

	if name == "advanced" {
		return ok("Advanced MCP server template:\npackage main\n\nfunc main() {\n  // ...\n}")
}

	return err("template not found")
}

func HandleListTemplates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	templates := []string{"basic", "advanced"}
	return success("Available templates: " + strings.Join(templates, ", "))
}