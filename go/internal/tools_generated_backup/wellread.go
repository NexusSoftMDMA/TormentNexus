package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func HandleGetBookSummary(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	isbn, _ :=getString(args, "isbn")
	if isbn == "" {
		return err("isbn is required")
}

	u := fmt.Sprintf("https://openlibrary.org/api/books?bibkeys=ISBN:%s&format=json&jscmd=data", url.QueryEscape(isbn))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request")
}

	client := http.DefaultClient
	client.Timeout = 10 * time.Second
	resp, e := client.Do(req)
	if e != nil {
		return err("failed to fetch book data")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response")
}

	key := "ISBN:" + isbn
	data, found := result[key]
	if !found {
		return err("book not found")
}

	details := data.(map[string]interface{})
	title, _ := details["title"].(string)
	subtitle, _ := details["subtitle"].(string)
	return ok(fmt.Sprintf("Title: %s\nSubtitle: %s", title, subtitle))
}