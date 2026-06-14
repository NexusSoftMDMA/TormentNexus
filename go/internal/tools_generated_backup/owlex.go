package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleGetData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := fmt.Sprintf("Hello, %s! This is Owlex.", name)
	return ok(msg)
}

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.owlex.example/search?q="+query, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("search results for: " + query)
}