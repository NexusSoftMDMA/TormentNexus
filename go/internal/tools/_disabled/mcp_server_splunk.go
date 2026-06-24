package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleSearchSplunk(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	token, _ :=getString(args, "token")
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://localhost:8089"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/services/search/jobs", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer " + token)
	q := req.URL.Query()
	q.Add("search", query)
	req.URL.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}