package mcpimpl

import (
	"context"
	"net/http"
)

func HandleSayHello_janee(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello, " + name + "!")
}

func HandleEcho_janee(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message argument is required")
	}
	return ok("Echo: " + msg)
}