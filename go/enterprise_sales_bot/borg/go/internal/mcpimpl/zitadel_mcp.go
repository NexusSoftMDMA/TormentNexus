package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListUsers_zitadel_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseUrl, _ :=getString(args, "baseUrl")
	if baseUrl == "" {
		baseUrl = "https://api.zitadel.com"
	}
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseUrl+"/v2/users", nil)
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
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success(fmt.Sprintf("Users: %v", result))
}

func HandleGetUser_zitadel_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseUrl, _ :=getString(args, "baseUrl")
	if baseUrl == "" {
		baseUrl = "https://api.zitadel.com"
	}
	token, _ :=getString(args, "token")
	userId, _ :=getString(args, "userId")
	if token == "" || userId == "" {
		return err("token and userId are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseUrl+"/v2/users/"+userId, nil)
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
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success(fmt.Sprintf("User: %v", result))
}