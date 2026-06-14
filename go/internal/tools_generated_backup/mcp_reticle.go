package tools

import (
	"context"
	"net/http"
)

func HandleCaptureTraffic(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "filter")
	_ = filter
	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:9999/", nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to send request")
}

	resp.Body.Close()
	return success("traffic capture started")
}

func HandleViewProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	viewType, _ :=getString(args, "viewType")
	_ = viewType
	return ok("profile data retrieved")
}