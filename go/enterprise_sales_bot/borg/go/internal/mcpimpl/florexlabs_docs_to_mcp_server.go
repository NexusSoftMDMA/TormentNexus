package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func SearchDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	base := "https://docs.florexlabs.com/api/search"
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("failed to search docs: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return ok(string(body))
}

	return success(fmt.Sprintf("Found %v results", result["count"]))
}

func GetDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id parameter is required")
}

	base := "https://docs.florexlabs.com/api/docs/" + id
	resp, e := http.DefaultClient.Get(base)
	if e != nil {
		return err("failed to get doc: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}