//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type arxivEntry struct {
	Title string `xml:"title"`
	ID    string `xml:"id"`
}

type arxivFeed struct {
	XMLName xml.Name      `xml:"feed"`
	Entries []arxivEntry  `xml:"entry"`
}

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	u := fmt.Sprintf("http://export.arxiv.org/api/query?search_query=all:%s&max_results=5", url.QueryEscape(q))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var feed arxivFeed
	if e := xml.Unmarshal(body, &feed); e != nil {
		return err("parse failed: " + e.Error())
}

	if len(feed.Entries) == 0 {
		return ok("no results found")
}

	result := ""
	for _, entry := range feed.Entries {
		result += fmt.Sprintf("- [%s](%s)\n", entry.Title, entry.ID)

	return ok(result)
}

}

func HandleGetPaper(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	u := fmt.Sprintf("http://export.arxiv.org/api/query?id_list=%s", url.QueryEscape(id))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var feed arxivFeed
	if e := xml.Unmarshal(body, &feed); e != nil {
		return err("parse failed: " + e.Error())
}

	if len(feed.Entries) == 0 {
		return err("paper not found")
}

	entry := feed.Entries[0]
	return ok(fmt.Sprintf("Title: %s\nID: %s", entry.Title, entry.ID))
}
