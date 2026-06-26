package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListSatellites(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("JUNO_API_TOKEN")
	if token == "" {
		return err("JUNO_API_TOKEN not set")
}

	baseURL := os.Getenv("JUNO_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.juno.build/v1"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/satellites", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("%v", result))
}

func HandleGetSatellite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	token := os.Getenv("JUNO_API_TOKEN")
	if token == "" {
		return err("JUNO_API_TOKEN not set")
}

	baseURL := os.Getenv("JUNO_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.juno.build/v1"
	}
	url := fmt.Sprintf("%s/satellites/%s", baseURL, name)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("%v", result))
}