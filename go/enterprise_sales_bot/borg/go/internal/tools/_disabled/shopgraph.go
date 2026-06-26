package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func fetchJSON(path string, target interface{}) error {
	base := os.Getenv("SHOPGRAPH_API_BASE")
	if base == "" {
		base = "https://api.shopgraph.app"
	}
	resp, e := http.DefaultClient.Get(base + path)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return e
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
}

	return json.Unmarshal(body, target)
}

func HandleGetProduct(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	productID, _ :=getString(args, "product_id")
	if productID == "" {
		return err("product_id is required")
}

	var product map[string]interface{}
	e := fetchJSON("/products/"+productID, &product)
	if e != nil {
		return err("failed to fetch product: " + e.Error())
}

	return success(product)
}

func HandleListProducts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	path := "/products"
	if category != "" {
		path += "?category=" + category
	}
	var products []map[string]interface{}
	e := fetchJSON(path, &products)
	if e != nil {
		return err("failed to list products: " + e.Error())
}

	return success(products)
}