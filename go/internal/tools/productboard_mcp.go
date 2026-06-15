//go:build ignore
// +build ignore

package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

func HandleGetProduct(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    apiKey, _ :=getString(args, "api_key")
    productID, _ :=getString(args, "product_id")
    if apiKey == "" || productID == "" {
        return err("api_key and product_id are required")
}

    url := fmt.Sprintf("https://api.productboard.com/products/%s", productID)
    req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
    if e != nil {
        return err("failed to create request: " + e.Error())
}

    req.Header.Set("Authorization", "Bearer " + apiKey)
    req.Header.Set("Content-Type", "application/json")
    resp, e := http.DefaultClient.Do(req)
    if e != nil {
        return err("failed to execute request: " + e.Error())
}

    defer resp.Body


-reasoner (deepseek)*
}