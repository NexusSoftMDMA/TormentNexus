package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListActors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	actors := []map[string]interface{}{
		{"name": "Tom Hanks", "birth": 1956},
		{"name": "Meryl Streep", "birth": 1949},
	}
	if len(actors) > limit {
		actors = actors[:limit]
	}
	return ok(fmt.Sprintf("Found %d actors: %v", len(actors), actors))
}

func HandleGetActor(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("actor name is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com/actors/"+name, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	return success(fmt.Sprintf("Actor info: %v", result))
}