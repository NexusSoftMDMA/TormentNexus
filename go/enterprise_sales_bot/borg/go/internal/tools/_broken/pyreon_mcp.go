package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandlePyreonAssist(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("missing prompt argument")
	}
	body, e := json.Marshal(map[string]string{"query": prompt})
	if e != nil {
		return err("failed to marshal request"), e
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.pyreon.ai/v1/assist", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request"), e
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed"), e
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed"), e
	}
	return success(result["answer"].(string))
}

func HandlePyreonStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.pyreon.ai/v1/status", nil)
	if e != nil {
		return err("request creation failed"), e
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("status check failed"), e
	}
	defer resp.Body.Close()
	return success("Pyreon MCP server is operational")
}