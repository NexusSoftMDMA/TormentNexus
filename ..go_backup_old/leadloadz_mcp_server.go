package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchLeads(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	industry, _ :=getString(args, "industry")
	location, _ :=getString(args, "location")
	title, _ :=getString(args, "title")

	u := fmt.Sprintf("https://api.leadloadz.com/search?industry=%s&location=%s&title=%s", industry, location, title)
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
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

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	out, _ := json.Marshal(result)
	return ok(string(out))
}

func HandleVerifyEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	email, _ :=getString(args, "email")
	if email == "" {
		return err("email is required")
}

	u := fmt.Sprintf("https://api.leadloadz.com/verify?email=%s", email)
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
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

	return ok(string(body))
}