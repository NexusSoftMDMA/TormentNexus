package tools

import (
    "context"
)

func HandleAutoConfigure(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    browser, _ :=getString(args, "browser")
    if browser == "" {
        browser = "chromium"
    }
    return ok("Pi Agent auto-configured for " + browser + " browser MCP tools")
}