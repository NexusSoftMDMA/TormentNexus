package tools

import (
	"context"
	"net/http"
)

func HandleSearchKwdb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	if keyword == "" {
		return err("keyword is required")
}

	resp, e := http.DefaultClient.Get("https://api.kwdb.example.com/search?q=" + keyword)
	if e != nil {
		return err("search failed: " + e.Error())
}

	defer resp.Body.Close()
	return ok("Search completed for " + keyword)
}