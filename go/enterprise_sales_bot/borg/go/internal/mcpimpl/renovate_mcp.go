package mcpimpl

import (
	"context"
	"fmt"
)

func HandleGetPackageRule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	packageName, _ :=getString(args, "packageName")
	if packageName == "" {
		return err("packageName is required")
}

	msg := fmt.Sprintf("Suggested rule for %s: enabled=true, automerge=true, labels=[renovate]", packageName)
	return ok(msg)
}