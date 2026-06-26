package tools

import (
	"context"
	"strings"
)

func HandleCreateMCPApp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectName, _ :=getString(args, "project_name")
	template, _ :=getString(args, "template")
	if projectName == "" {
		return err("project_name is required")
}

	msg := "Created MCP app: " + projectName
	if template != "" {
		msg += " using template: " + template
	}
	return success(msg)
}

func HandleListTemplates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	templates := []string{"basic", "advanced", "typescript", "python"}
	tmplStr := strings.Join(templates, ", ")
	return success("Available templates: " + tmplStr)
}