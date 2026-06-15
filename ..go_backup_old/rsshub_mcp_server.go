package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleFetchFeed(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	route, _ :=getString(args, "route")
	if route == "" {
		return err("route is required")
}

	url := fmt.Sprintf("https://rsshub.app/%s", route)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return success(string(body))
}

func HandleSearchRoutes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://rsshub.app/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return success(string(body))
}// touch 1781132140
