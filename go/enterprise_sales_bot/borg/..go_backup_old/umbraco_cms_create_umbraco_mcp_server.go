package tools

import (
	"context"
	"os"
)

func HandleCreateUmbracoMcpServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "projectName")
	if name == "" {
		return err("projectName is required")
}

	e := os.Mkdir(name, 0755)
	if e != nil {
		return err("failed to create project directory: " + e.Error())
}

	return ok("Umbraco MCP server project created in directory: " + name)
}