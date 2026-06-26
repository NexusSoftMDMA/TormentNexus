package mcpimpl

import "context"

func HandleNavigate_puppeteer_mcp_claude(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	return success("navigated to " + url)
}

func HandleScreenshot_puppeteer_mcp_claude(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	return success("screenshot saved to " + path)
}