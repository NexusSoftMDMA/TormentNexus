package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetMusic(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := fmt.Sprintf("https://api.near-music-suite.com/music/%s", id)
	e := doGet(ctx, url)
	if e != nil {
		return err(e.Error())
	}
	return ok("music retrieved")
}

func HandleListMusic(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	url := fmt.Sprintf("https://api.near-music-suite.com/music?limit=%d", limit)
	e := doGet(ctx, url)
	if e != nil {
		return err(e.Error())
	}
	return ok("music list retrieved")
}

func doGet(ctx context.Context, url string) error {
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return e
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
}

	return nil
}