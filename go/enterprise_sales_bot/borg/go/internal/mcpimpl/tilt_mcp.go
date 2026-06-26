package mcpimpl

import (
	"context"
)

func HandleListResources_tilt_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resourceType, _ :=getString(args, "type")
	if resourceType == "" {
		resourceType = "all"
	}
	return ok(`{"resources": ["service1","service2"],"type":"` + resourceType + `"}`)
}

func HandleGetStatus_tilt_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resource, _ :=getString(args, "resource")
	if resource == "" {
		return err("resource name is required")
}

	return ok(`{"status":"running","name":"` + resource + `"}`)
}