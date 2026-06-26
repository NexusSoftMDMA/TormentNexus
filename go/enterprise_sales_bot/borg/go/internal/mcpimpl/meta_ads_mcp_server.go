package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HandleListAdAccounts fetches ad accounts for the authenticated user.
func HandleListAdAccounts_meta_ads_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	url := fmt.Sprintf("https://graph.facebook.com/v19.0/me/adaccounts?access_token=%s&fields=id,name,account_status", token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to request ad accounts: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	if data, found := result["data"]; found {
		return ok(fmt.Sprintf("Ad accounts: %v", data))
}

	if er, found := result["error"]; found {
		return err(fmt.Sprintf("API error: %v", er))
}

	return err("unexpected response")
}

// HandleListCampaigns fetches campaigns for a given ad account.
func HandleListCampaigns_meta_ads_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	actId, _ :=getString(args, "ad_account_id")
	url := fmt.Sprintf("https://graph.facebook.com/v19.0/act_%s/campaigns?access_token=%s&fields=id,name,status", actId, token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to request campaigns: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	if data, found := result["data"]; found {
		return ok(fmt.Sprintf("Campaigns: %v", data))
}

	if er, found := result["error"]; found {
		return err(fmt.Sprintf("API error: %v", er))
}

	return err("unexpected response")
}