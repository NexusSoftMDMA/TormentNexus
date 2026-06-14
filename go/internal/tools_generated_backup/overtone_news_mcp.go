package tools

import (
	"context"
	"encoding/json"
)

func HandleGetHeadlines(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ :=getString(args, "source")
	data := map[string]interface{}{
		"headlines": []string{"Headline 1", "Headline 2"},
		"source":    source,
	}
	respBytes, e := json.Marshal(data)
	if e != nil {
		return err("failed to marshal response")
	}
	return ok(string(respBytes))
}

func HandleGetArticle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	article := map[string]interface{}{
		"id":      id,
		"title":   "Sample Article",
		"content": "This is a sample article content.",
	}
	respBytes, e := json.Marshal(article)
	if e != nil {
		return err("failed to marshal article")
	}
	return success(string(respBytes))
}