package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleChat_spring_ai_mcp_client(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	prompt, _ :=getString(args, "prompt")
	if prompt == "" {
		return err("missing prompt argument")
	}
	payload := map[string]interface{}{"input": prompt}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/chat", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed")
	}
	return success(fmt.Sprintf("%v", result["response"]))
}

func HandleStatus_spring_ai_mcp_client(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/health", nil)
	if e != nil {
		return err("request creation failed")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("health check failed")
	}
	defer resp.Body.Close()
	return ok("Spring AI MCP Client is active")
}