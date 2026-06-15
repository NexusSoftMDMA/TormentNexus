//go:build ignore
// +build ignore

package tools

import (
	"context"
	"io"
	"net/http"
)

func HandlePrismaQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "api_url")
	table, _ :=getString(args, "table_name")
	if url == "" {
		return err("api_url is required")
		if table == "" {
		return err("table_name is required")
	}
	fullURL := url + "/tables/" + table
	resp, e :=


-reasoner (deepseek)*
}
}