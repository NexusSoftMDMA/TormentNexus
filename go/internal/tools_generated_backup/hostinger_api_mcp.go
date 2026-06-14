package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListDomains(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
	}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.hostinger.com/v1/domains", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
	}

	return ok(string(body))
}

func HandleGetDomain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	domain, _ :=getString(args, "domain")
	if apiKey == "" || domain == "" {
		return err("api_key and domain are required")
	}

	url := fmt.Sprintf("https://api.hostinger.com/v1/domains/%s", domain)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
	}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
	}

	return success(fmt.Sprintf("Domain: %v", result))
}// touch 1781132127
