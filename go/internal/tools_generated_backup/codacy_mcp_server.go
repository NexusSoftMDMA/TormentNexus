package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListRepositories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("CODACY_API_TOKEN")
	if apiKey == "" {
		return err("CODACY_API_TOKEN not set")
}

	provider, _ :=getString(args, "provider")
	if provider == "" {
		provider = "gh"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.codacy.com/2.0/account/repositories/%s", provider), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("api_token", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return ok("repositories retrieved: " + string(body))
}

func HandleGetCoverage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("CODACY_API_TOKEN")
	if apiKey == "" {
		return err("CODACY_API_TOKEN not set")
}

	provider, _ :=getString(args, "provider")
	if provider == "" {
		provider = "gh"
	}
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repository")
	if owner == "" || repo == "" {
		return err("owner and repository are required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.codacy.com/2.0/coverage/%s/%s/%s", provider, owner, repo), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("api_token", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return ok(fmt.Sprintf("coverage data: %s", string(body)))
}