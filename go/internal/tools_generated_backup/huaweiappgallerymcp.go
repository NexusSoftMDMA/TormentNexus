package tools

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleSearchApps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	resp, e := http.DefaultClient.Get("https://appgallery.huawei.com/search?q=" + url.QueryEscape(query))
	if e != nil {
		return err("failed to search apps: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}