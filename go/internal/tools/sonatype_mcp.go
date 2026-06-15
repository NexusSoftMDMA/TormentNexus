//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchComponents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "url")
	query, _ :=getString(args, "query")
	if baseURL == "" || query == "" {
		return err("url and query required")
}

	u := fmt.Sprintf("%s/service/rest/v1/search?q=%s", baseURL, url.QueryEscape(query


-reasoner (deepseek)*
}