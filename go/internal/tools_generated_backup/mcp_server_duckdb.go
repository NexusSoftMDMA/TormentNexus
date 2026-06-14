package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "sql")
	if query == "" {
		return err("sql is required")
}

	u := "http://localhost:8000/query?sql=" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("query request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	return ok(string(body))
}

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:8000/tables")
	if e != nil {
		return err("list tables request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	return ok(string(body))
}