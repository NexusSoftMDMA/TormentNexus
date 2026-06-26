package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleLearnSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://learn.microsoft.com/api/search?query=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

func HandleLearnGetModule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := fmt.Sprintf("https://learn.microsoft.com/api/modules/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}