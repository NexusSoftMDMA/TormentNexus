package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetCampaign(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	campaignID, _ :=getString(args, "campaign_id")
	if campaignID == "" {
		return err("campaign_id is required")
}

	resp, e := http.DefaultClient.Get("https://api.kariad.com/campaigns/" + campaignID)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}

func HandleAnalyzeAd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	adID, _ :=getString(args, "ad_id")
	if adID == "" {
		return err("ad_id is required")
}

	resp, e := http.DefaultClient.Get("https://api.kariad.com/ads/" + adID + "/analysis")
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}