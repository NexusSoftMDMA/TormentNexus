//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleArxivSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	maxResults, _ :=getInt(args, "max_results")
	if maxResults <= 0 {
		maxResults = 10
	}
	u := fmt.Sprintf("http://export.arxiv.org/api/query?search_query=all:%s&max_results=%d", url.QueryEscape(query), maxResults)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch arXiv: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read arXiv response: " + e.Error())
}

	return ok(string(body))
}

func HandleOpenAlexSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	maxResults, _ :=getInt(args, "max_results")
	if maxResults <= 0 {
		maxResults = 10
	}
	u := fmt.Sprintf("https://api.openalex.org/works?search=%s&per_page=%d", url.QueryEscape(query), maxResults)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch OpenAlex: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read OpenAlex response: " + e.Error())
}

	return ok(string(body))
}
