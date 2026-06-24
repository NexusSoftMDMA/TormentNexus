package tools

import (
	"context"
)

func HandleGetPharaoh(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	return ok("Pharaoh: " + name)
}

func HandleGetFact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Ancient Egypt was one of the greatest civilizations in history.")
}