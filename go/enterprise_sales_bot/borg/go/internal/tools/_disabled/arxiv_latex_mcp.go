package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetArxivLatex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "arxiv_id")
	if id == "" {
		return err("arxiv_id is required")
}

	url := fmt.Sprintf("https://arxiv.org/e-print/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	return success(string(body))
}

func HandleSearchArxiv(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	max, _ :=getInt(args, "max_results")
	if max == 0 {
		max = 10
	}
	url := fmt.Sprintf("https://export.arxiv.org/api/query?search_query=all:%s&max_results=%d", query, max)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	return success(string(body))
}