package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandlePhpcodearcheology(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("missing code argument")
	}
	reqBody, e := json.Marshal(map[string]string{"code": code})
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.phpcodearcheology.com/analyze", strings.NewReader(string(reqBody)))
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
		return err("failed to decode response")
	}
	return success("Analysis complete")
}

func HandlePhpcodearcheologyInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Phpcodearcheology MCP server ready")
}// touch 1781132138
