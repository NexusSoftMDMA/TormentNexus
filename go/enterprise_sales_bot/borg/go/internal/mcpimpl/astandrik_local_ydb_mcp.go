package mcpimpl

import (
	"context"
)

func HandleListDeployments_astandrik_local_ydb_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("local-ydb deployments: [test-1, test-2]")
}

func HandleStartDeployment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("deployment name is required")
}

	return ok("started deployment " + name)
}