package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetRepo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	if owner == "" || repo == "" {
		return err("owner or repo missing")
}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch repo")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}

func HandleSearchRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query missing")
}

	url := fmt.Sprintf("https://api.github.com/search/repositories?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read")
}

	return ok(string(body))
}