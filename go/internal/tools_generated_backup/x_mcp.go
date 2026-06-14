package tools

import (
	"context"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello, " + name + "!")
}

func HandleHttpRequest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("status: " + resp.Status)
}