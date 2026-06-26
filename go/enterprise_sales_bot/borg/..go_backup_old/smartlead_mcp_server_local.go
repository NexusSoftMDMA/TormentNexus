package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListCampaigns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.smartlead.ai/api/v1/campaigns", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("API-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(fmt.Sprintf("Campaigns: %s", string(data)))
}

func HandleGetCampaignStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	campaignID, _ :=getInt(args, "campaignId")
	url := fmt.Sprintf("https://api.smartlead.ai/api/v1/campaigns/%d/stats", campaignID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("API-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(fmt.Sprintf("Campaign stats: %s", string(data)))
}