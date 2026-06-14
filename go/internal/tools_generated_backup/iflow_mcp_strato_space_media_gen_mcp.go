package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleGenerateMedia(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("prompt is required")
}

	width, _ :=getInt(args, "width")
	if width == 0 {
		width = 1024
	}
	height, _ :=getInt(args, "height")
	if height == 0 {
		height = 1024
	}
	body := map[string]interface{}{
		"prompt": prompt,
		"width":  width,
		"height": height,
	}
	b, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.strato.space/v1/media/generate", strings.NewReader(string(b)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("api request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("api returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Media generated: %v", result["id"]))
}

func HandleGetMediaStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.strato.space/v1/media/%s", id), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("api request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("api returned status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	status, found := result["status"].(string)
	if !found {
		return err("status not found in response")
}

	return success(fmt.Sprintf("Media status: %s", status))
}