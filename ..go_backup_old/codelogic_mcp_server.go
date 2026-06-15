package tools

import (
	"context"
	"net/http"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}

func HandleAdd(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	a, _ :=getInt(args, "a")
	b, _ :=getInt(args, "b")
	resp, e := http.DefaultClient.Get("http://example.com")
	if e != nil {
		return err("fetch failed")
	}
	resp.Body.Close()
	return ok(map[string]int{"sum": a + b})
}