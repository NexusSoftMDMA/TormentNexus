package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type searchResult struct {
	NumFound int `json:"numFound"`
	Docs     []struct {
		Title      string   `json:"title"`
		AuthorName []string `json:"author_name"`
	} `json:"docs"`
}

type workResult struct {
	Title string `json:"title"`
	Authors []struct {
		Name string `json:"name"`
	} `json:"authors"`
}

func HandleSearchBooks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	resp, e := http.DefaultClient.Get("https://openlibrary.org/search.json?q=" + query)
	if e != nil {
		return err("search failed: " + e.Error())
}

	defer resp.Body.Close()
	var data searchResult
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	if data.NumFound == 0 {
		return ok("No results found.")
}

	first := data.Docs[0]
	author := ""
	if len(first.AuthorName) > 0 {
		author = " by " + first.AuthorName[0]
	}
	return ok(fmt.Sprintf("Found %d results. First: %s%s", data.NumFound, first.Title, author))
}

func HandleGetBook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	olid, _ :=getString(args, "olid")
	if olid == "" {
		return err("olid is required")
}

	resp, e := http.DefaultClient.Get("https://openlibrary.org/works/" + olid + ".json")
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	var data workResult
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	author := ""
	if len(data.Authors) > 0 {
		author = " by " + data.Authors[0].Name
	}
	return ok(fmt.Sprintf("Title: %s%s", data.Title, author))
}