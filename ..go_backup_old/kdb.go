package tools

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleExecuteQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	base, _ :=getString(args, "server_url")
	if base == "" {
		base = "http://localhost:5000"
	}
	u, e := url.JoinPath(base, "/query")
	if e != nil {
		return err("invalid server URL: " + e.Error())
}

	u += "?q=" + url.QueryEscape(query)
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("server returned " + resp.Status + ": " + string(body))
}

	return ok(string(body))
}