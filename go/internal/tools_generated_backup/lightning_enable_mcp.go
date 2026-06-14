package tools

import (
	"context"
)

func HandleGetServerInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Lightning Enable MCP server")
}