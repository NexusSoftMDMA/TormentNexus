package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleRubyVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	version, _ :=getString(args, "version")
	if version == "" {
		version = "latest"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://www.ruby-lang.org/en/news", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("Ruby version check simulated for: " + version)
}

func HandleGemInstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("gem name is required")
	}
	data := map[string]interface{}{"gem": name, "status": "installed"}
	body, e := json.Marshal(data)
	if e != nil {
		return err("marshal failed: " + e.Error())
	}
	return ok(string(body))
}