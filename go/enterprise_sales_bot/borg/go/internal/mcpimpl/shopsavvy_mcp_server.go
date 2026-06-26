package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

func HandleSearchProducts_shopsavvy_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := "https://api.shopsavvy.com/search?q=" + url.QueryEscape(query)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch data: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(result)
}

func HandleGetProductDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "productId")
	if id == "" {
		return err("productId is required")
}

	u := "https://api.shopsavvy.com/products/" + url.PathEscape(id)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch data: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(result)
}