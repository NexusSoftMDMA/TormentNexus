package tools

import (
	"context"
)

func HandleSshConnect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("SSH MCP Server is ready")
}