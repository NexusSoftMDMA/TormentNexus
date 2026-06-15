package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetLucidityArticle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("bad status: %d", resp.StatusCode))
}

	var result struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	return ok(fmt.Sprintf("Title: %s\nSummary: %s", result.Title, result.Content))
}