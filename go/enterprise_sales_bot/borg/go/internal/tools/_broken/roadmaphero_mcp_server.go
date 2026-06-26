package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetRoadmap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.roadmaphero.com/roadmaps/"+id, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch roadmap: %v", e))
}

	defer resp.Body.Close()
	var result struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	return ok(fmt.Sprintf("Roadmap: %s (ID: %s)", result.Name, result.ID))
}

func HandleListRoadmaps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.roadmaphero.com/roadmaps", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to list roadmaps: %v", e))
}

	defer resp.Body.Close()
	var result []struct {
		Name string `json:"name"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	names := []string{}
	for _, r := range result {
		names = append(names, r.Name)

	return ok(fmt.Sprintf("Roadmaps: %v", names))
}
}