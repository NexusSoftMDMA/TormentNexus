package tools

import (
	"context"
)

func HandleGetIntel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	msg := "Intel report for " + name + ": All clear."
	return ok(msg)
}