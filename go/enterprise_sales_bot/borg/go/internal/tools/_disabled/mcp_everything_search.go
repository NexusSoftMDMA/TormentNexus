package tools

import (
	"context"
	"fmt"
	"net/http"
	"encoding/json"
	"io"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json", query))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	abstract, found := result["AbstractText"]
	if !found {
		abstract = "no abstract"
	}

	return ok(fmt.Sprintf("Search result for '%s': %s", query, abstract))
}