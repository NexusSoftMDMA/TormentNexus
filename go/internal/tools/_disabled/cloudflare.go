package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func HandleListZones(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "api_token")
	if token == "" {
		token = os.Getenv("CLOUDFLARE_API_TOKEN")

	if token == "" {
		return err("missing api_token")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.cloudflare.com/client/v4/zones", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Success bool              `json:"success"`
		Errors  []json.RawMessage `json:"errors"`
		Result  []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"result"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if !result.Success {
		return err("API error: " + fmt.Sprintf("%v", result.Errors))
}

	jsonOut, _ := json.Marshal(result.Result)
	return ok(string(jsonOut))
}

}

func HandleGetDNSRecords(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "api_token")
	if token == "" {
		token = os.Getenv("CLOUDFLARE_API_TOKEN")

	if token == "" {
		return err("missing api_token")
}

	zoneID, _ :=getString(args, "zone_id")
	if zoneID == "" {
		return err("missing zone_id")
}

	apiURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", url.PathEscape(zoneID))
	req, e := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Success bool              `json:"success"`
		Errors  []json.RawMessage `json:"errors"`
		Result  []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Type    string `json:"type"`
			Content string `json:"content"`
		} `json:"result"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if !result.Success {
		return err("API error: " + fmt.Sprintf("%v", result.Errors))
}

	jsonOut, _ := json.Marshal(result.Result)
	return ok(string(jsonOut))
}
}