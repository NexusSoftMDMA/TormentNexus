package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	apiKey, _ :=getString(args, "api_key")
	query, _ :=getString(args, "query")
	if host == "" || apiKey == "" || query == "" {
		return err("host, api_key, and query are required")
	}

	u, e := url.Parse(host)
	if e != nil {
		return err("invalid host: " + e.Error())
	}
	u.Path = "/collections/documents/search"
	u.RawQuery = url.Values{"q": {query}, "query_by": {"title"}}.Encode()

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("X-TYPESENSE-API-KEY", apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
	}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("unmarshal failed: " + e.Error())
	}

	hits, found := result["hits"].([]interface{})
	if !found {
		return success(fmt.Sprintf("No results. Response: %s", string(body)))
	}
	return success(fmt.Sprintf("Found %d hits", len(hits)))
}