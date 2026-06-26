package mcpimpl

import (
	"context"
	"fmt"
	"net/http"
)

func HandleEdgeeSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.edgee.ai/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("api returned status %d", resp.StatusCode))
}

	return ok("Search results retrieved successfully from Edgee")
}

func HandleEdgeeAnalyze(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	if text == "" {
		return err("text is required")
}

	return success(fmt.Sprintf("Analyzed text: %s", text))
}// touch 1781132132
