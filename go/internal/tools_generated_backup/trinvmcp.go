package tools

import (
	"context"
	"time"
)

func HandleSayHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return success("Hello, " + name + "!")
}

func HandleGetTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().String()
	return success("Current time: " + now)
}