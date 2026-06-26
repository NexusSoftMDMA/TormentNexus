package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPlayer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "player_id")
	url := fmt.Sprintf("https://www.balldontlie.io/api/v1/players/%d", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch player: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Player data: %v", data))
}

func HandleGetTeam_balldontlie_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "team_id")
	url := fmt.Sprintf("https://www.balldontlie.io/api/v1/teams/%d", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch team: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Team data: %v", data))
}