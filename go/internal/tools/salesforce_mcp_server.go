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

func HandleQueryToolingApi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query required")
}

	base, _ :=getString(args, "instance_url")
	token, _ :=getString(args, "access_token")
	if base == "" || token == "" {
		return err("credentials required")
}

	u := fmt.Sprintf("%s/services/data/v60.0/tooling/query?q=%s", base, url.QueryEscape(query))
	req, e


-reasoner (deepseek)*
}