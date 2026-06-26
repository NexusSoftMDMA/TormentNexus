package mcpimpl

import (
    "context"
)

func HandlePing_mcpico(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("pong")
}