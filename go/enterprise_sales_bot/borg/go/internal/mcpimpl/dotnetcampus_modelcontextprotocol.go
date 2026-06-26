package mcpimpl

import (
	"context"
	"strconv"
)

func HandleEcho_dotnetcampus_modelcontextprotocol(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return ok(message)
}

func HandleAdd_dotnetcampus_modelcontextprotocol(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	sum := a + b
	return success(strconv.Itoa(sum))
}