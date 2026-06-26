package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetProduct_supplymaven_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	productID, _ :=getString(args, "product_id")
	if productID == "" {
		return err("product_id is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.supplymaven.com/products/%s", productID))
	if e != nil {
		return err(fmt.Sprintf("API request failed: %v", e))
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("Product: %v", data))
}

func HandleGetInventory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sku, _ :=getString(args, "sku")
	if sku == "" {
		return err("sku is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.supplymaven.com/inventory/%s", sku))
	if e != nil {
		return err(fmt.Sprintf("API request failed: %v", e))
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(fmt.Sprintf("Inventory: %v", data))
}