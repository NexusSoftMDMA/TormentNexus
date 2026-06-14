package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchOperator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
	}
	u := fmt.Sprintf("https://prts.wiki/api.php?action=query&list=search&srsearch=%s&format=json", url.QueryEscape(query))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed: " + e.Error())
	}
	return ok(fmt.Sprintf("Search results for '%s': %v", query, data))
}

func HandleGetOperator(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter required")
	}
	u := fmt.Sprintf("https://prts.wiki/api.php?action=parse&page=%s&prop=text&format=json", url.QueryEscape(name))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed: " + e.Error())
	}
	return ok(fmt.Sprintf("Operator data for '%s': %v", name, data))
}

func HandleSearchItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
	}
	u := fmt.Sprintf("https://prts.wiki/api.php?action=query&list=search&srsearch=%s+item&format=json", url.QueryEscape(query))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed: " + e.Error())
	}
	return ok(fmt.Sprintf("Item search results for '%s': %v", query, data))
}// touch 1781132139
