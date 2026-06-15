package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleScanGitHub(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://api.github.com/search/repositories?q=%s", q)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("GitHub request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse error: " + e.Error())
}

	items, found := data["items"].([]interface{})
	if !found {
		return ok("no results")
}

	return ok(fmt.Sprintf("Found %d repositories", len(items)))
}

func HandleScanPyPI(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", q)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("PyPI request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse error: " + e.Error())
}

	info, found := data["info"].(map[string]interface{})
	if !found {
		return ok("package not found")
}

	name, _ := info["name"].(string)
	return ok(fmt.Sprintf("PyPI package: %s", name))
}