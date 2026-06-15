package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetCampaigns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	profileId, _ :=getString(args, "profileId")
	accessToken, _ :=getString(args, "accessToken")
	region, _ :=getString(args, "region")
	if region == "" {
		region = "NA"
	}
	baseURL := "https://advertising-api.amazon.com"
	if region == "EU" {
		baseURL = "https://advertising-api-eu.amazon.com"
	}
	url := fmt.Sprintf("%s/v2/sp/campaigns", baseURL)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Amazon-Advertising-API-ClientId", getString(args, "clientId"))
	req.Header.Set("Amazon-Advertising-API-Scope", profileId)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: "+e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status "+resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}