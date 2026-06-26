package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type repo struct {
	Name string `json:"name"`
	URL  string `json:"html_url"`
}

func HandleListRepos_bifrost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		q = "mcp-client"
	}
	u := "https://api.github.com/search/repositories?q=" + q
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to call GitHub: " + e.Error())
}

	defer resp.Body.Close()
	var data struct {
		Items []repo `json:"items"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("json decode: " + e.Error())
}

	if len(data.Items) == 0 {
		return ok("no repositories found")
}

	var s string
	for _, r := range data.Items {
		s += fmt.Sprintf("- %s (%s)\n", r.Name, r.URL)

	return ok(s)
}

}

func HandleGetRepo_bifrost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repoName, _ :=getString(args, "repo")
	u := "https://api.github.com/repos/" + owner + "/" + repoName
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to call GitHub: " + e.Error())
}

	defer resp.Body.Close()
	var r repo
	if e := json.NewDecoder(resp.Body).Decode(&r); e != nil {
		return err("json decode: " + e.Error())
}

	return ok(fmt.Sprintf("%s (%s)", r.Name, r.URL))
}