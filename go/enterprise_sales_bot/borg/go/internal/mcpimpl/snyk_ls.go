package mcpimpl

import (
	"context"
)

func HandleListProjects_snyk_ls(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	org, _ :=getString(args, "org")
	if org == "" {
		return err("org is required")
}

	return ok("Listing Snyk projects for org: " + org)
}