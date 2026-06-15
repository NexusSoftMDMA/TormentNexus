package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleHyperbolicSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.hyperbolic.xyz/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("api returned status %d", resp.StatusCode))
}

	return success("Search results for " + query + " retrieved successfully")
}

func HandleHyperbolicStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.hyperbolic.xyz/status")
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return ok("Hyperbolic MCP server is online")
}

	return err("Hyperbolic MCP server is offline")
}// touch 1781132127
