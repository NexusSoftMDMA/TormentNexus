package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetScreenInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	display, _ :=getString(args, "display")
	if display == "" {
		display = "primary"
	}
	url := fmt.Sprintf("http://localhost:9090/api/screen?display=%s", display)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to get screen info: %v", e))
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	return success(fmt.Sprintf("Screen info: %v", data))
}

func HandleCaptureScreen(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	region, _ :=getString(args, "region")
	if region == "" {
		region = "full"
	}
	url := fmt.Sprintf("http://localhost:9090/api/capture?region=%s", region)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to capture screen: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	return success(fmt.Sprintf("Capture result: %v", result))
}