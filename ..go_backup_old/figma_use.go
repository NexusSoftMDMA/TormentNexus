package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetFigmaFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	fileKey, _ :=getString(args, "file_key")
	if fileKey == "" {
		return err("file_key is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.figma.com/v1/files/%s", fileKey), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("X-Figma-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	return ok(fmt.Sprintf("Figma file data: %s", string(body)))
}

func HandleGetFigmaComponents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	fileKey, _ :=getString(args, "file_key")
	if fileKey == "" {
		return err("file_key is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.figma.com/v1/files/%s/components", fileKey), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("X-Figma-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	return ok(fmt.Sprintf("Figma components data: %s", string(body)))
}