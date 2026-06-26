package mcpimpl

import "context"

func HandleHello_futuretea_yunxiao_mcp_server_linux_amd64(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello, " + name + "!")
}

func HandleEcho_futuretea_yunxiao_mcp_server_linux_amd64(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}