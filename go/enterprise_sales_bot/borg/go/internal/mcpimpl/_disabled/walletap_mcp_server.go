package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func WalletapCreatePass(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	baseURL, _ :=getString(args, "base_url")
	passData, _ :=getString(args, "pass_data")
	if apiKey == "" || baseURL == "" || passData == "" {
		return err("api_key, base_url and pass_data are required")
	}
	url := baseURL + "/pass"
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(passData))
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
	}
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %d %s", resp.StatusCode, string(body)))
	}
	return success(string(body))
}

func WalletapGetPass(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	baseURL, _ :=getString(args, "base_url")
	passID, _ :=getString(args, "pass_id")
	if apiKey == "" || baseURL == "" || passID == "" {
		return err("api_key, base_url and pass_id are required")
	}
	url := baseURL + "/pass/" + passID
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	req.Header.Set("X-API-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
	}
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %d %s", resp.StatusCode, string(body)))
	}
	return success(string(body))
}