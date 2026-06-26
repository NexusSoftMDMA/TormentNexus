package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetPullRequestChanges(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	org, _ :=getString(args, "organization")
	proj, _ :=getString(args, "project")
	repo, _ :=getString(args, "repositoryId")
	prID, _ :=getInt(args, "pullRequestId")
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/git/repositories/%s/pullrequests/%d/changes?api-version=7.1-preview.1", org, proj, repo, prID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth("", getString(args, "pat"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Changes []struct {
			Item struct {
				Path string `json:"path"`
			} `json:"item"`
			ChangeType string `json:"changeType"`
		} `json:"changes"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	summary := fmt.Sprintf("Found %d changes in PR %d", len(result.Changes), prID)
	return ok(summary)
}