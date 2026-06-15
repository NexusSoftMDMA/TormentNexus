package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetLatestFiling(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticker, _ :=getString(args, "ticker")
	if ticker == "" {
		return err("ticker is required")
}

	url := fmt.Sprintf("https://api.signal8.ai/filings/latest?ticker=%s", ticker)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	return ok(string(body))
}

func HandleSearchFilings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.signal8.ai/filings/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	return ok(string(body))
}