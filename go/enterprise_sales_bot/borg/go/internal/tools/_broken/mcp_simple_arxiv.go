package mcpimpl

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
)

type arxivFeed struct {
	Entries []arxivEntry `xml:"entry"`
}

type arxivEntry struct {
	ID      string `xml:"id"`
	Title   string `xml:"title"`
	Summary string `xml:"summary"`
}

func HandleSearchArxiv_mcp_simple_arxiv(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	maxResults, _ :=getInt(args, "max_results")
	if maxResults <= 0 {
		maxResults = 10
	}
	u := fmt.Sprintf("http://export.arxiv.org/api/query?search_query=all:%s&max_results=%d", url.QueryEscape(query), maxResults)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to query arxiv: " + e.Error())
}

	defer resp.Body.Close()
	var feed arxivFeed
	if e := xml.NewDecoder(resp.Body).Decode(&feed); e != nil {
		return err("failed to parse response: " + e.Error())
}

	var out string
	for _, entry := range feed.Entries {
		out += fmt.Sprintf("Title: %s\nID: %s\nSummary: %s\n\n", entry.Title, entry.ID, entry.Summary)

	return success(out)
}
}