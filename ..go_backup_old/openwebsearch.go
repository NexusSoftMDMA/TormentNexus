package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func HandleOpenWebSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	engine, _ :=getString(args, "engine")
	if query == "" {
		return err("query is required")
	}
	if engine == "" {
		engine = "bing"
	}
	baseURLs := map[string]string{
		"bing":       "https://www.bing.com/search?q=",
		"baidu":      "https://www.baidu.com/s?wd=",
		"duckduckgo": "https://duckduckgo.com/?q=",
		"brave":      "https://search.brave.com/search?q=",
		"exa":        "https://exa.ai/search?q=",
		"github":     "https://github.com/search?q=",
		"juejin":     "https://juejin.cn/search?query=",
		"csdn":       "https://so.csdn.net/so/search?q=",
	}
	base, found := baseURLs[strings.ToLower(engine)]
	if !found {
		return err("unsupported engine: " + engine)
	}
	fullURL := base + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(fullURL)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return success(fmt.Sprintf("Search initiated on %s for '%s'. Direct URL: %s (HTML preview not parsed)", engine, query, fullURL))
	}
	return success(fmt.Sprintf("Search results from %s for '%s': %v", engine, query, result))
}

func HandleOpenWebSearchList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	engines := []string{"bing", "baidu", "duckduckgo", "brave", "exa", "github", "juejin", "csdn"}
	return success(fmt.Sprintf("Available engines: %s", strings.Join(engines, ", ")))
}// touch 1781132137
