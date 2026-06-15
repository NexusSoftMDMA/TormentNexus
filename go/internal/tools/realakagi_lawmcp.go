//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchCases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 {

	}
	offset, _ :=getInt(args, "offset")
	if offset < 0 {

	}
	apiURL := "https://api.alphalawyer.com/v1/cases"
	u, e := url.Parse(apiURL)
	if e !=


-reasoner (deepseek)*,
}