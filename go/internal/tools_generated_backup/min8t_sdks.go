package tools

import (
	"context"
)

func HandleGetSdk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return success("SDK: " + name)
}

func HandleListSdks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("List of SDKs: Min8T Core, Min8T Analytics, Min8T AI")
}