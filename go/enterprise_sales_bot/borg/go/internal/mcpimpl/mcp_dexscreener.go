package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchPairs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.dexscreener.com/latest/dex/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result struct {
		Pairs []map[string]interface{} `json:"pairs"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	pairs := result.Pairs
	if pairs == nil {
		pairs = []map[string]interface{}{}
	}
	return success(fmt.Sprintf("Found %d pairs", len(pairs)))
}

func HandleGetTokenInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chain, _ :=getString(args, "chain")
	address, _ :=getString(args, "address")
	if chain == "" || address == "" {
		return err("chain and address are required")
}

	url := fmt.Sprintf("https://api.dexscreener.com/latest/dex/tokens/%s", address)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result struct {
		Pairs []map[string]interface{} `json:"pairs"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	if len(result.Pairs) == 0 {
		return err("no token data found")
}

	return success(fmt.Sprintf("Found token with %d pairs", len(result.Pairs)))
}