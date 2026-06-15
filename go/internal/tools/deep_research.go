//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleDeepResearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	maxDepth, _ :=getInt(args, "max_depth")
	if maxDepth <= 0 {
		maxDepth = 3
	}
	apiURL := os.Getenv("DEEP_RESEARCH_API_URL")
	if apiURL == "" {
		return err("DEEP_RESEARCH_API_URL not set")
}

	url := fmt.Sprintf("%s?query=%s&max_depth=%d", apiURL, query, maxDepth)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse response failed: %v", e))
}

	return ok(fmt.Sprintf("Research result: %v", result))
}
