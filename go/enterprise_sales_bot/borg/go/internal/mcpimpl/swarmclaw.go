package mcpimpl

import (
	"context"
)

func HandleSwarmclawEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "Swarmclaw is ready"
	}
	return ok(msg)
}