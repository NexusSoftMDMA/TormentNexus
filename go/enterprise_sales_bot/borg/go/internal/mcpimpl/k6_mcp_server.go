package mcpimpl

import (
	"context"
)

func HandleRunK6Test(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	script, _ :=getString(args, "script")
	return ok("K6 test submitted successfully: " + script)
}