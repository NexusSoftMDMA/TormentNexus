package mcpimpl

import (
	"context"
	"fmt"
	"net/http"
	"encoding/json"
	"io"
)

func HandleCheckDomain_mcp_dnstwist(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("domain is required")
	}
	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.dnstwist.io/api?domain=%s", domain))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
	}
	return success(fmt.Sprintf("Results: %+v", result))
}