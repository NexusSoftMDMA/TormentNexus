package tools

import (
	"context"
)

func HandleGetAccessibilityInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	return ok("Accessibility info for " + url + " is not implemented")
}

func HandleCheckContrast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fg, _ :=getString(args, "foreground")
	bg, _ :=getString(args, "background")
	if fg == "" || bg == "" {
		return err("foreground and background are required")
}

	return ok("Contrast ratio not computed (dummy)")
}