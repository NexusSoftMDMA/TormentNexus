package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetCampaigns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	if token == "" {
		return err("access_token is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://business-api.tiktok.com/open_api/v1.3/campaign/get", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Access-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("api request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	message, found := result["message"]
	if !found {
		return err("no message in response")
}

	return ok(fmt.Sprintf("Campaigns: %v", message))
}

func HandleGetAdPerformance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	campaignID, _ :=getString(args, "campaign_id")
	if token == "" || campaignID == "" {
		return err("access_token and campaign_id are required")
}

	url := fmt.Sprintf("https://business-api.tiktok.com/open_api/v1.3/report/integrated/get?campaign_id=%s", campaignID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Access-Token", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("api request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	message, found := result["message"]
	if !found {
		return err("no message in response")
}

	return success(fmt.Sprintf("Performance: %v", message))
}