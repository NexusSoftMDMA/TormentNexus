package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleScreenshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deviceID, _ :=getString(args, "device_id")
	url := fmt.Sprintf("http://localhost:8080/screenshot?device_id=%s", deviceID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
	}
	_, _ = io.ReadAll(resp.Body)
	return ok("Screenshot captured successfully")
}

func HandleTap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	x, _ :=getInt(args, "x")
	y, _ :=getInt(args, "y")
	url := fmt.Sprintf("http://localhost:8080/tap?x=%d&y=%d", x, y)
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("tap request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("tap failed with status: %d", resp.StatusCode))
	}
	return ok(fmt.Sprintf("Tapped at (%d, %d)", x, y))
}