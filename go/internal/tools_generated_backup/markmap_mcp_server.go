package tools

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HandleGenerateMarkdownMindmap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	md, _ :=getString(args, "markdown")
	if md == "" {
		return err("markdown is required")
}

	u := "https://markmap.vercel.app/api?md=" + url.QueryEscape(md)
	resp, e := http.Get(u)
	if e != nil {
		return err("API call failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	return ok(string(body))
}