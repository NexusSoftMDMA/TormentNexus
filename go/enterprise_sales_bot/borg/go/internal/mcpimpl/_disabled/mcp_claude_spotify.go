package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleSearchTracks_mcp_claude_spotify(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	token := os.Getenv("SPOTIFY_ACCESS_TOKEN")
	if token == "" {
		return err("SPOTIFY_ACCESS_TOKEN not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/search?q="+q+"&type=track&limit=5", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
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
		return err("Spotify API error: " + string(body))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	tracks, found := result["tracks"].(map[string]interface{})
	if !found {
		return err("no tracks in response")
}

	items, found := tracks["items"].([]interface{})
	if !found {
		return err("no items in tracks")
}

	output, _ := json.MarshalIndent(items, "", "  ")
	return ok(string(output))
}

func HandleGetCurrentTrack(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("SPOTIFY_ACCESS_TOKEN")
	if token == "" {
		return err("SPOTIFY_ACCESS_TOKEN not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/me/player/currently-playing", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode == 204 {
		return ok("No track currently playing")
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("Spotify API error: " + string(body))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	output, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(output))
}