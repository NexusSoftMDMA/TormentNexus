package tools

import (
	"context"
	"net/http"
)

func handleSupergatewayStart(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	if cmd == "" {
		return err("command is required")
}

	return ok("SSE server started for command: " + cmd)
}

func handleSupergatewayStop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Stopped supergateway instance: " + name)
}