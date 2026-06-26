package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearchCoupons(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	resp, e := http.DefaultClient.Get("https://api.budgetfitter.com/search?q=" + query)
	if e != nil {
		return err("failed to search: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	deals, found := result["deals"].([]interface{})
	if !found {
		return ok("no deals found")
}

	return ok(fmt.Sprintf("found %d deals", len(deals)))
}

func HandleBrandIntelligence(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	brand, _ :=getString(args, "brand")
	if brand == "" {
		return err("brand is required")
}

	resp, e := http.DefaultClient.Get("https://api.budgetfitter.com/brand?name=" + brand)
	if e != nil {
		return err("failed to lookup brand: " + e.Error())
}

	defer resp.Body.Close()
	var info map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&info); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Brand: %s, offers: %v", brand, info["offers"]))
}