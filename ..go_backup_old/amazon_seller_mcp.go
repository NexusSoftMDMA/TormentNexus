package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetOrders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	marketplaceId, _ :=getString(args, "marketplace_id")
	createdAfter, _ :=getString(args, "created_after")
	url := fmt.Sprintf("https://api.amazon.seller.com/orders?marketplaceId=%s&createdAfter=%s", marketplaceId, createdAfter)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch orders: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Orders: %v", result))
}

func HandleSearchProducts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	marketplaceId, _ :=getString(args, "marketplace_id")
	url := fmt.Sprintf("https://api.amazon.seller.com/products/search?keyword=%s&marketplaceId=%s", keyword, marketplaceId)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search products: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Products: %v", result))
}