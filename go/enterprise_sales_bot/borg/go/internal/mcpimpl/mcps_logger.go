package mcpimpl

import (
	"context"
	"log"
)

func HandleLog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	log.Println(msg)
	return ok("logged: " + msg)
}