package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListRepositories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	owner, _ :=getString(args, "owner")
	url := fmt.Sprintf("%s/api/v1/users/%s/repos", host, owner)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch repos: " + e.Error())
}

	defer resp.Body.Close()
	var repos []struct{ Name string }
	if e := json.NewDecoder(resp.Body).Decode(&repos); e != nil {
		return err("failed to decode repos: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d repos", len(repos)))
}

func HandleGetIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	num, _ :=getInt(args, "issue_number")
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/issues/%d", host, owner, repo, num)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch issue: " + e.Error())
}

	defer resp.Body.Close()
	var issue map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&issue); e != nil {
		return err("failed to decode issue: " + e.Error())
}

	title, found := issue["title"].(string)
	if !found {
		return err("issue has no title")
}

	return success(fmt.Sprintf("Issue #%d: %s", num, title))
}