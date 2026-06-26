package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleGetLatestNews_globalpulse_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	country, _ :=getString(args, "country")
	if country == "" {
		country = "us"
	}
	apiKey := os.Getenv("NEWS_API_KEY")
	url := fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=%s&apiKey=%s", country, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch news: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("News: %v", result))
}

func HandleGetNewsByCategory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	if category == "" {
		category = "general"
	}
	country, _ :=getString(args, "country")
	if country == "" {
		country = "us"
	}
	apiKey := os.Getenv("NEWS_API_KEY")
	url := fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=%s&category=%s&apiKey=%s", country, category, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch news: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Category news: %v", result))
}