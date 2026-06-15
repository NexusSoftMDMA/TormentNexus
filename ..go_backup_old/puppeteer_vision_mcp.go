package tools

import (
	"context"
	"encoding/base64"
	"fmt"
)

func HandleTakeScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url parameter is required")
}

	// Mock screenshot data
	screenshot := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
	data := base64.StdEncoding.EncodeToString([]byte(screenshot))
	return ok(fmt.Sprintf("Screenshot of %s captured: data:image/png;base64,%s", url, data))
}

func HandleClickElement(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	selector, _ :=getString(args, "selector")
	url, _ :=getString(args, "url")
	if selector == "" || url == "" {
		return err("selector and url parameters are required")
}

	return success(fmt.Sprintf("Clicked element '%s' on page %s", selector, url))
}