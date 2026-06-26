package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleClickHouseQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	port, _ :=getString(args, "port")
	user, _ :=getString(args, "user")
	password, _ :=getString(args, "password")
	database, _ :=getString(args, "database")
	query, _ :=getString(args, "query")
	u := fmt.Sprintf("http://%s:%s?user=%s&password=%s&database=%s&query=%s",
		host, port, url.QueryEscape(user), url.QueryEscape(password), url.QueryEscape(database), url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("query failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("HTTP error: " + resp.Status)
}

	return ok(string(body))
}

func HandleClickHouseExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	port, _ :=getString(args, "port")
	user, _ :=getString(args, "user")
	password, _ :=getString(args, "password")
	database, _ :=getString(args, "database")
	query, _ :=getString(args, "query")
	u := fmt.Sprintf("http://%s:%s?user=%s&password=%s&database=%s", host, port, url.QueryEscape(user), url.QueryEscape(password), url.QueryEscape(database))
	resp, e := http.DefaultClient.PostForm(u, url.Values{"query": {query}})
	if e != nil {
		return err("execute failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("HTTP error: " + resp.Status)
}

	return ok(string(body))
}