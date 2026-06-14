package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetUserProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := fmt.Sprintf("https://api.hashnode.com/v1/users/%s", username)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	user, found := data["user"].(map[string]interface{})
	if !found {
		return err("user not found")
}

	return ok(fmt.Sprintf("User: %v", user))
}