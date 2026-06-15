package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "q")
	if query == "" {
		return err("query 'q' is required")
}

	apiKey := os.Getenv("SERPAPI_KEY")
	cx := os.Getenv("SERPAPI_CX")
	if apiKey == "" || cx == "" {
		return err("SERPAPI_KEY and SERPAPI_CX environment variables required")
}

	url := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s", apiKey, cx, query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	items, found := result["items"].([]interface{})
	if !found {
		return success("No results found")
}

	out := ""
	for i, item := range items {
		m, found := item.(map[string]interface{})
		if !found {
			continue
		}
		title, _ := m["title"].(string)
		link, _ := m["link"].(string)
		out += fmt.Sprintf("%d. %s - %s\n", i+1, title, link)

	return success(out)
}
}