package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetFanTokenInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://api.chiliz.com/marketing/fan-token/%s", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("Fan token info: %v", result))
}

func HandleGetCampaignMetrics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "campaign_id")
	if id == "" {
		return err("campaign_id is required")
}

	url := fmt.Sprintf("https://api.chiliz.com/marketing/campaign/%s/metrics", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("Campaign metrics: %v", result))
}