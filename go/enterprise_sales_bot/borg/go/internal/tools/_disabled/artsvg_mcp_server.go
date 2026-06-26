package tools

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func HandleSearchSVG(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query parameter is required")
}

	u := "https://artsvg.asifsofficial.com/api/search?q=" + url.QueryEscape(q)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to search: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("search request failed: " + string(body))
}

	var data []map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(string(body))
}

func HandleFetchSVG(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter is required")
}

	u := "https://artsvg.asifsofficial.com/api/svg?name=" + url.QueryEscape(name)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("fetch request failed: " + string(body))
}

	if !strings.HasPrefix(string(body), "<svg") {
		return err("response is not SVG")
}

	return ok(string(body))
}