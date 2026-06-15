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
	if token == "" {
		return err("missing access_token")
}

	url := fmt.Sprintf("https://graph.facebook.com/v18.0/me/adaccounts?access_token=%s", token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	if data, found := result["data"]; found {
		arr, _ := data.([]interface{})
		return ok(fmt.Sprintf("Found %d ad accounts", len(arr)))
}

	return ok("No ad accounts found")
}

func HandleGetCampaigns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "access_token")
	accountID, _ :=getString(args, "account_id")
	if token == "" || accountID == "" {
		return err("missing access_token or account_id")
}

	url := fmt.Sprintf("https://graph.facebook.com/v18.0/act_%s/campaigns?access_token=%s", accountID, token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	if data, found := result["data"]; found {
		arr, _ := data.([]interface{})
		return ok(fmt.Sprintf("Found %d campaigns", len(arr)))
}

	return ok("No campaigns found")
}