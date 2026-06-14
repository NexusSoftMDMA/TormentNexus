package tools

import (
	"context"
)

func HandleLaunchBrowser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	headless, _ :=getBool(args, "headless")
	_ = headless
	return ok("Browser launched successfully")
}

func HandleCloseBrowser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Browser closed")
}