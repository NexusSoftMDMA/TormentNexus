package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetProduct(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	productID, _ :=getString(args, "product_id")
	if productID == "" {
		return err("product_id is required")
}

	token := os.Getenv("GUMROAD_TOKEN")
	if token == "" {
		return err("GUMROAD_TOKEN not set")
}

	url := fmt.Sprintf("https://api.gumroad.com/v2/products/%s", productID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	product, found := result["product"]
	if !found {
		return err("product not found in response")
}

	productJSON, e := json.Marshal(product)
	if e != nil {
		return err("failed to marshal product: " + e.Error())
}

	return ok(string(productJSON))
}

func HandleListProducts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("GUMROAD_TOKEN")
	if token == "" {
		return err("GUMROAD_TOKEN not set")
}

	url := "https://api.gumroad.com/v2/products"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return ok(string(body))
}