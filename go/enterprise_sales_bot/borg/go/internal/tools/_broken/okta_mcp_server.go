package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListUsers_okta_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		domain = os.Getenv("OKTA_DOMAIN")

	if domain == "" {
		return err("missing domain")
}

	token, _ :=getString(args, "token")
	if token == "" {
		token = os.Getenv("OKTA_API_TOKEN")

	if token == "" {
		return err("missing token")
}

	q, _ :=getString(args, "query")
	url := fmt.Sprintf("https://%s/api/v1/users?limit=100", domain)
	if q != "" {
		url += "&q=" + q
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "SSWS "+token)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var users []interface{}
	if e := json.Unmarshal(body, &users); e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Found %d users", len(users)))
}

}
}

func HandleGetUser_okta_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		domain = os.Getenv("OKTA_DOMAIN")

	if domain == "" {
		return err("missing domain")
}

	token, _ :=getString(args, "token")
	if token == "" {
		token = os.Getenv("OKTA_API_TOKEN")

	if token == "" {
		return err("missing token")
}

	userId, _ :=getString(args, "userId")
	if userId == "" {
		return err("missing userId")
}

	url := fmt.Sprintf("https://%s/api/v1/users/%s", domain, userId)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "SSWS "+token)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	var user map[string]interface{}
	if e := json.Unmarshal(body, &user); e != nil {
		return err(e.Error())
}

	profile, found := user["profile"].(map[string]interface{})
	if !found {
		return success("User found but no profile")
}

	email, _ := profile["email"].(string)
	return success(fmt.Sprintf("User: %s, email: %s", userId, email))
}
}
}