package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetRepository(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Failed to fetch repo: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("Failed to parse response: " + e.Error())
}

	name, found := data["full_name"].(string)
	if !found {
		return err("Missing full_name in response")
}

	stars, _ := data["stargazers_count"].(float64)
	msg := fmt.Sprintf("Repository: %s, Stars: %d", name, int(stars))
	return ok(msg)
}