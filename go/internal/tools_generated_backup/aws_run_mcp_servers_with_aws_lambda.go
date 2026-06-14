package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func HandleRunMCP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	endpoint, _ :=getString(args, "endpoint")
	serverName, _ :=getString(args, "serverName")
	action, _ :=getString(args, "action")

	payload, e := json.Marshal(map[string]string{
		"server": serverName,
		"action": action,
	})
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(payload)))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 {
		return err("server returned " + resp.Status + ": " + string(body))
}

	return ok("MCP action executed successfully: " + string(body))
}