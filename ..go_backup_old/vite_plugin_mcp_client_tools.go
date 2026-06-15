package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleViteBuild(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	configPath, _ :=getString(args, "config")
	if configPath == "" {
		configPath = "vite.config.js"
	}
	reqBody, e := json.Marshal(map[string]string{"config": configPath, "action": "build"})
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:5173/__mcp/build", nil)
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("Vite build triggered for " + configPath)
}

func HandleVitePreview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	port, _ :=getInt(args, "port")
	if port == 0 {
		port = 4173
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:"+string(rune(port)), nil)
	if e != nil {
		return err("invalid port")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return ok("Preview server not running on port " + string(rune(port)))
	}
	defer resp.Body.Close()
	return success("Preview server is active on port " + string(rune(port)))
}