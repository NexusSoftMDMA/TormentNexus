package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchSongs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	term, _ :=getString(args, "term")
	if term == "" {
		return err("missing term")
}

	token, _ :=getString(args, "token")
	u := fmt.Sprintf("https://api.music.apple.com/v1/catalog/us/search?term=%s&types=songs", url.QueryEscape(term))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error")
}

	return ok(fmt.Sprintf("found %d songs", len(result)))
}

func HandleGetPlaylists(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	u := "https://api.music.apple.com/v1/me/library/playlists"
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error")
}

	return ok(fmt.Sprintf("playlists: %s", result))
}