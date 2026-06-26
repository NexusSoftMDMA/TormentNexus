package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleAnalyzeSnapshot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "snapshot_path")
	if path == "" {
		return err("snapshot_path is required")
}

	body, _ := json.Marshal(map[string]string{"path": path})
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.memlab.dev/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}

func HandleListLeaks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "snapshot_id")
	if id == "" {
		return err("snapshot_id is required")
}

	url := fmt.Sprintf("https://api.memlab.dev/leaks/%s", id)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return ok(string(data))
}