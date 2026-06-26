package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleSearchBoycott(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
	}
	url := "https://api.boikot.com/v1/search?q=" + query
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
	}
	return success("search results")
}

func HandleGetAlternatives(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	productID, _ :=getString(args, "product_id")
	if productID == "" {
		return err("product_id parameter required")
	}
	url := "https://api.boikot.com/v1/products/" + productID + "/alternatives"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
	}
	return success("alternatives found")
}// touch 1781132120
