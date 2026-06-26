package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListEpisodes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.podbean.com/v1/episodes")
	if e != nil {
		return err("failed to fetch episodes: " + e.Error())
}

	defer resp.Body.Close()
	var data interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	return success(fmt.Sprintf("episodes: %v", data))
}

func HandleGetEpisode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "episode_id")
	if id == "" {
		return err("episode_id is required")
}

	url := "https://api.podbean.com/v1/episodes/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch episode: " + e.Error())
}

	defer resp.Body.Close()
	var data interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	return success(fmt.Sprintf("episode: %v", data))
}