package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleSearchJules(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.jules.ai/search?q="+query, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	return ok("Jules API response: " + string(body))
}