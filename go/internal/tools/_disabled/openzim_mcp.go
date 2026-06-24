package tools

import "context"

func HandleFetchZimPage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	content := "Page content for: " + path
	return ok(content)
}