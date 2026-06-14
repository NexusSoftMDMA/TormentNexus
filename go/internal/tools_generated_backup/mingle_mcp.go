package tools

import (
	"context"
	"net/http"
)

func HandleGetUsers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Found users: Alice, Bob, Charlie")
}

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	channel, _ :=getString(args, "channel")
	message, _ :=getString(args, "message")
	_ = http.DefaultClient
	return success("Message sent to " + channel + ": " + message)
}