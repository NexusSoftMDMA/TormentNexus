//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query")
}

	token, _ :=getString(args, "token")
	if token == "" {
		return err("missing token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/search?q="+url.QueryEscape(query)+"&type=track&limit=5", nil)
	if e != nil {
		return err("request creation: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode: " + e.Error())
}

	data, e := json.Marshal(result)
	if e != nil {
		return err("marshal: " + e.Error())
	return ok(string(data))
}

func HandlePlay(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	uri, _ :=getString(args, "track_uri")
	if uri == "" {
		return err("missing track_uri")
}

	token, _ :=getString(args, "token")
	if token == "" {
		return err("missing token")
}

	body := `{"uris":["` + strings.ReplaceAll(uri, `"`, `\"`) + `"]}`")
	req, e := http.NewRequestWithContext(ctx, "PUT", "https://api.spotify.com/v1/me/player/play", strings.NewReader(body))
	if e != nil {
		return err("request creation: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err("status " + resp.Status)
	return ok("playback started")
}
}
}
}