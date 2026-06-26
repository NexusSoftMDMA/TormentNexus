package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleGetUsers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	realm, _ :=getString(args, "realm")
	token, _ :=getString(args, "token")

	url := fmt.Sprintf("%s/admin/realms/%s/users", baseURL, realm)
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

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("non-200 status: %d", resp.StatusCode))
}

	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	return ok(string(body))
}

func HandleCreateUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	realm, _ :=getString(args, "realm")
	token, _ :=getString(args, "token")
	username, _ :=getString(args, "username")
	email, _ :=getString(args, "email")
	enabled, _ :=getBool(args, "enabled")

	user := map[string]interface{}{
		"username": username,
		"email":    email,
		"enabled":  enabled,
	}

	bodyBytes, e := json.Marshal(user)
	if e != nil {
		return err("failed to marshal user: " + e.Error())
}

	url := fmt.Sprintf("%s/admin/realms/%s/users", baseURL, realm)
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return err(fmt.Sprintf("non-201 status: %d, body: %s", resp.StatusCode, string(body)))
}

	return ok("user created")
}