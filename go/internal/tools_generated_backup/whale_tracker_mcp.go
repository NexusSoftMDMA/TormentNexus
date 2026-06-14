package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetWhaleTransactions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token_address")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.whaletracker.com/v1/transactions?token="+token, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch transactions")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(string(body))
}

func HandleGetWhalePortfolio(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	whale, _ :=getString(args, "whale_address")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.whaletracker.com/v1/portfolio?address="+whale, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch portfolio")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(string(body))
}