package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleGetProduct(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "product_id")
	if id == "" {
		return err("product_id is required")
}

	base := "https://api.agentshop.example.com/products/" + url.PathEscape(id)
	resp, e := http.DefaultClient.Get(base)
	if e != nil {
		return err("failed to fetch product: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(fmt.Sprintf("Product: %v", data))
}