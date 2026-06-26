package mcpimpl

import (
	"context"
	"fmt"
)

func HandleSearchDocs_garmin_documentation_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok(fmt.Sprintf("Search results for '%s': See https://developer.garmin.com", query))
}