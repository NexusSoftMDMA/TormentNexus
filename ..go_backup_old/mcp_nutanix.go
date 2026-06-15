package tools

import (
	"context"
	"fmt"
)

func HandleListVMs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cluster, _ :=getString(args, "cluster")
	if cluster == "" {
		return success("Listing all VMs")
}

	return success(fmt.Sprintf("Listing VMs in cluster: %s", cluster))
}

func HandleGetVM(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("VM name is required")
}

	return success(fmt.Sprintf("Getting VM: %s", name))
}