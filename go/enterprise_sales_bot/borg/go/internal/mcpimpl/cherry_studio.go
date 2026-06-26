package mcpimpl

import "context"

func HandleEcho_cherry_studio(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return success("Echo: " + message)
}

func HandleGreet_cherry_studio(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return success("Hello, " + name + "!")
}