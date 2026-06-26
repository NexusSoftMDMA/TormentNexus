package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleFigmaGetFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey, _ :=getString(args, "file_key")
	if fileKey == "" {
		return err("file_key is required")
}

	token, _ :=getString(args, "access_token")
	if token == "" {
		return err("access_token is required")
}

	url := fmt.Sprintf("https://api.figma.com/v1/files/%s", fileKey)
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
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Figma API returned status %d: %s", resp.StatusCode, string(body)))
}

	return success(string(body))
}

func HandleFigmaGetNode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileKey, _ :=getString(args, "file_key")
	if fileKey == "" {
		return err("file_key is required")
}

	nodeId, _ :=getString(args, "node_id")
	if nodeId == "" {
		return err("node_id is required")
}

	token, _ :=getString(args, "access_token")
	if token == "" {
		return err("access_token is required")
}

	url := fmt.Sprintf("https://api.figma.com/v1/files/%s/nodes?ids=%s", fileKey, nodeId)
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
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Figma API returned status %d: %s", resp.StatusCode, string(body)))
}

	return success(string(body))
}