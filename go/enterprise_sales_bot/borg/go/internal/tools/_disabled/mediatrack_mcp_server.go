package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetTrack(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing track id")
}

	url := "https://api.mediatrack.dev/tracks/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Track: %v", result))
}

func HandleSearchTracks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing search query")
}

	url := "https://api.mediatrack.dev/search?q=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var results []map[string]interface{}
	e = json.Unmarshal(body, &results)
	if e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Found %d tracks", len(results)))
}