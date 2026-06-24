package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchLittleSis(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query parameter")
	}
	base := "https://api.littlesis.org/api/search"
	params := url.Values{}
	params.Add("q", query)
	params.Add("format", "json")
	reqURL := base + "?" + params.Encode()
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}
	return success(fmt.Sprintf("Search results for '%s': %v", query, result))
}

func HandleGetEntityDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing id parameter")
	}
	base := "https://api.littlesis.org/api/entity/" + id
	params := url.Values{}
	params.Add("format", "json")
	reqURL := base + "?" + params.Encode()
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}
	return success(fmt.Sprintf("Entity %s details: %v", id, result))
}// touch 1781132130
