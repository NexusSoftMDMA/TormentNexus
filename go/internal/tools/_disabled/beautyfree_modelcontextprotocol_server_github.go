package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetUserRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username parameter is required")
}

	url := fmt.Sprintf("https://api.github.com/users/%s/repos", username)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var repos []interface{}
	if e = json.NewDecoder(resp.Body).Decode(&repos); e != nil {
		return err(e.Error())
}

	names := ""
	for _, r := range repos {
		repo, found := r.(map[string]interface{})
		if !found {
			continue
		}
		name, _ := repo["name"].(string)
		if names != "" {
			names += ", "
		}
		names += name
	}
	if names == "" {
		return ok("No repositories found for user " + username)
}

	return success(fmt.Sprintf("Repositories: %s", names))
}