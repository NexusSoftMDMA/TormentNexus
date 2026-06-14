package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleDuckDuckGoSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("search request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	var data map[string]interface{}
	if e = json.Unmarshal(body, &data); e != nil {
		return err("parse response failed: " + e.Error())
}

	abstract := ""
	if v, found := data["AbstractText"]; found {
		if s, found := v.(string); found {
			abstract = s
		}
	}
	if abstract == "" {
		return ok("No summary available.")
}

	return ok(abstract)
}