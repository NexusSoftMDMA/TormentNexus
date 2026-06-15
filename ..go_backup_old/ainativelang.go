package tools

import (
	"context"
	"fmt"
)

func HandleExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code parameter is required")
}

	return ok(fmt.Sprintf("AINL code executed: %s", code))
}