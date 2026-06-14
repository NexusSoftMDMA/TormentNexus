package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetUserInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := fmt.Sprintf("https://api.roblox.com/users/get-by-username?username=%s", username)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	if id, found := data["Id"]; found {
		return ok(fmt.Sprintf("User found: ID %v, Username %v", id, data["Username"]))
}

	return err("user not found")
}