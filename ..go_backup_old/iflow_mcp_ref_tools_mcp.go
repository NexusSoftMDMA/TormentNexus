package tools

import (
	"context"
)

func HandleGetRef(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ref, _ :=getString(args, "ref")
	if ref == "" {
		return err("ref is required")
	}
	return ok("Ref: " + ref)
}

func HandleValidateRef(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ref, _ :=getString(args, "ref")
	if ref == "" {
		return err("ref is required")
	}
	return success("ref " + ref + " is valid")
}