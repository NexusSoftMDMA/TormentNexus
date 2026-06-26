package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleListGames(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.romm.example/games"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch games: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok("games list retrieved")
}

func HandleGetGameInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	gameID, _ :=getString(args, "game_id")
	if gameID == "" {
		return err("game_id is required")
}

	url := "https://api.romm.example/games/" + gameID
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch game info: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok("game info retrieved")
}