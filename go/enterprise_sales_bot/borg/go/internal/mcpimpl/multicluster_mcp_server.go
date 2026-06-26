package mcpimpl

import (
	"context"
)

func HandleListClusters_multicluster_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return ok("Available clusters: cluster-a, cluster-b, cluster-c")
}

	return ok("Cluster " + name + " is active")
}

func HandleGetCluster_multicluster_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("cluster name is required")
}

	return ok("Cluster " + name + " status: healthy")
}