package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListRepositories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	workspace, _ :=getString(args, "workspace")
	if workspace == "" {
		return err("workspace is required")
}

	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s", workspace)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("repos: %v", result))
}

func HandleGetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.bitbucket.org/2.0/user")
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("user: %v", result))
}