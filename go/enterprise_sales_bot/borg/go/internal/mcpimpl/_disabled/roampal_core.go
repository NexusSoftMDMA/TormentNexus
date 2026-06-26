package mcpimpl

import (
	"context"
	"errors"
	"fmt"
)

func HandleRoampalInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name argument is required")
}

	msg := fmt.Sprintf("Roampal Core says hello to %s", name)
	return ok(msg)
}

func HandleRoampalGreeting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	greeting, _ :=getString(args, "greeting")
	if greeting == "" {
		return err("greeting argument is required")
}

	return success(greeting)
}