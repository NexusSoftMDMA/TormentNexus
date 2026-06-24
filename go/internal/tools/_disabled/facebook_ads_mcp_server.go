package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListAdAccounts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	url := fmt.Sprintf("https://graph.facebook.com/v20.0/me/adaccounts?access_token=%s", token)
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

	return success(fmt.Sprintf("Ad accounts: %v", result["data"]))
}

func HandleListCampaigns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	accountID, _ :=getString(args, "ad_account_id")
	url := fmt.Sprintf("https://graph.facebook.com/v20.0/act_%s/campaigns?access_token=%s", accountID, token)
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

	return success(fmt.Sprintf("Campaigns: %v", result["data"]))
}