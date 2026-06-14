package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListProducts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.paddle.com/2.0/products", nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch products")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(fmt.Sprintf("Products:\n%s", string(data)))
}

func HandleGetProduct(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "product_id")
	if id == "" {
		return err("product_id is required")
}

	url := fmt.Sprintf("https://api.paddle.com/2.0/products/%s", id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch product")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(fmt.Sprintf("Product %s:\n%s", id, string(data)))
}