package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleExecuteSQL(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	token, _ :=getString(args, "token")
	if query == "" || token == "" {
		return err("query and token are required")
}

	u := "https://api.tinybird.co/v0/sql?q=" + url.QueryEscape(query)
	req, e := http.NewRequest("GET", u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("Tinybird API error: " + string(body))
}

	return success(string(body))
}