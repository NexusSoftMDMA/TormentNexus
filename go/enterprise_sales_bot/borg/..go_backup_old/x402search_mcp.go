package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type x402Result struct {
	AbstractText string `json:"AbstractText"`
}

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := "https://api.duckduckgo.com/?q=" + url.QueryEscape(query) + "&format=json"
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("search failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result x402Result
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	return ok(fmt.Sprintf("Result: %s", result.AbstractText))
}