package mcpimpl

import (
	"context"
)

func HandleGetServerInfo_lightning_enable_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Lightning Enable MCP server")
}