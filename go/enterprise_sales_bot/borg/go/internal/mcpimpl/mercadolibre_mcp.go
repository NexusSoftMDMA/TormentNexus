package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearchItems_mercadolibre_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.mercadolibre.com/sites/MLA/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Results []struct {
			ID    string  `json:"id"`
			Title string  `json:"title"`
			Price float64 `json:"price"`
		} `json:"results"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	if len(result.Results) == 0 {
		return ok("no results found")
}

	item := result.Results[0]
	msg := fmt.Sprintf("Found: %s (ID: %s, Price: %.2f)", item.Title, item.ID, item.Price)
	return ok(msg)
}

func HandleGetItem_mercadolibre_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	itemID, _ :=getString(args, "item_id")
	if itemID == "" {
		return err("item_id is required")
}

	url := fmt.Sprintf("https://api.mercadolibre.com/items/%s", itemID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get item: " + e.Error())
}

	defer resp.Body.Close()
	var item struct {
		ID    string  `json:"id"`
		Title string  `json:"title"`
		Price float64 `json:"price"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&item); e != nil {
		return err("failed to decode: " + e.Error())
}

	msg := fmt.Sprintf("Item: %s (ID: %s, Price: %.2f)", item.Title, item.ID, item.Price)
	return ok(msg)
}