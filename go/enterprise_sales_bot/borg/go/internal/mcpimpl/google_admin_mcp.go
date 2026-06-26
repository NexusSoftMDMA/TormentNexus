package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetUser_google_admin_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	userKey, _ :=getString(args, "userKey")
	token, _ :=getString(args, "accessToken")
	if userKey == "" || token == "" {
		return err("userKey and accessToken required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://admin.googleapis.com/admin/directory/v1/users/%s", userKey), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("parse error: " + e.Error())
}

	return ok("user: " + fmt.Sprint(result))
}

func HandleListUsers_google_admin_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	token, _ :=getString(args, "accessToken")
	if token == "" {
		return err("accessToken required")
}

	url := "https://admin.googleapis.com/admin/directory/v1/users"
	if query != "" {
		url += "?query=" + query
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("parse error: " + e.Error())
}

	return ok("users: " + fmt.Sprint(result))
}