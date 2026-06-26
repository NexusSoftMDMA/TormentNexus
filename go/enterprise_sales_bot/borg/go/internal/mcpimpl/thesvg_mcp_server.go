package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchBrand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := fmt.Sprintf("https://thesvg.org/api/search?q=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("search request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Brands []string `json:"brands"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d brands: %v", len(result.Brands), result.Brands))
}

func HandleFetchBrandSvg(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	brand, _ :=getString(args, "brand")
	if brand == "" {
		return err("brand is required")
}

	u := fmt.Sprintf("https://thesvg.org/api/brand/%s", url.PathEscape(brand))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return success(string(body))
}