package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := fmt.Sprintf("https://api.github.com/users/%s", username)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch user: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("failed to parse json: %v", e))
}

	login, found := data["login"].(string)
	if !found {
		return err("login field not found")
}

	return ok(fmt.Sprintf("User login: %s", login))
}

func HandleListRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := fmt.Sprintf("https://api.github.com/users/%s/repos", username)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch repos: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var repos []map[string]interface{}
	if e := json.Unmarshal(body, &repos); e != nil {
		return err(fmt.Sprintf("failed to parse json: %v", e))
}

	names := []string{}
	for _, repo := range repos {
		name, found := repo["name"].(string)
		if found {
			names = append(names, name)

	}
	return ok(fmt.Sprintf("Repos: %v", names))
}
}