package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListFiles_putio_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("PUTIO_TOKEN")
	if token == "" {
		return err("PUTIO_TOKEN not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.put.io/v2/files/list", nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse JSON: %v", e))
}

	return ok(fmt.Sprintf("list files result: %v", result))
}

func HandleGetFile_putio_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileID, _ :=getString(args, "file_id")
	if fileID == "" {
		return err("file_id is required")
}

	token := os.Getenv("PUTIO_TOKEN")
	if token == "" {
		return err("PUTIO_TOKEN not set")
}

	url := fmt.Sprintf("https://api.put.io/v2/files/%s", fileID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse JSON: %v", e))
}

	return ok(fmt.Sprintf("file info: %v", result))
}