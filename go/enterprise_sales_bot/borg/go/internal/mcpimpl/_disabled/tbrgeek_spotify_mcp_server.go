package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandlePlay_tbrgeek_spotify_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("SPOTIFY_ACCESS_TOKEN")
	if token == "" {
		return err("missing SPOTIFY_ACCESS_TOKEN")
}

	req, e := http.NewRequestWithContext(ctx, "PUT", "https://api.spotify.com/v1/me/player/play", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("play request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return err(fmt.Sprintf("play returned status %d", resp.StatusCode))
}

	return ok("playback started")
}

func HandlePause(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("SPOTIFY_ACCESS_TOKEN")
	if token == "" {
		return err("missing SPOTIFY_ACCESS_TOKEN")
}

	req, e := http.NewRequestWithContext(ctx, "PUT", "https://api.spotify.com/v1/me/player/pause", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("pause request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return err(fmt.Sprintf("pause returned status %d", resp.StatusCode))
}

	return ok("playback paused")
}