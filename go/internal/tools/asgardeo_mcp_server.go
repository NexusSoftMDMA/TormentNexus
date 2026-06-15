//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
)

func HandleGreetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return ok("Hello, world!")
}

	return ok(fmt.Sprintf("Hello, %s!", name))
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok("You said: " + msg)
}
