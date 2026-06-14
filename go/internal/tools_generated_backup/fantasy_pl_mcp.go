package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPlayerStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	player, _ :=getString(args, "player_name")
	if player == "" {
		return err("player_name is required")
}

	url := fmt.Sprintf("https://api.fantasypl.com/players/%s", player)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch player stats: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Player stats: %v", data))
}

func HandleGetLeagueStandings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	league, _ :=getString(args, "league_id")
	if league == "" {
		return err("league_id is required")
}

	url := fmt.Sprintf("https://api.fantasypl.com/leagues/%s/standings", league)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch standings: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("League standings: %v", data))
}