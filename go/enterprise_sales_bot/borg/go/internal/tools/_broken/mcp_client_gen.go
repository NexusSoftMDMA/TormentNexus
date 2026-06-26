package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGenerateSDK(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	if serverURL == "" {
		return err("server_url is required")
	}
	outputDir, _ :=getString(args, "output_dir")
	if outputDir == "" {
		outputDir = "./sdk"
	}
	payload := map[string]interface{}{
		"server_url": serverURL,
		"output_dir": outputDir,
		"oauth":      getBool(args, "enable_oauth"),
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request"), e
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.mcp-gen.example/v1/generate", nil)
	if e != nil {
		return err("failed to create request"), e
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed"), e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("generation failed with status " + resp.Status)
	}
	return success("TypeScript SDK generated successfully in " + outputDir)
}

func HandleValidateConfig(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	config, _ :=getString(args, "config_json")
	if config == "" {
		return err("config_json is required")
	}
	var data map[string]interface{}
	e := json.Unmarshal([]byte(config), &data)
	if e != nil {
		return err("invalid JSON configuration"), e
	}
	found, _ := data["providers"].([]interface{})
	if !found {
		return err("missing providers array in config")
	}
	return ok("Configuration validated successfully")
}