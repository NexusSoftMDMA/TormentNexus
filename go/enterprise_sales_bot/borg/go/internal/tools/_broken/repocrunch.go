package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetRepoStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	if owner == "" || repo == "" {
		return err("owner and repo are required")
}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("GitHub API returned %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(e.Error())
}

	return success(fmt.Sprintf("Repo stats: %+v", data))
}

func HandleGetRepoContributors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	if owner == "" || repo == "" {
		return err("owner and repo are required")
}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contributors", owner, repo)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("GitHub API returned %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var data []map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(e.Error())
}

	names := []string{}
	for _, c := range data {
		if name, found := c["login"].(string); found {
			names = append(names, name)

	}
	return success(fmt.Sprintf("Contributors: %v", names))
}
}