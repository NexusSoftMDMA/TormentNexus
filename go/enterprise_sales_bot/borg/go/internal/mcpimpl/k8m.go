package mcpimpl

import (
	"context"
)

func HandleGetClusterInfo_k8m(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("K8M cluster info retrieved")
}

func HandleListNodeStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	nodeName, _ :=getString(args, "nodeName")
	if nodeName == "" {
		return err("nodeName is required")
}

	return success("node status: Ready")
}