package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleLogin_expensebot_mcp_server_auth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	password, _ :=getString(args, "password")

	body := map[string]string{"username": username, "password": password}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.expensebot.com/auth/login", bytes.NewReader(payload))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err("authentication failed")
}

	var result struct {
		Token string `json:"token"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response")
}

	return success(result.Token)
}