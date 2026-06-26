package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func HandleAddTracks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	playlistID, _ :=getString(args, "playlist_id")
	trackURIs := strings.Split(getString(args, "track_uris"), ",")
	if playlistID == "" || len(trackURIs) == 0 || trackURIs[0] == "" {
		return err("playlist_id and track_uris are required")
}

	token := os.Getenv("SPOTIFY_TOKEN")
	if token == "" {
		return err("SPOTIFY_TOKEN environment variable not set")
}

	body := map[string]interface{}{"uris": trackURIs}
	b, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal body: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID), strings.NewReader(string(b)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return err(fmt.Sprintf("Spotify API returned status %d", resp.StatusCode))
}

	return success("Tracks added successfully")
}

func HandleRemoveTracks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	playlistID, _ :=getString(args, "playlist_id")
	trackURIs := strings.Split(getString(args, "track_uris"), ",")
	if playlistID == "" || len(trackURIs) == 0 || trackURIs[0] == "" {
		return err("playlist_id and track_uris are required")
}

	token := os.Getenv("SPOTIFY_TOKEN")
	if token == "" {
		return err("SPOTIFY_TOKEN environment variable not set")
}

	body := map[string]interface{}{"uris": trackURIs}
	b, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal body: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID), strings.NewReader(string(b)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Spotify API returned status %d", resp.StatusCode))
}

	return success("Tracks removed successfully")
}