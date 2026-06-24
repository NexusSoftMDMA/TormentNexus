package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := "https://api.peliqan.io/query?q=" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&result)
	if decodeErr != nil {
		return err(fmt.Sprintf("decode failed: %v", decodeErr))
}

	return ok(fmt.Sprintf("Result: %v", result))
}