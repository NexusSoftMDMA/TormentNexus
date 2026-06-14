package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSanctionsSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	apiURL := "https://api.opensanctions.org/search?q=" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err("failed to fetch sanctions: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Sanctions results: %v", result))
}