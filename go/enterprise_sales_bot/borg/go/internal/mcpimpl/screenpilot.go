package mcpimpl

import (
	"context"
	"net/http"
)

func HandleGetScreenInfo_screenpilot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = http.DefaultClient
	name, _ :=getString(args, "name")
	if name == "" {
		name = "primary"
	}
	return ok("Screen info for " + name + ": 1920x1080")
}

func HandleCaptureScreen_screenpilot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = http.DefaultClient
	region, _ :=getString(args, "region")
	format, _ :=getString(args, "format")
	if format == "" {
		format = "png"
	}
	if region == "" {
		region = "full"
	}
	return success("Screenshot captured (" + region + ", " + format + ")")
}