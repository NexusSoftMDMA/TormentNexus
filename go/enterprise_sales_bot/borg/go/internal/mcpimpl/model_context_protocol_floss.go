package mcpimpl

import "context"

func HandleFlossInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoName, _ :=getString(args, "repository")
	if repoName == "" {
		repoName = "modelcontextprotocol/floss"
	}
	return ok("Repository: " + repoName + " – dedicated to the MCP talk at the Free/Libre Open Source event")
}