package mcpimpl

import (
	"context"
	"strconv"
)

func HandleX_xert_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message is required")
}

	return ok("Echo: " + msg)
}

func HandleY_xert_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	num, _ :=getInt(args, "number")
	if num < 0 {
		return err("number must be non-negative")
}

	sq := num * num
	return success("Square is " + strconv.Itoa(sq))
}