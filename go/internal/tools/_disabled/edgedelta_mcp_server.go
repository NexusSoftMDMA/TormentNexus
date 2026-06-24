package tools

import "context"

func HandleGetStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Edgedelta server is running")
}

func HandleGetVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Edgedelta MCP Server v1.0.0")
}