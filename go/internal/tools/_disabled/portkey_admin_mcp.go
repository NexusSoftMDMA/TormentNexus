package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListApiKeys(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workspaceID, _ :=getString(args, "workspace_id")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 50
	}
	url := fmt.Sprintf("https://api.portkey.ai/v1/admin/api-keys?workspace_id=%s&limit=%d", workspaceID, limit)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("x-api-key", getString(args, "api_key"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success(string(body))
}

func HandleCreateApiKey(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	workspaceID, _ :=getString(args, "workspace_id")
	payload := map[string]interface{}{
		"name":         name,
		"workspace_id": workspaceID,
	}
	jsonData, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	url := "https://api.portkey.ai/v1/admin/api-keys"
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", getString(args, "api_key"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	return success(string(body))
}