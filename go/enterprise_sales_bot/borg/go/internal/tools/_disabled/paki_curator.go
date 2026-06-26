package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCurate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repo")
	if repo == "" {
		return err("repo argument is required")
}

	resp, e := http.DefaultClient.Get("https://api.github.com/repos/" + repo)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch repo: %v", e))
}

	defer resp.Body.Close()

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	name, found := data["full_name"].(string)
	if !found {
		return err("invalid repo data")
}

	desc, _ := data["description"].(string)
	stars := int(data["stargazers_count"].(float64))

	msg := fmt.Sprintf("Repo: %s\nDescription: %s\nStars: %d", name, desc, stars)
	return ok(msg)
}