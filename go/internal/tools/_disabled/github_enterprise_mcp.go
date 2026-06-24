package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleListRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	if owner == "" {
		return err("owner is required")
}

	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos", owner)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var repos []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&repos); e != nil {
		return err("decode failed: " + e.Error())
}

	names := []string{}
	for _, r := range repos {
		name, found := r["full_name"].(string)
		if found {
			names = append(names, name)

	}
	return ok(fmt.Sprintf("Found %d repos: %s", len(names), strings.Join(names, ", ")))
}

}

func HandleCreateRepo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	name, _ :=getString(args, "name")
	if owner == "" || name == "" {
		return err("owner and name are required")
}

	body := map[string]interface{}{"name": name, "auto_init": true}
	b, e := json.Marshal(body)
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos", owner)
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(b)))
	if e != nil {
		return err("create request failed: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("POST failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok("Repository created successfully")
}