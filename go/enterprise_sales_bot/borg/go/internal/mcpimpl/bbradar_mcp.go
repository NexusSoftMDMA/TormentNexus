package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleRunScan_bbradar_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	if target == "" {
		return err("target is required")
}

	apiKey, _ :=getString(args, "api_key")
	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.bbradar.io/pro/scan?target=%s&api_key=%s", target, apiKey))
	if e != nil {
		return err("failed to call API: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(result)
}

func HandleGetResult_bbradar_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	scanID, _ :=getString(args, "scan_id")
	if scanID == "" {
		return err("scan_id is required")
}

	apiKey, _ :=getString(args, "api_key")
	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.bbradar.io/pro/result?scan_id=%s&api_key=%s", scanID, apiKey))
	if e != nil {
		return err("failed to call API: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(result)
}