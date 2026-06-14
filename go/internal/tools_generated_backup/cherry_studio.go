package tools

import "context"

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	return success("Echo: " + message)
}

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return success("Hello, " + name + "!")
}