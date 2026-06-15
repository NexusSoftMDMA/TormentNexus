//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "url")
	if baseURL == "" {
		baseURL = os.Getenv("ATLASSIAN_URL")
		if baseURL == "" {
			return err("ATLASSIAN_URL not set and no url arg")

	}
	token, _ :=getString(args, "token")
	if token == "" {
		token = os.Getenv("ATLASSIAN_TOKEN")

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/rest/api/3/project", baseURL), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse: " + e.Error())
}

	return ok(fmt.Sprintf("Projects: %+v", result))
}

}
}

func HandleGetIssue(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "url")
	if baseURL == "" {
		baseURL = os.Getenv("ATLASSIAN_URL")
		if baseURL == "" {
			return err("ATLASSIAN_URL not set")

	}
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key arg required")
}

	token, _ :=getString(args, "token")
	if token == "" {
		token = os.Getenv("ATLASSIAN_TOKEN")

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/rest/api/3/issue/%s", baseURL, key), nil)
	if e != nil {
		return err("request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("do: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("unmarshal: " + e.Error())
}

	return ok(fmt.Sprintf("Issue %s: %+v", key, result))
}
}
}
