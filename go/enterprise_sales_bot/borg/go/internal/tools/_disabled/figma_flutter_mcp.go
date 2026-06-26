package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetFileStyles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey, _ :=getString(args, "file_key")
	token, _ :=getString(args, "access_token")
	if fileKey == "" || token == "" {
		return err("missing required parameters: file_key and access_token")
}

	url := fmt.Sprintf("https://api.figma.com/v1/files/%s/styles", fileKey)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("X-Figma-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("figma API returned %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	data, found := result["meta"]
	if !found {
		return err("unexpected response structure")
}

	return ok(fmt.Sprintf("styles: %v", data))
}

func HandleGetFileComponents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey, _ :=getString(args, "file_key")
	token, _ :=getString(args, "access_token")
	if fileKey == "" || token == "" {
		return err("missing required parameters: file_key and access_token")
}

	url := fmt.Sprintf("https://api.figma.com/v1/files/%s/components", fileKey)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("X-Figma-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("figma API returned %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	data, found := result["meta"]
	if !found {
		return err("unexpected response structure")
}

	return ok(fmt.Sprintf("components: %v", data))
}