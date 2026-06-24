package tools

import (
	"context"
)

func HandlePointElement(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	selector, _ :=getString(args, "selector")
	if selector == "" {
		return err("selector parameter is required")
}

	return ok("Pointing to element: " + selector)
}

func HandleGetElementInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	selector, _ :=getString(args, "selector")
	if selector == "" {
		return err("selector parameter is required")
}

	return success("Element info for selector: " + selector)
}