package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetAlerts_netskope_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	token, _ :=getString(args, "token")
	if host == "" || token == "" {
		return err("missing required arguments: host and token")
}

	url := "https://" + host + "/api/v1/alerts"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer " + token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(result)
}