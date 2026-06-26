package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchProducts_buywhere_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing required argument: query")
	}
	market, _ :=getString(args, "market")
	if market == "" {
		market = "SG"
	}
	base := "https://api.buywhere.io/v1/search"
	params := url.Values{}
	params.Set("q", query)
	params.Set("market", market)
	req, e := http.NewRequestWithContext(ctx, "GET", base+"?"+params.Encode(), nil)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
	}
	return success(fmt.Sprintf("Found products for '%s' in %s", query, market))
}

func HandleComparePrices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	productID, _ :=getString(args, "product_id")
	if productID == "" {
		return err("missing required argument: product_id")
	}
	base := "https://api.buywhere.io/v1/compare"
	params := url.Values{}
	params.Set("id", productID)
	req, e := http.NewRequestWithContext(ctx, "GET", base+"?"+params.Encode(), nil)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http error: %v", e))
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
	}
	return success(fmt.Sprintf("Price comparison retrieved for product %s", productID))
}// touch 1781132121
