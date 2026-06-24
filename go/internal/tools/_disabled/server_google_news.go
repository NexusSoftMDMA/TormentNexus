package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetNews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lang, _ :=getString(args, "language")
	if lang == "" {
		lang = "en-US"
	}
	url := fmt.Sprintf("https://news.google.com/rss?hl=%s&gl=US&ceid=US:en", lang)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch news: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	return ok(string(body))
}

func HandleSearchNews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
	}
	lang, _ :=getString(args, "language")
	if lang == "" {
		lang = "en-US"
	}
	url := fmt.Sprintf("https://news.google.com/rss/search?q=%s&hl=%s&gl=US&ceid=US:en", query, lang)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search news: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	return ok(string(body))
}