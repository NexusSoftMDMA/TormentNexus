package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type bicscanBalanceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	tag, _ :=getString(args, "tag")
	if tag == "" {
		tag = "latest"
	}
	url := fmt.Sprintf("https://api.bicscan.io/api?module=account&action=balance&address=%s&tag=%s", address, tag)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result bicscanBalanceResponse
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	if result.Status != "1" {
		return err(fmt.Sprintf("API error: %s", result.Message))
}

	return ok(fmt.Sprintf("Balance: %s", result.Result))
}