package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func HandleSearchFood_mcp_opennutrition(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://world.openfoodfacts.org/cgi/search.pl?search_terms=%s&json=1&page_size=5", query)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]json.RawMessage
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	products := result["products"]
	if products == nil {
		return ok("no products found")
}

	data, _ := json.MarshalIndent(products, "", "  ")
	return success(string(data))
}

func HandleGetNutrition(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	barcode, _ :=getString(args, "barcode")
	if barcode == "" {
		return err("barcode is required")
}

	url := fmt.Sprintf("https://world.openfoodfacts.org/api/v0/product/%s.json", barcode)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]json.RawMessage
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	product := result["product"]
	if product == nil {
		return ok("no product found for barcode " + barcode)
}

	data, _ := json.MarshalIndent(product, "", "  ")
	return success(string(data))
}