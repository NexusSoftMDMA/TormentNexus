package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetUserRepos_developer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := "https://api.github.com/users/" + username + "/repos"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch repos: " + e.Error())
}

	defer resp.Body.Close()
	var repos []struct{ Name string `json:"name"` }
	if e := json.NewDecoder(resp.Body).Decode(&repos); e != nil {
		return err("failed to decode: " + e.Error())
}

	names := make([]string, 0, len(repos))
	for _, r := range repos {
		names = append(names, r.Name)

	return ok("Repos: " + join(names))
}

}

func HandleGetUserInfo_developer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := "https://api.github.com/users/" + username
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch user: " + e.Error())
}

	defer resp.Body.Close()
	var user struct{ Login, Name string }
	if e := json.NewDecoder(resp.Body).Decode(&user); e != nil {
		return err("failed to decode: " + e.Error())
}

	return ok("User: " + user.Login + " (" + user.Name + ")")
}

func join_developer(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	res := strs[0]
	for _, s := range strs[1:] {
		res += ", " + s
	}
	return res
}