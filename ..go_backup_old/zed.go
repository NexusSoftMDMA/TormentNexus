package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleZedStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://zed.mcp.server/status")
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	var data map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err("failed to decode response")
}

	status, found := data["status"].(string)
	if !found {
		return err("status field missing")
}

	return ok(fmt.Sprintf("Zed server status: %s", status))
}

func HandleZedPlayers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://zed.mcp.server/players")
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	var data map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err("failed to decode response")
}

	players, found := data["players"].([]interface{})
	if !found {
		return err("players field missing")
}

	return ok(fmt.Sprintf("Online players: %d", len(players)))
}// touch 1781132144
