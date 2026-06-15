//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	org, _ :=getString(args, "organization")
	pat, _ :=getString(args, "pat")
	if org == "" {
		org = "myorg"
	}
	if pat == "" {
		pat = "mypat"
	}
	url := fmt.Sprintf("https://dev.azure.com/%s/_apis/projects?api-version=7.0", org)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth("", pat)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}
