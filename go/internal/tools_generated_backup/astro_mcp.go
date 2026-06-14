package tools

import "context"

func HandleAstroInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Astro MCP server ready. Use 'list-components' or 'analyze-project'.")
}

func HandleAstroListComponents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dir, _ :=getString(args, "directory")
	if dir == "" {
		dir = "./src"
	}
	return ok("Components in " + dir)
}