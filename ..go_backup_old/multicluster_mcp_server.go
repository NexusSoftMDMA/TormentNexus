package tools

import (
	"context"
)

func HandleListClusters(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return ok("Available clusters: cluster-a, cluster-b, cluster-c")
}

	return ok("Cluster " + name + " is active")
}

func HandleGetCluster(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("cluster name is required")
}

	return ok("Cluster " + name + " status: healthy")
}