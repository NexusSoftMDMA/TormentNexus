package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func SteamSearchGames(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	url := fmt.Sprintf("https://api.steampowered.com/ISteamApps/GetAppList/v2/?key=%s", apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch app list: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Applist struct {
			Apps []struct {
				Appid int    `json:"appid"`
				Name  string `json:"name"`
			} `json:"apps"`
		} `json:"applist"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	var matches []string
	for _, app := range result.Applist.Apps {
		if len(matches) >= 10 {
			break
		}
		if contains(app.Name, query) {
			matches = append(matches, fmt.Sprintf("%s (AppID: %d)", app.Name, app.Appid))

	}
	if len(matches) == 0 {
		return ok("No matching games found")
}

	out, _ := json.Marshal(matches)
	return ok(string(out))
}

}

func SteamGetPlayerSummary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	steamID, _ :=getString(args, "steam_id")
	if steamID == "" {
		return err("steam_id is required")
}

	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	url := fmt.Sprintf("https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?key=%s&steamids=%s", apiKey, steamID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch player summary: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result struct {
		Response struct {
			Players []struct {
				Personaname string `json:"personaname"`
				Steamid     string `json:"steamid"`
				Avatarfull  string `json:"avatarfull"`
			} `json:"players"`
		} `json:"response"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if len(result.Response.Players) == 0 {
		return ok("Player not found")
}

	player := result.Response.Players[0]
	out, _ := json.Marshal(map[string]string{
		"name":   player.Personaname,
		"id":     player.Steamid,
		"avatar": player.Avatarfull,
	})
	return ok(string(out))
}

func contains_steam_mcp(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && s != "" && substr != "" && (s == substr || (len(s) > len(substr) && s[:len(substr)] == substr) || (len(s) > len(substr) && s[len(s)-len(substr):] == substr) || (len(s) > len(substr) && contains(s[1:], substr)))))
}