package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	port, _ :=getString(args, "port")
	database, _ :=getString(args, "database")
	if host == "" || port == "" {
		return err("host and port are required")
}

	u := fmt.Sprintf("http://%s:%s/tables?database=%s", host, port, url.QueryEscape(database))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}

func HandleExecuteQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	port, _ :=getString(args, "port")
	query, _ :=getString(args, "query")
	if host == "" || port == "" || query == "" {
		return err("host, port, and query are required")
}

	u := fmt.Sprintf("http://%s:%s/query?q=%s", host, port, url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON response: " + e.Error())
}

	return ok(string(body))
}