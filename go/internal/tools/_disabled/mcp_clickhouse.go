package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleClickhouseQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("sql parameter is required")
}

	base := "http://localhost:8123"
	u, e := url.Parse(base)
	if e != nil {
		return err("invalid base URL")
}

	q := u.Query()
	q.Set("query", sql)
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}