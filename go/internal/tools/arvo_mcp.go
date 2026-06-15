//go:build ignore
// +build ignore

package tools

import (
	"context"
	"net/http"
)

func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	return ok("Hello, " + name + "!")
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	resp, e := http.DefaultClient.Get("https://echo.example.com?msg=" + msg)
	if e != nil {
		return err("HTTP error: " + e.Error())
}

	defer resp.Body.Close()
	return success("Echoed: " + msg)
}
