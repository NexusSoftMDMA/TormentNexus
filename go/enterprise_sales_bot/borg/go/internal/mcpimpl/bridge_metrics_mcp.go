package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetBridgeMetrics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	bridgeID, _ :=getString(args, "bridge_id")
	if bridgeID == "" {
		return err("bridge_id is required")
}

	url := fmt.Sprintf("https://api.example.com/bridge/%s/metrics", bridgeID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	output, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(output))
}