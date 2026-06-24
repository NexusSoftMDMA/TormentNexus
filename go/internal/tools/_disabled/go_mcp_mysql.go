package tools

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("sql is required")
	}
	resp, e := http.DefaultClient.Get("http://localhost:8080/query?sql=" + url.QueryEscape(sql))
	if e != nil {
		return err("http error: "+e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: "+e.Error())
	}
	return ok(string(body))
}

func HandleExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("sql is required")
	}
	resp, e := http.DefaultClient.Post("http://localhost:8080/execute", "text/plain", nil)
	if e != nil {
		return err("http error: "+e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: "+e.Error())
	}
	return ok(string(body))
}