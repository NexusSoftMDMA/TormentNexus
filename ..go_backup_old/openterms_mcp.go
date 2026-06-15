package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchTerms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
}

	apiURL := fmt.Sprintf("https://api.openterms.io/search?q=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(apiURL)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
}

	results, found := data["results"].([]interface{})
	if !found {
		return ok("no results found")
}

	msg := "Search results:\n"
	for _, item := range results {
		m, _ := item.(map[string]interface{})
		if name, found := m["name"].(string); found {
			msg += fmt.Sprintf("- %s\n", name)

	}
	return success(msg)
}
}