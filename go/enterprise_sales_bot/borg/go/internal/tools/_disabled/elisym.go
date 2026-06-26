package tools

import (
	"context"
	"net/http"
)

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://httpbin.org/get")
	if e != nil {
		return err("request failed: " + e.Error())
}

	resp.Body.Close()
	return ok("status: " + resp.Status)
}

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	return success("Hello, " + name + "!")
}