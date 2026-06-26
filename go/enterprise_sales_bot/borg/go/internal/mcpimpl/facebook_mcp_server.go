package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func HandleFacebookGetUser(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("FACEBOOK_ACCESS_TOKEN")
	if token == "" {
		return err("Missing FACEBOOK_ACCESS_TOKEN environment variable")
}

	u := fmt.Sprintf("https://graph.facebook.com/v12.0/me?access_token=%s", url.QueryEscape(token))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("Failed to fetch user: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("Failed to parse response: " + e.Error())
}

	return success(fmt.Sprintf("User info: %v", result))
}

func HandleFacebookPostStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("FACEBOOK_ACCESS_TOKEN")
	if token == "" {
		return err("Missing FACEBOOK_ACCESS_TOKEN environment variable")
}

	message, _ :=getString(args, "message")
	if message == "" {
		return err("Missing 'message' argument")
}

	u := fmt.Sprintf("https://graph.facebook.com/v12.0/me/feed?access_token=%s", url.QueryEscape(token))
	payload := strings.NewReader(url.Values{"message": {message}}.Encode())
	resp, e := http.Post(u, "application/x-www-form-urlencoded", payload)
	if e != nil {
		return err("Failed to post status: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	return success(fmt.Sprintf("Status posted: %s", string(body)))
}