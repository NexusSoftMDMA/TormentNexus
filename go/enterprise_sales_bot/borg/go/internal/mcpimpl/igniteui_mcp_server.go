package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGitHubRepo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repo")
	if repo == "" {
		repo = "IgniteUI/igniteui-cli"
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s", repo)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch repo info: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	result := fmt.Sprintf("Repo: %s\nStars: %v\nDescription: %v", repo, data["stargazers_count"], data["description"])
	return ok(result)
}

func HandleCLIScaffold(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	component, _ :=getString(args, "component")
	if component == "" {
		component = "grid"
	}
	result := fmt.Sprintf("Scaffolded Ignite UI component: %s. Run 'igniteui-cli new %s' to create.", component, component)
	return ok(result)
}