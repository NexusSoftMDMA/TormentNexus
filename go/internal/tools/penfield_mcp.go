//go:build ignore
// +build ignore

package tools

import (
	"context"
	"net/http"
	"net/url"
)

func HandleHello(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello, " + name + "!")
}

func HandleFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	u, _ :=getString(args, "url")
	if u == "" {
		return err("missing url")
}

	parsed, e := url.Parse(u)
	if e != nil {
		return err("invalid url: " + e.Error())
}

	resp, e := http.DefaultClient.Get(parsed.String())
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	return ok("fetched " + u + " with status " + resp.Status)
}
