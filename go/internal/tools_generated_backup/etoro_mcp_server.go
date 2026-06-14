package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func HandleGetUserProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := fmt.Sprintf("https://api.etoro.com/api/v1/users/%s", username)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch profile")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("eToro API returned %d", resp.StatusCode))
}

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response")
}

	return ok(fmt.Sprintf("User profile: %v", data))
}

func HandleGetPortfolio(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := fmt.Sprintf("https://api.etoro.com/api/v1/users/%s/portfolio", username)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch portfolio")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("eToro API returned %d", resp.StatusCode))
}

	var data []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response")
}

	return ok(fmt.Sprintf("Portfolio: %v", data))
}