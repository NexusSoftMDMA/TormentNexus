package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGitlabProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token required")
}

	url := fmt.Sprintf("https://gitlab.com/api/v4/projects?private_token=%s", token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}

func HandleZentaoTasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	api, _ :=getString(args, "api_url")
	token, _ :=getString(args, "token")
	if api == "" || token == "" {
		return err("api_url and token required")
}

	url := fmt.Sprintf("%s/api.php/v1/tasks?token=%s", api, token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}