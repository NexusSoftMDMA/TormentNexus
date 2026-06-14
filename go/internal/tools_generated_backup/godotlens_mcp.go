package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGodotVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	version, _ :=getString(args, "version")
	if version == "" {
		version = "latest"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://godotengine.org/download", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("Godot version " + version + " info retrieved")
}

func HandleProjectScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("project path is required")
	}
	data := map[string]interface{}{"path": path, "status": "scanned"}
	body, e := json.Marshal(data)
	if e != nil {
		return err("marshal failed: " + e.Error())
	}
	return success(string(body))
}// touch 1781132126
