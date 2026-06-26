package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchCloudDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	provider, _ :=getString(args, "provider")
	if provider == "" {
		provider = "all"
	}
	url := fmt.Sprintf("https://docs.example.com/search?q=%s&provider=%s", query, provider)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(body))
}// touch 1781132135
