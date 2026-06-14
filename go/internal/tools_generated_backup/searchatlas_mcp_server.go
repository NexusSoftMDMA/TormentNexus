package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleSearchQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
	}
	url := fmt.Sprintf("https://api.searchatlas.com/search?q=%s", strings.ReplaceAll(query, " ", "+"))
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
	}
	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
	}
	return ok(fmt.Sprintf("Search results: %v", data["results"]))
}

func HandleGetRanking(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	domain, _ :=getString(args, "domain")
	if keyword == "" || domain == "" {
		return err("keyword and domain are required")
	}
	url := fmt.Sprintf("https://api.searchatlas.com/ranking?keyword=%s&domain=%s",
		strings.ReplaceAll(keyword, " ", "+"), domain)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
	}
	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
	}
	rank, found := data["rank"].(float64)
	if !found {
		return err("rank not found")
	}
	return success(fmt.Sprintf("Keyword '%s' ranks at position %.0f", keyword, rank))
}