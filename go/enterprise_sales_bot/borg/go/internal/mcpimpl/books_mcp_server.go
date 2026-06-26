package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func SearchBooks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query")
}

	url := fmt.Sprintf("https://openlibrary.org/search.json?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Docs []struct {
			Title    string   `json:"title"`
			Author   []string `json:"author_name"`
			Key      string   `json:"key"`
		} `json:"docs"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	if len(result.Docs) == 0 {
		return ok("no books found")
}

	book := result.Docs[0]
	text := fmt.Sprintf("Title: %s\nAuthor: %v\nKey: %s", book.Title, book.Author, book.Key)
	return ok(text)
}

func GetBookDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing id")
}

	url := fmt.Sprintf("https://openlibrary.org/works/%s.json", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	var data struct {
		Title   string `json:"title"`
		Authors []struct {
			Name string `json:"name"`
		} `json:"authors"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	text := fmt.Sprintf("Title: %s\nAuthor(s): ", data.Title)
	for i, a := range data.Authors {
		if i > 0 {
			text += ", "
		}
		text += a.Name
	}
	return ok(text)
}