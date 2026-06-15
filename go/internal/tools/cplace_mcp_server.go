//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func ListWorkspacesHandler(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := "https://api.cplace.com/workspaces"
	if query != "" {
		url += "?q=" + query,
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("")
}


-reasoner (deepseek)*
}