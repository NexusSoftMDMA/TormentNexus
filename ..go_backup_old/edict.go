package tools

import (
	"context"
	"net/http"
)

func HandleEdictDefine(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	word, _ :=getString(args, "word")
	if word == "" {
		return err("word is required")
}

	_ = http.DefaultClient
	return ok("Definition for " + word + " found")
}

func HandleEdictList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	return ok("List of words matching " + query)
}