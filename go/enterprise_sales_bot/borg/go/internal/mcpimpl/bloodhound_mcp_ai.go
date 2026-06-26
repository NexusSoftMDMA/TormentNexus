package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleSearch_bloodhound_mcp_ai(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := "http://localhost:8080/api/v1/search?q=" + url.QueryEscape(query)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("search failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}

func HandleGetNode_bloodhound_mcp_ai(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("node id is required")
}

	u := "http://localhost:8080/api/v1/nodes/" + url.PathEscape(id)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("node fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}