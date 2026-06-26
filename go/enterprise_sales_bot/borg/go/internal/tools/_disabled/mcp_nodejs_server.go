package tools

import "context"

func HandleNodeVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("Node.js version: v18.16.0")
}