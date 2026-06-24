package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type monzoAccount struct {
	ID      string `json:"id"`
	Balance int64  `json:"balance"`
}

type monzoBalance struct {
	Balance int64 `json:"balance"`
}

func HandleGetAccounts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("MONZO_ACCESS_TOKEN")
	if token == "" {
		return err("MONZO_ACCESS_TOKEN not set")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.monzo.com/accounts", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data struct {
		Accounts []monzoAccount `json:"accounts"`
	}
	if e = json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	out, _ := json.Marshal(data.Accounts)
	return ok(string(out))
}

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	accountID, _ :=getString(args, "account_id")
	if accountID == "" {
		return err("account_id is required")
}

	token := os.Getenv("MONZO_ACCESS_TOKEN")
	if token == "" {
		return err("MONZO_ACCESS_TOKEN not set")
}

	url := fmt.Sprintf("https://api.monzo.com/balance?account_id=%s", accountID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var bal monzoBalance
	if e = json.Unmarshal(body, &bal); e != nil {
		return err("failed to parse response: " + e.Error())
}

	out, _ := json.Marshal(bal)
	return ok(string(out))
}