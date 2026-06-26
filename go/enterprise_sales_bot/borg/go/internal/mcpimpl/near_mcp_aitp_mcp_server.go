package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetNearBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	account, _ :=getString(args, "account_id")
	if account == "" {
		return err("missing account_id")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.near.org/account/"+account+"/balance", nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}
	return success("balance retrieved")
}

func HandleGetNearStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.near.org/status", nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	return success("near network is operational")
}