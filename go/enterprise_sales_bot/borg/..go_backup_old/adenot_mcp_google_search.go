package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleGoogleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	num, _ :=getInt(args, "num")
	if num == 0 {
		num = 10
	}
	cx, _ :=getString(args, "cx")
	if cx == "" {
		cx = os.Getenv("GOOGLE_CX")

	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")

	if apiKey == "" {
		return err("GOOGLE_API_KEY is not set")
}

	u, _ := url.Parse("https://www.googleapis.com/customsearch/v1")
	q := u.Query()
	q.Set("key", apiKey)
	q.Set("cx", cx)
	q.Set("q", query)
	q.Set("num", fmt.Sprintf("%d", num))
	u.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("reading body failed: " + e.Error())
}

	var result struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("unmarshal failed: " + e.Error())
}

	output := ""
	for _, item := range result.Items {
		output += fmt.Sprintf("Title: %s\nLink: %s\nSnippet: %s\n\n", item.Title, item.Link, item.Snippet)

	if output == "" {
		return ok("No results found.")
}

	return ok(output)
}
}
}
}