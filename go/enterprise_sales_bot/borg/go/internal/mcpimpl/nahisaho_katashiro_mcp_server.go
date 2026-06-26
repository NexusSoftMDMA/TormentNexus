package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchPapers_nahisaho_katashiro_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	apiURL := fmt.Sprintf("https://api.katashiro.ai/search?q=%s&limit=%d", url.QueryEscape(query), limit)
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(fmt.Sprintf("search request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("search returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse response failed: %v", e))
}

	return ok(fmt.Sprintf("Search results: %v", result))
}

func HandleFetchPaper(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paperURL, _ :=getString(args, "url")
	if paperURL == "" {
		return err("missing paper URL")
}

	resp, e := http.DefaultClient.Get(paperURL)
	if e != nil {
		return err(fmt.Sprintf("fetch paper request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read paper response failed: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("fetch paper returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse paper data failed: %v", e))
}

	return success(fmt.Sprintf("Paper data: %v", result))
}