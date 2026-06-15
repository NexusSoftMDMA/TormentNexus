package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func HandleListStreams(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "server_url")
	email, _ :=getString(args, "email")
	apiKey, _ :=getString(args, "api_key")
	if baseURL == "" || email == "" || apiKey == "" {
		return err("missing required parameters: server_url, email, api_key")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/v1/streams", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(email, apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "server_url")
	email, _ :=getString(args, "email")
	apiKey, _ :=getString(args, "api_key")
	to, _ :=getString(args, "to")
	content, _ :=getString(args, "content")
	msgType, _ :=getString(args, "type")
	if baseURL == "" || email == "" || apiKey == "" || to == "" || content == "" || msgType == "" {
		return err("missing required parameters")
}

	form := url.Values{}
	form.Set("type", msgType)
	form.Set("to", to)
	form.Set("content", content)
	req, e := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/v1/messages", strings.NewReader(form.Encode()))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(email, apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}